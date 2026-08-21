package novaredis

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/luaxlou/nova/internal/registry"
	"github.com/luaxlou/nova/starter/config/novaconfig"
	"github.com/redis/go-redis/v9"
)

type Instance interface {
	Client() (*redis.Client, error)
	Reload() error
	Close() error
}

type redisInstance struct {
	handle *registry.Instance[*redis.Client]
}

var (
	initialized bool
	initMu      sync.Mutex
	reg         = registry.New[*redis.Client]()
)

const singletonName = "single"

func initFromConfig() error {
	initMu.Lock()
	defer initMu.Unlock()
	if initialized {
		return nil
	}

	root, ok := asStringMap(novaconfig.Get("redis"))
	if !ok {
		return fmt.Errorf("redis config not found")
	}

	definitions, selectedName := buildDefinitions(root)
	if len(definitions) == 0 {
		return fmt.Errorf("redis config missing addr or instances")
	}

	reg.Configure(selectedName, definitions)
	initialized = true
	log.Printf("Redis Starter initialized, selected=%s", selectedName)
	return nil
}

func Get() *redisInstance {
	_ = ensureInit()
	return &redisInstance{handle: reg.Get()}
}

func Named(name string) *redisInstance {
	_ = ensureInit()
	return &redisInstance{handle: reg.Named(name)}
}

func Client() (*redis.Client, error) {
	_ = ensureInit()
	if reg.SelectedName() == "" && len(reg.Definitions()) > 1 {
		return nil, fmt.Errorf("redis instance name is required when multiple instances are configured")
	}
	return Named("").Client()
}

func (h *redisInstance) Client() (*redis.Client, error) {
	return h.handle.Get()
}

func (h *redisInstance) Reload() error {
	return h.handle.Reload()
}

func (h *redisInstance) Close() error {
	return h.handle.Close()
}

func Reload() {
	_ = ensureInit()
	_ = Get().Reload()
}

func Close() error {
	_ = ensureInit()
	if reg.SelectedName() == "" && len(reg.Definitions()) > 1 {
		return fmt.Errorf("redis instance name is required when multiple instances are configured")
	}
	return Get().Close()
}

func CloseAll() error {
	_ = ensureInit()
	return reg.CloseAll()
}

func buildDefinitions(root map[string]any) (map[string]registry.Builder[*redis.Client], string) {
	instances := map[string]redisConfig{}
	definitions := map[string]registry.Builder[*redis.Client]{}

	for name, raw := range root {
		if isReservedConfigKey(name) {
			continue
		}
		cfgMap, ok := asStringMap(raw)
		if !ok {
			continue
		}
		cfg := parseRedisConfig(cfgMap)
		if cfg.Addr == "" {
			continue
		}
		instances[name] = cfg
	}

	if len(instances) == 0 {
		if addr := asString(root["addr"]); addr != "" {
			instances[singletonName] = parseRedisConfig(root)
		}
	}

	if len(instances) == 0 {
		return nil, ""
	}

	for name, cfg := range instances {
		cfgCopy := cfg
		definitions[name] = func(_ string) (*redis.Client, error) {
			return newRedisConnection(cfgCopy)
		}
	}

	return definitions, chooseSingleName(instances)
}

func isReservedConfigKey(key string) bool {
	switch key {
	case "addr", "username", "password", "db", "pool_size", "min_idle_conns", "max_retries", "dial_timeout", "read_timeout", "write_timeout":
		return true
	default:
		return false
	}
}

func chooseSingleName(instances map[string]redisConfig) string {
	if len(instances) != 1 {
		return ""
	}
	for name := range instances {
		return name
	}
	return ""
}

func newRedisConnection(cfg redisConfig) (*redis.Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("redis addr is empty")
	}

	options := &redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.PoolSize > 0 {
		options.PoolSize = cfg.PoolSize
	}
	if cfg.MinIdleConns > 0 {
		options.MinIdleConns = cfg.MinIdleConns
	}
	if cfg.MaxRetries > 0 {
		options.MaxRetries = cfg.MaxRetries
	}
	if cfg.DialTimeout > 0 {
		options.DialTimeout = time.Duration(cfg.DialTimeout) * time.Second
	}
	if cfg.ReadTimeout > 0 {
		options.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	}
	if cfg.WriteTimeout > 0 {
		options.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	}

	client := redis.NewClient(options)
	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}

func ensureInit() error {
	if initialized {
		return nil
	}
	return initFromConfig()
}

type redisConfig struct {
	Addr         string
	Username     string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  int
	ReadTimeout  int
	WriteTimeout int
}

func parseRedisConfig(raw map[string]any) redisConfig {
	return redisConfig{
		Addr:         asString(raw["addr"]),
		Username:     asString(raw["username"]),
		Password:     asString(raw["password"]),
		DB:           asInt(raw["db"]),
		PoolSize:     asInt(raw["pool_size"]),
		MinIdleConns: asInt(raw["min_idle_conns"]),
		MaxRetries:   asInt(raw["max_retries"]),
		DialTimeout:  asInt(raw["dial_timeout"]),
		ReadTimeout:  asInt(raw["read_timeout"]),
		WriteTimeout: asInt(raw["write_timeout"]),
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
