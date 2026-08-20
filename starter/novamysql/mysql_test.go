package novamysql

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/luaxlou/nova/starter/internal/registry"
	"github.com/luaxlou/nova/starter/novaconfig"
)

func TestInitAcceptsNestedNovaConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	requireWriteFile(t, configPath, []byte(`
mysql:
  dsn: "user:password@tcp(127.0.0.1:3306)/app"
  max_open_conns: 20
  max_idle_conns: 10
  conn_max_lifetime_sec: 1800
`))

	novaconfig.SetConfigPath(configPath)
	if err := novaconfig.Reload(); err != nil {
		t.Fatalf("reload config: %v", err)
	}
	resetForTest()

	if err := Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if got := reg.Definitions(); len(got) != 1 || got[0] != "default" {
		t.Fatalf("registered definitions = %v, want [default]", got)
	}
}

func TestParseMySQLConfigAcceptsConnectionPoolAliases(t *testing.T) {
	got := parseMySQLConfig(map[string]any{
		"dsn":                   "user:password@tcp(127.0.0.1:3306)/app",
		"max_open_conns":        20,
		"max_idle_conns":        10,
		"conn_max_lifetime_sec": 1800,
	})

	if got.MaxOpen != 20 {
		t.Fatalf("MaxOpen = %d, want 20", got.MaxOpen)
	}
	if got.MaxIdle != 10 {
		t.Fatalf("MaxIdle = %d, want 10", got.MaxIdle)
	}
	if got.ConnMaxLifetime != 1800 {
		t.Fatalf("ConnMaxLifetime = %d, want 1800", got.ConnMaxLifetime)
	}
}

func requireWriteFile(t *testing.T, path string, contents []byte) {
	t.Helper()
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func resetForTest() {
	initialized = false
	reg = registry.New[*sql.DB]()
}
