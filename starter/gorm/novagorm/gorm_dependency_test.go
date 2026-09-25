package novagorm

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/luaxlou/nova/internal/registry"
	"gorm.io/gorm"
)

func TestPackageDoesNotImportMySQLDriverDirectly(t *testing.T) {
	imports := productionImports(t)
	for _, path := range imports {
		if path == "github.com/go-sql-driver/mysql" {
			t.Fatalf("novagorm should use GORM dialectors instead of importing database drivers directly; found %q", path)
		}
	}
}

func TestDBFunctionReturnsGormDB(t *testing.T) {
	var _ func() (*gorm.DB, error) = DB
}

func TestRegisterProvidesDynamicAssembly(t *testing.T) {
	resetForTest()
	want := &gorm.DB{}

	Register("custom", func(name string) (*gorm.DB, error) {
		if name != "custom" {
			t.Fatalf("builder name = %q, want custom", name)
		}
		return want, nil
	})

	got, err := Named("custom").DB()
	if err != nil {
		t.Fatalf("Named(custom).DB() error = %v", err)
	}
	if got != want {
		t.Fatalf("Named(custom).DB() returned %#v, want registered db", got)
	}
}

func TestBuildDefinitionsSupportsDirectMySQLDialector(t *testing.T) {
	defs, selectedName := buildDefinitions(map[string]any{
		"driver": "mysql",
		"mysql": map[string]any{
			"dsn": "root:password@tcp(localhost:3306)/app",
		},
	})

	if selectedName != singletonName {
		t.Fatalf("selected name = %q, want %s", selectedName, singletonName)
	}
	if defs[singletonName] == nil {
		t.Fatalf("single definition was not built")
	}
}

func TestBuildDefinitionsSupportsNamedDirectMySQLDialectors(t *testing.T) {
	defs, selectedName := buildDefinitions(map[string]any{
		"main": map[string]any{
			"driver": "mysql",
			"mysql": map[string]any{
				"dsn": "root:password@tcp(localhost:3306)/app",
			},
		},
		"analytics": map[string]any{
			"driver": "mysql",
			"mysql": map[string]any{
				"dsn": "analytics:password@tcp(localhost:3306)/analytics",
			},
		},
	})

	if selectedName != "" {
		t.Fatalf("selected name = %q, want empty selection for multiple instances", selectedName)
	}
	if defs["main"] == nil {
		t.Fatalf("main definition was not built")
	}
	if defs["analytics"] == nil {
		t.Fatalf("analytics definition was not built")
	}
}

func TestBuildDefinitionsSelectsOnlyNamedInstanceWhenThereIsOne(t *testing.T) {
	defs, selectedName := buildDefinitions(map[string]any{
		"analytics": map[string]any{
			"driver": "mysql",
			"mysql": map[string]any{
				"dsn": "analytics:password@tcp(localhost:3306)/analytics",
			},
		},
	})

	if selectedName != "analytics" {
		t.Fatalf("selected name = %q, want analytics", selectedName)
	}
	if defs["analytics"] == nil {
		t.Fatalf("analytics definition was not built")
	}
}

func TestParseGormConfigKeepsMySQLConfigUnderDriver(t *testing.T) {
	got := parseGormConfig(map[string]any{
		"driver": "mysql",
		"mysql": map[string]any{
			"dsn":                   "root:password@tcp(localhost:3306)/app",
			"max_open_conns":        20,
			"max_idle_conns":        10,
			"conn_max_lifetime_sec": 1800,
		},
	})

	if got.Driver != "mysql" {
		t.Fatalf("Driver = %q, want mysql", got.Driver)
	}
	if got.MySQL.DSN == "" {
		t.Fatalf("MySQL.DSN was empty")
	}
	if got.MySQL.MaxOpen != 20 || got.MySQL.MaxIdle != 10 || got.MySQL.ConnMaxLifetime != 1800 {
		t.Fatalf("MySQL pool config = %#v, want parsed pool fields", got.MySQL)
	}
}

func TestParseGormConfigKeepsPostgresConfigUnderDriver(t *testing.T) {
	got := parseGormConfig(map[string]any{
		"driver": "postgres",
		"postgres": map[string]any{
			"dsn":                    "host=127.0.0.1 user=nova password=secret dbname=nova port=5432 sslmode=disable",
			"max_open":               20,
			"max_idle":               10,
			"conn_max_lifetime":      1800,
			"conn_max_idle_time":     300,
			"prefer_simple_protocol": true,
		},
	})

	if got.Driver != "postgres" || got.Postgres.DSN == "" || !got.Postgres.PreferSimpleProtocol {
		t.Fatalf("postgres config = %#v", got)
	}
	if got.Postgres.MaxOpen != 20 || got.Postgres.MaxIdle != 10 || got.Postgres.ConnMaxLifetime != 1800 || got.Postgres.ConnMaxIdleTime != 300 {
		t.Fatalf("postgres pool config = %#v", got.Postgres)
	}
}

func TestBuildDefinitionsSupportsNamedMySQLAndPostgres(t *testing.T) {
	defs, selectedName := buildDefinitions(map[string]any{
		"default": "main",
		"main": map[string]any{
			"driver": "postgres",
			"postgres": map[string]any{
				"dsn": "host=127.0.0.1 user=nova dbname=main port=5432 sslmode=disable",
			},
		},
		"legacy": map[string]any{
			"driver": "mysql",
			"mysql": map[string]any{
				"dsn": "root:password@tcp(localhost:3306)/legacy",
			},
		},
	})

	if selectedName != "main" {
		t.Fatalf("selected name = %q, want main", selectedName)
	}
	if defs["main"] == nil || defs["legacy"] == nil {
		t.Fatalf("definitions = %#v, want main and legacy", defs)
	}
}

func TestBuildDefinitionsSupportsDirectPostgresDialector(t *testing.T) {
	defs, selectedName := buildDefinitions(map[string]any{
		"driver": "postgres",
		"postgres": map[string]any{
			"dsn": "host=127.0.0.1 user=nova dbname=app port=5432 sslmode=disable",
		},
	})

	if selectedName != singletonName || defs[singletonName] == nil {
		t.Fatalf("selected name = %q, definitions = %#v", selectedName, defs)
	}
}

func TestConfiguredConnectionRejectsPostgresqlAlias(t *testing.T) {
	_, err := newConfiguredConnection(gormConfig{Driver: "postgresql"})
	if err == nil || !strings.Contains(err.Error(), `unsupported gorm driver "postgresql"`) {
		t.Fatalf("error = %v, want unsupported postgresql alias", err)
	}
}

func TestPostgresMissingDSNDoesNotLeakConfiguration(t *testing.T) {
	const sentinelPassword = "never-echo-this-password"
	defs, _ := buildDefinitions(map[string]any{
		"analytics": map[string]any{
			"driver": "postgres",
			"postgres": map[string]any{
				"password": sentinelPassword,
			},
		},
	})

	_, err := defs["analytics"]("analytics")
	if err == nil || !strings.Contains(err.Error(), "analytics") || !strings.Contains(err.Error(), "postgres.dsn") {
		t.Fatalf("error = %v, want instance-aware postgres.dsn error", err)
	}
	if strings.Contains(err.Error(), sentinelPassword) {
		t.Fatalf("error leaked password: %v", err)
	}
}

func TestApplyPostgresPoolConfigHonorsPositiveValuesAndPreservesDefaults(t *testing.T) {
	db, _ := openGuardTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	if err := applyPostgresPoolConfig(db, postgresConfig{MaxOpen: 7}); err != nil {
		t.Fatalf("apply positive pool config: %v", err)
	}
	if got := sqlDB.Stats().MaxOpenConnections; got != 7 {
		t.Fatalf("max open connections = %d, want 7", got)
	}

	if err := applyPostgresPoolConfig(db, postgresConfig{}); err != nil {
		t.Fatalf("apply zero pool config: %v", err)
	}
	if got := sqlDB.Stats().MaxOpenConnections; got != 7 {
		t.Fatalf("max open connections after zero config = %d, want 7", got)
	}
}

func TestRegisteredBuilderErrorIsReturned(t *testing.T) {
	resetForTest()
	wantErr := errors.New("boom")

	Register("broken", func(name string) (*gorm.DB, error) {
		return nil, wantErr
	})

	if _, err := Named("broken").DB(); !errors.Is(err, wantErr) {
		t.Fatalf("Named(broken).DB() error = %v, want %v", err, wantErr)
	}
}

func TestDBRequiresNameWhenMultipleBuildersAreRegistered(t *testing.T) {
	resetForTest()

	Register("main", func(string) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	})
	Register("analytics", func(string) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	})

	if _, err := DB(); err == nil || !strings.Contains(err.Error(), "gorm instance name is required") {
		t.Fatalf("DB() error = %v, want instance name required error", err)
	}
}

func TestOpenMySQLFromSQLDBRejectsNilDB(t *testing.T) {
	if _, err := OpenMySQLFromSQLDB(nil); err == nil {
		t.Fatalf("OpenMySQLFromSQLDB(nil) error = nil, want error")
	}
}

func productionImports(t *testing.T) []string {
	t.Helper()

	files, err := parser.ParseDir(token.NewFileSet(), ".", func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse package: %v", err)
	}

	var imports []string
	for _, pkg := range files {
		for _, file := range pkg.Files {
			for _, imported := range file.Imports {
				imports = append(imports, strings.Trim(imported.Path.Value, `"`))
			}
		}
	}
	return imports
}

func TestPackageExposesNamedInstances(t *testing.T) {
	var _ interface {
		DB() (*gorm.DB, error)
	} = (*gormInstance)(nil)
}

func TestPackageHasNoDirectSQLOpenCall(t *testing.T) {
	files, err := parser.ParseDir(token.NewFileSet(), ".", func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse package: %v", err)
	}

	for _, pkg := range files {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if ok && selector.Sel.Name == "Open" {
					if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == "sql" {
						t.Fatalf("novagorm should not call sql.Open directly")
					}
				}
				return true
			})
		}
	}
}

func resetForTest() {
	initialized = false
	reg = registry.New[*gormResource]()
	manualDefinitions = map[string]Builder{}
	selectedInstanceName = ""
}
