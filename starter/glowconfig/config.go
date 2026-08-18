package glowconfig

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	globalConfig Config
	initialized  bool
	mu           sync.RWMutex
	configPath   = "config.yaml"
)

// Config represents the generic configuration map.
type Config map[string]any

// Get returns a value from the global config using dot notation for nested keys.
func Get(key string) any {
	config, err := getGlobalConfig()
	if err != nil {
		return nil
	}
	return config.getNested(key)
}

// GetString returns a string value or empty string.
func GetString(key string) string {
	val := Get(key)
	if v, ok := val.(string); ok {
		return v
	}
	return ""
}

// GetInt returns an int value or 0.
func GetInt(key string) int {
	val := Get(key)
	if v, ok := val.(int64); ok {
		return int(v)
	}
	if v, ok := val.(uint64); ok {
		return int(v)
	}
	if v, ok := val.(float64); ok {
		return int(v)
	}
	if v, ok := val.(int); ok {
		return v
	}
	return 0
}

// GetBool returns a bool value or false.
func GetBool(key string) bool {
	val := Get(key)
	if v, ok := val.(bool); ok {
		return v
	}
	return false
}

// getGlobalConfig returns the global config singleton (lazy-loaded).
func getGlobalConfig() (Config, error) {
	mu.RLock()
	if initialized && globalConfig != nil {
		defer mu.RUnlock()
		return globalConfig, nil
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	// Double-check after acquiring write lock
	if initialized && globalConfig != nil {
		return globalConfig, nil
	}

	log.Printf("Lazy loading config from %s...", configPath)
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	globalConfig = config
	initialized = true
	log.Println("Config loaded successfully.")

	return globalConfig, nil
}

// Reload forces a reload of the configuration file.
func Reload() error {
	mu.Lock()
	defer mu.Unlock()

	log.Printf("Reloading config from %s...", configPath)
	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	globalConfig = config
	initialized = true
	log.Println("Config reloaded successfully.")

	return nil
}

// SetConfigPath sets a custom config file path (must be called before first Get).
func SetConfigPath(path string) {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		log.Println("Warning: Config already initialized, path change will not take effect until Reload()")
	}

	configPath = path
}

// loadConfig reads and parses the config file.
func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config %s: %w", path, err)
	}

	var config Config
	if parsed, ok := normalizeConfig(raw); ok {
		config = parsed
	} else {
		return nil, fmt.Errorf("invalid config format in %s", path)
	}

	return config, nil
}

func (c Config) getNested(key string) any {
	keys := splitKey(key)
	var current any = c

	for _, k := range keys {
		m, ok := current.(map[string]any)
		if !ok {
			// Try Config type alias
			if cm, ok := current.(Config); ok {
				m = map[string]any(cm)
			} else {
				return nil
			}
		}
		val, ok := m[k]
		if !ok {
			return nil
		}
		current = val
	}
	return current
}

func splitKey(key string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			parts = append(parts, key[start:i])
			start = i + 1
		}
	}
	parts = append(parts, key[start:])
	return parts
}

func normalizeConfig(v any) (Config, bool) {
	switch typed := v.(type) {
	case Config:
		return typed, true
	case map[string]any:
		return normalizeStringMap(typed), true
	case map[any]any:
		normalized := make(map[string]any, len(typed))
		for key, value := range typed {
			keyStr, ok := key.(string)
			if !ok {
				return nil, false
			}

			normalized[keyStr] = normalizeValue(value)
		}
		return Config(normalized), true
	default:
		return nil, false
	}
}

func normalizeStringMap(in map[string]any) Config {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = normalizeValue(value)
	}
	return Config(out)
}

func normalizeValue(v any) any {
	switch value := v.(type) {
	case map[any]any:
		normalized := make(map[string]any, len(value))
		for k, nested := range value {
			keyStr, ok := k.(string)
			if !ok {
				continue
			}
			normalized[keyStr] = normalizeValue(nested)
		}
		return normalizeStringMap(normalized)
	case map[string]any:
		return normalizeStringMap(value)
	default:
		return v
	}
}
