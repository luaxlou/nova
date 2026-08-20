package novamysql

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestPackageDoesNotDependOnGorm(t *testing.T) {
	files, err := parser.ParseDir(token.NewFileSet(), ".", func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse package: %v", err)
	}

	for _, pkg := range files {
		for _, file := range pkg.Files {
			for _, imported := range file.Imports {
				path := strings.Trim(imported.Path.Value, `"`)
				if strings.HasPrefix(path, "gorm.io/") {
					t.Fatalf("novamysql must stay database/sql only; found GORM import %q", path)
				}
			}
		}
	}
}

func TestPackageDoesNotExposeGormAdapter(t *testing.T) {
	files, err := parser.ParseDir(token.NewFileSet(), ".", func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse package: %v", err)
	}

	for _, pkg := range files {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if ok && fn.Recv == nil && strings.EqualFold(fn.Name.Name, "gorm") {
					t.Fatalf("novamysql must not expose ORM adapters; found function %s", fn.Name.Name)
				}
			}
		}
	}
}
