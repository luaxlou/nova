package novagorm

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/luaxlou/nova/starter/internal/registry"
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

func TestPackageDoesNotDependOnNovaMySQL(t *testing.T) {
	imports := productionImports(t)
	for _, path := range imports {
		if path == "github.com/luaxlou/nova/starter/novamysql" {
			t.Fatalf("novagorm should be dynamically assembled and must not import novamysql directly")
		}
	}
}

func TestDefaultDBFunctionReturnsGormDB(t *testing.T) {
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
	defs, defaultName := buildDefinitions(map[string]any{
		"driver": "mysql",
		"dsn":    "root:password@tcp(localhost:3306)/app",
	})

	if defaultName != "default" {
		t.Fatalf("defaultName = %q, want default", defaultName)
	}
	if defs["default"] == nil {
		t.Fatalf("default definition was not built")
	}
}

func TestBuildDefinitionsSupportsNamedDirectMySQLDialectors(t *testing.T) {
	defs, defaultName := buildDefinitions(map[string]any{
		"default": "reporting",
		"instances": map[string]any{
			"reporting": map[string]any{
				"driver": "mysql",
				"dsn":    "analytics:password@tcp(localhost:3306)/analytics",
			},
		},
	})

	if defaultName != "reporting" {
		t.Fatalf("defaultName = %q, want reporting", defaultName)
	}
	if defs["reporting"] == nil {
		t.Fatalf("reporting definition was not built")
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
}
