package novagorm

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/luaxlou/nova/internal/registry"
	"github.com/luaxlou/nova/starter/config/novaconfig"
	gormmysql "gorm.io/driver/mysql"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type Builder = registry.Builder[*gorm.DB]

var ErrORMMagic = errors.New("novagorm: ORM magic blocked")

type gormResource struct {
	db *gorm.DB
}

func (r *gormResource) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("get sql db while closing gorm: %w", err)
	}
	return sqlDB.Close()
}

type Instance interface {
	DB() (*gorm.DB, error)
	Reload() error
	Close() error
}

type gormInstance struct {
	handle *registry.Instance[*gormResource]
}

const singletonName = "single"

var (
	initialized bool
	initMu      sync.Mutex
	reg         = registry.New[*gormResource]()

	manualDefinitions    = map[string]Builder{}
	selectedInstanceName = ""
)

func initFromConfig() error {
	initMu.Lock()
	defer initMu.Unlock()
	if initialized {
		return nil
	}

	definitions := make(map[string]Builder, len(manualDefinitions))
	for name, builder := range manualDefinitions {
		definitions[name] = builder
	}

	selectedName := ""
	if root, ok := asStringMap(novaconfig.Get("gorm")); ok {
		configDefinitions, configSelected := buildDefinitions(root)
		for name, builder := range configDefinitions {
			definitions[name] = builder
		}
		selectedName = configSelected
	}

	if selectedName == "" {
		selectedName = chooseSingleName(definitions)
	}

	reg.Configure(selectedName, wrapBuilders(definitions))
	selectedInstanceName = selectedName
	initialized = true
	log.Printf("GORM tool initialized, selected=%s", selectedName)
	return nil
}

func Register(name string, builder Builder) {
	initMu.Lock()
	defer initMu.Unlock()

	manualDefinitions[name] = builder
	reg.Register(name, wrapBuilder(builder))
}

func Get() *gormInstance {
	_ = ensureInit()
	return &gormInstance{handle: reg.Get()}
}

func Named(name string) *gormInstance {
	_ = ensureInit()
	return &gormInstance{handle: reg.Named(name)}
}

func DB() (*gorm.DB, error) {
	_ = ensureInit()
	if selectedInstanceName == "" && len(reg.Definitions()) > 1 {
		return nil, fmt.Errorf("gorm instance name is required when multiple instances are configured")
	}
	return Named("").DB()
}

func (h *gormInstance) DB() (*gorm.DB, error) {
	resource, err := h.handle.Get()
	if err != nil {
		return nil, err
	}
	return resource.db, nil
}

func (h *gormInstance) Reload() error {
	return h.handle.Reload()
}

func (h *gormInstance) Close() error {
	return h.handle.Close()
}

func Reload() {
	_ = ensureInit()
	_ = Get().Reload()
}

func Close() error {
	_ = ensureInit()
	return Get().Close()
}

func CloseAll() error {
	_ = ensureInit()
	return reg.CloseAll()
}

func wrapBuilders(definitions map[string]Builder) map[string]registry.Builder[*gormResource] {
	wrapped := make(map[string]registry.Builder[*gormResource], len(definitions))
	for name, builder := range definitions {
		wrapped[name] = wrapBuilder(builder)
	}
	return wrapped
}

func wrapBuilder(builder Builder) registry.Builder[*gormResource] {
	return func(name string) (*gormResource, error) {
		db, err := builder(name)
		if err != nil {
			return nil, err
		}
		installAutoMigrateGuard(db)
		return &gormResource{db: db}, nil
	}
}

type autoMigrateGuardDialector struct {
	gorm.Dialector
}

func (d autoMigrateGuardDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return &autoMigrateGuard{
		Migrator: d.Dialector.Migrator(db),
		db:       db,
	}
}

type autoMigrateGuard struct {
	gorm.Migrator
	db *gorm.DB
}

func (g *autoMigrateGuard) AutoMigrate(models ...any) error {
	for _, model := range models {
		if err := validateAutoMigrateModel(g.db, model); err != nil {
			return err
		}
	}

	return g.Migrator.AutoMigrate(models...)
}

// BuildIndexOptions preserves the extended migrator contract used by GORM
// while AutoMigrate is routed through the Model First guard.
func (g *autoMigrateGuard) BuildIndexOptions(options []schema.IndexOption, statement *gorm.Statement) []interface{} {
	builder, ok := g.Migrator.(migrator.BuildIndexOptionsInterface)
	if !ok {
		return nil
	}
	return builder.BuildIndexOptions(options, statement)
}

func validateAutoMigrateModel(db *gorm.DB, model any) error {
	statement := &gorm.Statement{DB: db}
	if err := statement.Parse(model); err != nil {
		return err
	}

	modelType := statement.Schema.ModelType
	modelName := statement.Schema.Name
	modelPointer := reflect.New(modelType).Interface()
	if _, ok := modelPointer.(schema.Tabler); !ok {
		return ormMagicError(modelName, "TableName() is required")
	}

	if embedsGormModel(modelType) {
		return ormMagicError(modelName, "embedded gorm.Model is not allowed")
	}

	deletedAtType := reflect.TypeOf(gorm.DeletedAt{})
	for _, field := range statement.Schema.Fields {
		if field.IndirectFieldType == deletedAtType {
			return ormMagicError(modelName, fmt.Sprintf("field %s uses gorm.DeletedAt", field.Name))
		}
		if field.AutoCreateTime != 0 {
			return ormMagicError(modelName, fmt.Sprintf("field %s enables autoCreateTime", field.Name))
		}
		if field.AutoUpdateTime != 0 {
			return ormMagicError(modelName, fmt.Sprintf("field %s enables autoUpdateTime", field.Name))
		}
		if field.DBName != "" && !field.IgnoreMigration && field.TagSettings["COLUMN"] == "" {
			return ormMagicError(modelName, fmt.Sprintf("field %s requires an explicit column tag", field.Name))
		}
	}

	if hook := modelLifecycleHook(statement.Schema); hook != "" {
		return ormMagicError(modelName, fmt.Sprintf("lifecycle hook %s is not allowed", hook))
	}

	for name := range statement.Schema.Relationships.Relations {
		return ormMagicError(modelName, fmt.Sprintf("association %s is not allowed", name))
	}

	return nil
}

func embedsGormModel(modelType reflect.Type) bool {
	gormModelType := reflect.TypeOf(gorm.Model{})
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		fieldType := field.Type
		for fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		if field.Anonymous && fieldType == gormModelType {
			return true
		}
	}
	return false
}

func modelLifecycleHook(modelSchema *schema.Schema) string {
	switch {
	case modelSchema.BeforeCreate:
		return "BeforeCreate"
	case modelSchema.AfterCreate:
		return "AfterCreate"
	case modelSchema.BeforeUpdate:
		return "BeforeUpdate"
	case modelSchema.AfterUpdate:
		return "AfterUpdate"
	case modelSchema.BeforeSave:
		return "BeforeSave"
	case modelSchema.AfterSave:
		return "AfterSave"
	case modelSchema.BeforeDelete:
		return "BeforeDelete"
	case modelSchema.AfterDelete:
		return "AfterDelete"
	case modelSchema.AfterFind:
		return "AfterFind"
	default:
		return ""
	}
}

func ormMagicError(modelName, reason string) error {
	return fmt.Errorf("%w for model %s: %s", ErrORMMagic, modelName, reason)
}

func installAutoMigrateGuard(db *gorm.DB) {
	if db == nil || db.Config == nil || db.Dialector == nil {
		return
	}
	if _, guarded := db.Dialector.(autoMigrateGuardDialector); guarded {
		return
	}
	db.Dialector = autoMigrateGuardDialector{Dialector: db.Dialector}
}

func buildDefinitions(root map[string]any) (map[string]Builder, string) {
	instances := map[string]gormConfig{}
	definitions := map[string]Builder{}

	for name, raw := range root {
		if isReservedConfigKey(name) {
			continue
		}
		cfgMap, ok := asStringMap(raw)
		if !ok {
			continue
		}
		cfg := parseGormConfig(cfgMap)
		if cfg.Driver == "" {
			continue
		}
		instances[name] = cfg
	}

	if len(instances) == 0 {
		cfg := parseGormConfig(root)
		if cfg.Driver != "" {
			instances[singletonName] = cfg
		}
	}

	if len(instances) == 0 {
		return nil, ""
	}

	selectedName := chooseSingleConfigName(instances)
	if configuredDefault := asString(root["default"]); configuredDefault != "" {
		if _, ok := instances[configuredDefault]; ok {
			selectedName = configuredDefault
		}
	}

	for name, cfg := range instances {
		cfgCopy := cfg
		definitions[name] = func(instanceName string) (*gorm.DB, error) {
			db, err := newConfiguredConnection(cfgCopy)
			if err != nil {
				return nil, fmt.Errorf("open gorm instance %q with driver %q: %w", instanceName, cfgCopy.Driver, err)
			}
			return db, nil
		}
	}

	return definitions, selectedName
}

func isReservedConfigKey(key string) bool {
	switch key {
	case "default", "driver", "mysql", "postgres":
		return true
	default:
		return false
	}
}

func newConfiguredConnection(cfg gormConfig) (*gorm.DB, error) {
	driver := cfg.Driver

	switch driver {
	case "mysql":
		return newMySQLConnection(cfg)
	case "postgres":
		return newPostgresConnection(cfg)
	default:
		return nil, fmt.Errorf("unsupported gorm driver %q", driver)
	}
}

func newPostgresConnection(cfg gormConfig) (*gorm.DB, error) {
	if cfg.Postgres.DSN == "" {
		return nil, fmt.Errorf("gorm postgres config missing postgres.dsn")
	}

	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{
		DSN:                  cfg.Postgres.DSN,
		PreferSimpleProtocol: cfg.Postgres.PreferSimpleProtocol,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres gorm connection: %w", err)
	}
	if err := applyPostgresPoolConfig(db, cfg.Postgres); err != nil {
		return nil, err
	}
	return db, nil
}

func newMySQLConnection(cfg gormConfig) (*gorm.DB, error) {
	if cfg.MySQL.DSN != "" {
		db, err := gorm.Open(gormmysql.New(gormmysql.Config{
			DSN:                       cfg.MySQL.DSN,
			SkipInitializeWithVersion: cfg.MySQL.SkipInitializeWithVersion,
		}), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		if err := applyMySQLPoolConfig(db, cfg.MySQL); err != nil {
			return nil, err
		}
		return db, nil
	}
	return nil, fmt.Errorf("gorm mysql config missing mysql.dsn")
}

func applyMySQLPoolConfig(db *gorm.DB, cfg mysqlConfig) error {
	return applyPoolConfig(db, "mysql", poolConfig{
		MaxOpen:         cfg.MaxOpen,
		MaxIdle:         cfg.MaxIdle,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	})
}

func applyPostgresPoolConfig(db *gorm.DB, cfg postgresConfig) error {
	return applyPoolConfig(db, "postgres", poolConfig{
		MaxOpen:         cfg.MaxOpen,
		MaxIdle:         cfg.MaxIdle,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	})
}

type poolConfig struct {
	MaxOpen         int
	MaxIdle         int
	ConnMaxLifetime int
	ConnMaxIdleTime int
}

func applyPoolConfig(db *gorm.DB, driver string, cfg poolConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get %s sql db from gorm: %w", driver, err)
	}
	if cfg.MaxOpen > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}
	return nil
}

func OpenMySQLFromSQLDB(sqlDB *sql.DB) (*gorm.DB, error) {
	if sqlDB == nil {
		return nil, fmt.Errorf("sql db is nil")
	}

	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm from sql db: %w", err)
	}
	installAutoMigrateGuard(db)
	return db, nil
}

func OpenPostgresFromSQLDB(sqlDB *sql.DB) (*gorm.DB, error) {
	if sqlDB == nil {
		return nil, fmt.Errorf("sql db is nil")
	}

	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres gorm from sql db: %w", err)
	}
	installAutoMigrateGuard(db)
	return db, nil
}

func ensureInit() error {
	if initialized {
		return nil
	}
	return initFromConfig()
}

type gormConfig struct {
	Driver   string
	MySQL    mysqlConfig
	Postgres postgresConfig
}

type mysqlConfig struct {
	DSN                       string
	SkipInitializeWithVersion bool
	MaxOpen                   int
	MaxIdle                   int
	ConnMaxLifetime           int
	ConnMaxIdleTime           int
}

type postgresConfig struct {
	DSN                  string
	PreferSimpleProtocol bool
	MaxOpen              int
	MaxIdle              int
	ConnMaxLifetime      int
	ConnMaxIdleTime      int
}

func parseGormConfig(raw map[string]any) gormConfig {
	mysqlRaw, _ := asStringMap(raw["mysql"])
	postgresRaw, _ := asStringMap(raw["postgres"])
	return gormConfig{
		Driver:   asString(raw["driver"]),
		MySQL:    parseMySQLConfig(mysqlRaw),
		Postgres: parsePostgresConfig(postgresRaw),
	}
}

func parseMySQLConfig(raw map[string]any) mysqlConfig {
	return mysqlConfig{
		DSN:                       asString(raw["dsn"]),
		SkipInitializeWithVersion: asBool(raw["skip_initialize_with_version"]),
		MaxOpen:                   firstInt(raw, "max_open", "max_open_conns"),
		MaxIdle:                   firstInt(raw, "max_idle", "max_idle_conns"),
		ConnMaxLifetime:           firstInt(raw, "conn_max_lifetime", "conn_max_lifetime_sec"),
		ConnMaxIdleTime:           asInt(raw["conn_max_idle_time"]),
	}
}

func parsePostgresConfig(raw map[string]any) postgresConfig {
	return postgresConfig{
		DSN:                  asString(raw["dsn"]),
		PreferSimpleProtocol: asBool(raw["prefer_simple_protocol"]),
		MaxOpen:              firstInt(raw, "max_open", "max_open_conns"),
		MaxIdle:              firstInt(raw, "max_idle", "max_idle_conns"),
		ConnMaxLifetime:      firstInt(raw, "conn_max_lifetime", "conn_max_lifetime_sec"),
		ConnMaxIdleTime:      asInt(raw["conn_max_idle_time"]),
	}
}

func firstInt(raw map[string]any, keys ...string) int {
	for _, key := range keys {
		if value, ok := raw[key]; ok {
			return asInt(value)
		}
	}
	return 0
}

func chooseSingleName(definitions map[string]Builder) string {
	if len(definitions) != 1 {
		return ""
	}
	for name := range definitions {
		return name
	}
	return ""
}

func chooseSingleConfigName(instances map[string]gormConfig) string {
	if len(instances) != 1 {
		return ""
	}
	for name := range instances {
		return name
	}
	return ""
}

func asString(value any) string {
	if value == nil {
		return ""
	}
	s, ok := value.(string)
	if !ok {
		return ""
	}
	return s
}

func asBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		parsed, err := strconv.ParseBool(typed)
		if err != nil {
			return false
		}
		return parsed
	default:
		return false
	}
}

func asInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(typed)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func asStringMap(value any) (map[string]any, bool) {
	if value == nil {
		return nil, false
	}

	if typed, ok := value.(map[string]any); ok {
		return typed, true
	}

	if raw, ok := value.(map[any]any); ok {
		converted := make(map[string]any, len(raw))
		for k, v := range raw {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			converted[ks] = v
		}
		return converted, true
	}

	if raw, ok := value.(novaconfig.Config); ok {
		return map[string]any(raw), true
	}

	return nil, false
}
