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
	reg = registry.New[*gorm.DB]()
	manualDefinitions = map[string]Builder{}
	selectedInstanceName = ""
}
