package novamysql

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/luaxlou/nova/starter/internal/registry"
	"github.com/luaxlou/nova/starter/novaconfig"
)

type Instance interface {
	DB() (*sql.DB, error)
	Reload() error
	Close() error
}

type mysqlInstance struct {
	handle *registry.Instance[*sql.DB]
}

var (
	initialized bool
	initMu      sync.Mutex
	reg         = registry.New[*sql.DB]()
)

func Init() error {
	initMu.Lock()
	defer initMu.Unlock()
	if initialized {
		return nil
	}

	root, ok := asStringMap(novaconfig.Get("mysql"))
	if !ok {
		return fmt.Errorf("mysql config not found. call novamysql.Init() after config file load")
	}

	definitions, defaultName := buildDefinitions(root)
	if len(definitions) == 0 {
		return fmt.Errorf("mysql config missing dsn or instances")
	}

	reg.Init(defaultName, definitions)
	initialized = true
	log.Printf("MySQL Starter initialized, default=%s", defaultName)
	return nil
}

func Get() *mysqlInstance {
	_ = ensureInit()
	return &mysqlInstance{handle: reg.Get()}
}

func Named(name string) *mysqlInstance {
	_ = ensureInit()
	return &mysqlInstance{handle: reg.Named(name)}
}

func DB() (*sql.DB, error) {
	return Named("").DB()
}

func (h *mysqlInstance) DB() (*sql.DB, error) {
	return h.handle.Get()
}

func (h *mysqlInstance) Reload() error {
	return h.handle.Reload()
}

func (h *mysqlInstance) Close() error {
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

func buildDefinitions(root map[string]any) (map[string]registry.Builder[*sql.DB], string) {
	instances := map[string]mysqlConfig{}
	definitions := map[string]registry.Builder[*sql.DB]{}

	if rawInstances, ok := asStringMap(root["instances"]); ok {
		for name := range rawInstances {
			if cfgMap, ok := asStringMap(rawInstances[name]); ok {
				instances[name] = parseMySQLConfig(cfgMap)
			}
		}
	}

	if len(instances) == 0 {
		if dsn := asString(root["dsn"]); dsn != "" {
			instances["default"] = parseMySQLConfig(root)
		}
	}

	if len(instances) == 0 {
		return nil, ""
	}

	defaultName := asString(root["default"])
	if defaultName == "" {
		for name := range instances {
			defaultName = name
			break
		}
	}
	if defaultName == "" {
		defaultName = "default"
	}

	for name, cfg := range instances {
		cfgCopy := cfg
		definitions[name] = func(_ string) (*sql.DB, error) {
			return newMySQLConnection(cfgCopy)
		}
	}

	return definitions, defaultName
}

func newMySQLConnection(cfg mysqlConfig) (*sql.DB, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("mysql dsn is empty")
	}

	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql: %w", err)
	}

	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}
	if cfg.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}

	return db, nil
}

func ensureInit() error {
	if initialized {
		return nil
	}
	return Init()
}

type mysqlConfig struct {
	DSN             string
	MaxOpen         int
	MaxIdle         int
	ConnMaxLifetime int
	ConnMaxIdleTime int
}

func parseMySQLConfig(raw map[string]any) mysqlConfig {
	return mysqlConfig{
		DSN:             asString(raw["dsn"]),
		MaxOpen:         asInt(raw["max_open"]),
		MaxIdle:         asInt(raw["max_idle"]),
		ConnMaxLifetime: asInt(raw["conn_max_lifetime"]),
		ConnMaxIdleTime: asInt(raw["conn_max_idle_time"]),
	}
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

	return nil, false
}
