package nova

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type starterIndex struct {
	SchemaVersion int            `yaml:"schema_version"`
	Starters      []starterEntry `yaml:"starters"`
}

type starterEntry struct {
	ID           string        `yaml:"id"`
	Package      string        `yaml:"package"`
	Category     string        `yaml:"category"`
	Owner        string        `yaml:"owner"`
	Status       string        `yaml:"status"`
	ConfigRoot   string        `yaml:"config_root"`
	Lifecycle    []string      `yaml:"lifecycle"`
	Docs         string        `yaml:"docs"`
	Capabilities []string      `yaml:"capabilities"`
	Drivers      []driverEntry `yaml:"drivers"`
}

type driverEntry struct {
	ID             string `yaml:"id"`
	Status         string `yaml:"status"`
	VerifiedServer string `yaml:"verified_server"`
}

func TestStarterIndexMatchesPublishedStarters(t *testing.T) {
	content, err := os.ReadFile("starter-index.yaml")
	if err != nil {
		t.Fatalf("read starter-index.yaml: %v", err)
	}

	var index starterIndex
	if err := yaml.Unmarshal(content, &index); err != nil {
		t.Fatalf("parse starter-index.yaml: %v", err)
	}
	if index.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", index.SchemaVersion)
	}

	wantPackages := map[string]string{
		"novaconfig":    "github.com/luaxlou/nova/starter/config/novaconfig",
		"novagin":       "github.com/luaxlou/nova/starter/http/novagin",
		"novaredis":     "github.com/luaxlou/nova/starter/cache/novaredis",
		"novaoss":       "github.com/luaxlou/nova/starter/aliyun/novaoss",
		"novaqwen":      "github.com/luaxlou/nova/starter/ai/novaqwen",
		"novawebsocket": "github.com/luaxlou/nova/starter/realtime/novawebsocket",
		"novagorm":      "github.com/luaxlou/nova/starter/gorm/novagorm",
	}

	seenIDs := map[string]bool{}
	seenPackages := map[string]bool{}
	for _, starter := range index.Starters {
		if starter.ID == "" || starter.Package == "" || starter.Category == "" || starter.Owner == "" || starter.Status == "" || starter.ConfigRoot == "" || starter.Docs == "" {
			t.Fatalf("starter has missing required field: %#v", starter)
		}
		if len(starter.Lifecycle) == 0 || len(starter.Capabilities) == 0 {
			t.Fatalf("starter %q must declare lifecycle and capabilities", starter.ID)
		}
		if seenIDs[starter.ID] {
			t.Fatalf("duplicate starter id %q", starter.ID)
		}
		if seenPackages[starter.Package] {
			t.Fatalf("duplicate starter package %q", starter.Package)
		}
		seenIDs[starter.ID] = true
		seenPackages[starter.Package] = true

		wantPackage, published := wantPackages[starter.ID]
		if !published {
			t.Fatalf("starter %q is not in the published README inventory", starter.ID)
		}
		if starter.Package != wantPackage {
			t.Fatalf("starter %q package = %q, want %q", starter.ID, starter.Package, wantPackage)
		}
		packageDir := strings.TrimPrefix(starter.Package, "github.com/luaxlou/nova/")
		if info, err := os.Stat(filepath.FromSlash(packageDir)); err != nil || !info.IsDir() {
			t.Fatalf("starter %q package directory %q is missing", starter.ID, starter.Package)
		}
		if info, err := os.Stat(filepath.FromSlash(starter.Docs)); err != nil || info.IsDir() {
			t.Fatalf("starter %q docs file %q is missing", starter.ID, starter.Docs)
		}
	}

	for id := range wantPackages {
		if !seenIDs[id] {
			t.Fatalf("published starter %q is missing from starter-index.yaml", id)
		}
	}

	novagorm := findStarter(index.Starters, "novagorm")
	if novagorm == nil {
		t.Fatal("novagorm is missing from starter-index.yaml")
	}
	assertStableDriver(t, novagorm.Drivers, "mysql", "")
	assertStableDriver(t, novagorm.Drivers, "postgres", "PostgreSQL 18")
}

func findStarter(starters []starterEntry, id string) *starterEntry {
	for i := range starters {
		if starters[i].ID == id {
			return &starters[i]
		}
	}
	return nil
}

func assertStableDriver(t *testing.T, drivers []driverEntry, id, verifiedServer string) {
	t.Helper()
	for _, driver := range drivers {
		if driver.ID != id {
			continue
		}
		if driver.Status != "stable" || driver.VerifiedServer != verifiedServer {
			t.Fatalf("driver %q = %#v, want stable with verified_server %q", id, driver, verifiedServer)
		}
		return
	}
	t.Fatalf("driver %q is missing", id)
}
