package novagorm

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/luaxlou/nova/starter/internal/registry"
	"github.com/luaxlou/nova/starter/novaconfig"
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

var (
	initialized bool
	initMu      sync.Mutex
	reg         = registry.New[*gorm.DB]()

	manualDefinitions = map[string]Builder{}
)

func Init() error {
	initMu.Lock()
	defer initMu.Unlock()
	if initialized {
		return nil
	}

	definitions := make(map[string]Builder, len(manualDefinitions))
	for name, builder := range manualDefinitions {
		definitions[name] = builder
	}

	defaultName := ""
	if root, ok := asStringMap(novaconfig.Get("gorm")); ok {
		configDefinitions, configDefault := buildDefinitions(root)
		for name, builder := range configDefinitions {
			definitions[name] = builder
		}
		defaultName = configDefault
	}

	if defaultName == "" {
		defaultName = chooseDefaultName(definitions)
	}

	reg.Init(defaultName, definitions)
	initialized = true
	log.Printf("GORM Starter initialized, default=%s", defaultName)
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

	if rawInstances, ok := asStringMap(root["instances"]); ok {
		for name := range rawInstances {
			if cfgMap, ok := asStringMap(rawInstances[name]); ok {
				instances[name] = parseGormConfig(cfgMap)
			}
		}
	}

	if len(instances) == 0 {
		cfg := parseGormConfig(root)
		if cfg.Driver != "" || cfg.DSN != "" {
			instances["default"] = cfg
		}
	}

	if len(instances) == 0 {
		return nil, ""
	}

	defaultName := asString(root["default"])
	if defaultName == "" {
		defaultName = chooseConfigDefaultName(instances)
	}

	for name, cfg := range instances {
		cfgCopy := cfg
		definitions[name] = func(_ string) (*gorm.DB, error) {
			return newConfiguredConnection(cfgCopy)
		}
	}

	return definitions, defaultName
}

func newConfiguredConnection(cfg gormConfig) (*gorm.DB, error) {
	driver := cfg.Driver
	if driver == "" && cfg.DSN != "" {
		driver = "mysql"
	}

	switch driver {
	case "mysql":
		return newMySQLConnection(cfg)
	default:
		return nil, fmt.Errorf("unsupported gorm driver %q", driver)
	}
}

func newMySQLConnection(cfg gormConfig) (*gorm.DB, error) {
	if cfg.DSN != "" {
		return gorm.Open(gormmysql.New(gormmysql.Config{
			DSN:                       cfg.DSN,
			SkipInitializeWithVersion: cfg.SkipInitializeWithVersion,
		}), &gorm.Config{})
	}
	return nil, fmt.Errorf("gorm mysql config missing dsn")
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
	return Init()
}

type gormConfig struct {
	Driver                    string
	DSN                       string
	SkipInitializeWithVersion bool
}

func parseGormConfig(raw map[string]any) gormConfig {
	return gormConfig{
		Driver:                    asString(raw["driver"]),
		DSN:                       asString(raw["dsn"]),
		SkipInitializeWithVersion: asBool(raw["skip_initialize_with_version"]),
	}
}

func chooseDefaultName(definitions map[string]Builder) string {
	if definitions["default"] != nil {
		return "default"
	}
	for name := range definitions {
		return name
	}
	return ""
}

func chooseConfigDefaultName(instances map[string]gormConfig) string {
	if _, ok := instances["default"]; ok {
		return "default"
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
