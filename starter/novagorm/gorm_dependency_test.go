package novagorm

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestPackageDependsOnNovaMySQL(t *testing.T) {
	imports := productionImports(t)
	for _, path := range imports {
		if path == "github.com/luaxlou/nova/starter/novamysql" {
			return
		}
	}
	t.Fatalf("novagorm should adapt novamysql rather than open MySQL directly")
}

func TestPackageDoesNotImportMySQLDriverDirectly(t *testing.T) {
	imports := productionImports(t)
	for _, path := range imports {
		if path == "github.com/go-sql-driver/mysql" {
			t.Fatalf("novagorm should reuse novamysql connections; found direct driver import")
		}
	}
}

func TestDefaultDBFunctionReturnsGormDB(t *testing.T) {
	var _ func() (*gorm.DB, error) = DB
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
