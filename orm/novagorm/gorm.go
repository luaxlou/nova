package novagorm

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/luaxlou/nova/internal/registry"
	"github.com/luaxlou/nova/starter/config/novaconfig"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Builder = registry.Builder[*gorm.DB]

type Instance interface {
	DB() (*gorm.DB, error)
	Reload() error
	Close() error
}

type gormInstance struct {
	handle *registry.Instance[*gorm.DB]
}

const singletonName = "single"

var (
	initialized bool
	initMu      sync.Mutex
	reg         = registry.New[*gorm.DB]()

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

	reg.Configure(selectedName, definitions)
	selectedInstanceName = selectedName
	initialized = true
	log.Printf("GORM tool initialized, selected=%s", selectedName)
	return nil
}

func Register(name string, builder Builder) {
	initMu.Lock()
	defer initMu.Unlock()

	manualDefinitions[name] = builder
	reg.Register(name, builder)
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
	return h.handle.Get()
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

	for name, cfg := range instances {
		cfgCopy := cfg
		definitions[name] = func(_ string) (*gorm.DB, error) {
			return newConfiguredConnection(cfgCopy)
		}
	}

	return definitions, selectedName
}

func isReservedConfigKey(key string) bool {
	switch key {
	case "default", "driver", "mysql":
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
	default:
		return nil, fmt.Errorf("unsupported gorm driver %q", driver)
	}
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
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get mysql sql db from gorm: %w", err)
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
	return db, nil
}

func ensureInit() error {
	if initialized {
		return nil
	}
	return initFromConfig()
}

type gormConfig struct {
	Driver string
	MySQL  mysqlConfig
}

type mysqlConfig struct {
	DSN                       string
	SkipInitializeWithVersion bool
	MaxOpen                   int
	MaxIdle                   int
	ConnMaxLifetime           int
	ConnMaxIdleTime           int
}

func parseGormConfig(raw map[string]any) gormConfig {
	mysqlRaw, _ := asStringMap(raw["mysql"])
	return gormConfig{
		Driver: asString(raw["driver"]),
		MySQL:  parseMySQLConfig(mysqlRaw),
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
