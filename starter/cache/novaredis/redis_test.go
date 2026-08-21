package novaredis

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/luaxlou/nova/starter/config/novaconfig"
)

func TestRedisConfigAcceptsNestedNovaConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	config := []byte(`
redis:
  main:
    addr: 127.0.0.1:6379
    db: 2
    pool_size: 16
`)
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	novaconfig.SetConfigPath(configPath)
	if err := novaconfig.Reload(); err != nil {
		t.Fatalf("reload config: %v", err)
	}

	root, ok := asStringMap(novaconfig.Get("redis"))
	if !ok {
		t.Fatal("nested redis config must be accepted")
	}
	definitions, selected := buildDefinitions(root)
	if len(definitions) != 1 {
		t.Fatalf("expected one redis definition, got %d", len(definitions))
	}
	if selected != "main" {
		t.Fatalf("expected main to be selected, got %q", selected)
	}
}
