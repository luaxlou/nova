// Package novaoss provides lazy, named OSS bucket access through Alibaba Cloud's SDK.
package novaoss

import (
	"fmt"
	"log"
	"strings"
	"sync"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/luaxlou/nova/internal/registry"
	"github.com/luaxlou/nova/starter/config/novaconfig"
)

// Instance is a named OSS bucket handle.
type Instance interface {
	Bucket() (*aliyunoss.Bucket, error)
	Reload() error
	Close() error
}

type ossInstance struct {
	handle  *registry.Instance[*aliyunoss.Bucket]
	initErr error
}

var (
	initialized bool
	initMu      sync.Mutex
	reg         = registry.New[*aliyunoss.Bucket]()
)

const singletonName = "single"

func initFromConfig() error {
	initMu.Lock()
	defer initMu.Unlock()
	if initialized {
		return nil
	}

	root, ok := asStringMap(novaconfig.Get("aliyun.oss"))
	if !ok {
		return fmt.Errorf("aliyun.oss config not found")
	}

	definitions, selectedName, err := buildDefinitions(root)
	if err != nil {
		return err
	}
	if len(definitions) == 0 {
		return fmt.Errorf("oss config missing instance definitions")
	}

	reg.Configure(selectedName, definitions)
	initialized = true
	log.Printf("OSS Starter initialized, selected=%s", selectedName)
	return nil
}

// Get returns the default OSS bucket handle.
func Get() *ossInstance {
	return Named("")
}

// Named returns the OSS bucket handle for name.
func Named(name string) *ossInstance {
	initErr := ensureInit()
	return &ossInstance{
		handle:  reg.Named(name),
		initErr: initErr,
	}
}

// Bucket returns the default configured OSS bucket.
func Bucket() (*aliyunoss.Bucket, error) {
	if err := ensureInit(); err != nil {
		return nil, err
	}
	if reg.SelectedName() == "" && len(reg.Definitions()) > 1 {
		return nil, fmt.Errorf("oss instance name is required when multiple instances are configured")
	}
	return Get().Bucket()
}

// Bucket returns the configured OSS bucket for this handle.
func (h *ossInstance) Bucket() (*aliyunoss.Bucket, error) {
	if h.initErr != nil {
		return nil, h.initErr
	}
	return h.handle.Get()
}

// Reload rebuilds the default bucket handle.
func Reload() error {
	if err := ensureInit(); err != nil {
		return err
	}
	if reg.SelectedName() == "" && len(reg.Definitions()) > 1 {
		return fmt.Errorf("oss instance name is required when multiple instances are configured")
	}
	return Get().Reload()
}

// Reload rebuilds this bucket handle.
func (h *ossInstance) Reload() error {
	if h.initErr != nil {
		return h.initErr
	}
	return h.handle.Reload()
}

// Close releases the default cached bucket handle.
func Close() error {
	if err := ensureInit(); err != nil {
		return err
	}
	if reg.SelectedName() == "" && len(reg.Definitions()) > 1 {
		return fmt.Errorf("oss instance name is required when multiple instances are configured")
	}
	return Get().Close()
}

// Close releases this cached bucket handle.
func (h *ossInstance) Close() error {
	if h.initErr != nil {
		return h.initErr
	}
	return h.handle.Close()
}

// CloseAll releases all cached bucket handles.
func CloseAll() error {
	if err := ensureInit(); err != nil {
		return err
	}
	return reg.CloseAll()
}

func ensureInit() error {
	if initialized {
		return nil
	}
	return initFromConfig()
}

type ossConfig struct {
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	AccessKeySecret string
	SecurityToken   string
}

func buildDefinitions(root map[string]any) (map[string]registry.Builder[*aliyunoss.Bucket], string, error) {
	instances := map[string]ossConfig{}
	for name, raw := range root {
		if isReservedConfigKey(name) {
			continue
		}
		cfgMap, ok := asStringMap(raw)
		if !ok {
			continue
		}
		if strings.TrimSpace(name) == "" {
			return nil, "", fmt.Errorf("oss config contains an empty instance name")
		}
		cfg := parseOSSConfig(cfgMap)
		if !cfg.hasAnyRequiredField() {
			continue
		}
		instances[name] = cfg
	}

	if len(instances) == 0 {
		instances[singletonName] = parseOSSConfig(root)
	}

	definitions := make(map[string]registry.Builder[*aliyunoss.Bucket], len(instances))
	for name, cfg := range instances {
		nameCopy := name
		cfgCopy := cfg
		definitions[name] = func(_ string) (*aliyunoss.Bucket, error) {
			return newBucket(nameCopy, cfgCopy)
		}
	}

	return definitions, chooseSingleName(instances), nil
}

func (cfg ossConfig) hasAnyRequiredField() bool {
	return cfg.Endpoint != "" || cfg.Bucket != "" || cfg.AccessKeyID != "" || cfg.AccessKeySecret != ""
}

func isReservedConfigKey(key string) bool {
	switch key {
	case "endpoint", "bucket", "access_key_id", "access_key_secret", "security_token":
		return true
	default:
		return false
	}
}

func chooseSingleName(instances map[string]ossConfig) string {
	if len(instances) != 1 {
		return ""
	}
	for name := range instances {
		return name
	}
	return ""
}

func newBucket(name string, cfg ossConfig) (*aliyunoss.Bucket, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, fmt.Errorf("oss instance %q config missing endpoint", name)
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("oss instance %q config missing bucket", name)
	}
	if strings.TrimSpace(cfg.AccessKeyID) == "" {
		return nil, fmt.Errorf("oss instance %q config missing access_key_id", name)
	}
	if strings.TrimSpace(cfg.AccessKeySecret) == "" {
		return nil, fmt.Errorf("oss instance %q config missing access_key_secret", name)
	}

	options := make([]aliyunoss.ClientOption, 0, 1)
	if cfg.SecurityToken != "" {
		options = append(options, aliyunoss.SecurityToken(cfg.SecurityToken))
	}
	client, err := aliyunoss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret, options...)
	if err != nil {
		return nil, fmt.Errorf("create oss client for instance %q: %w", name, err)
	}

	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("select oss bucket %q for instance %q: %w", cfg.Bucket, name, err)
	}
	return bucket, nil
}

func parseOSSConfig(raw map[string]any) ossConfig {
	return ossConfig{
		Endpoint:        asString(raw["endpoint"]),
		Bucket:          asString(raw["bucket"]),
		AccessKeyID:     asString(raw["access_key_id"]),
		AccessKeySecret: asString(raw["access_key_secret"]),
		SecurityToken:   asString(raw["security_token"]),
	}
}

func asString(value any) string {
	if value == nil {
		return ""
	}
	stringValue, ok := value.(string)
	if !ok {
		return ""
	}
	return stringValue
}

func asStringMap(value any) (map[string]any, bool) {
	if value == nil {
		return nil, false
	}
	if value, ok := value.(map[string]any); ok {
		return value, true
	}
	if value, ok := value.(map[any]any); ok {
		converted := make(map[string]any, len(value))
		for key, item := range value {
			key, ok := key.(string)
			if !ok {
				continue
			}
			converted[key] = item
		}
		return converted, true
	}
	if value, ok := value.(novaconfig.Config); ok {
		return map[string]any(value), true
	}
	return nil, false
}
