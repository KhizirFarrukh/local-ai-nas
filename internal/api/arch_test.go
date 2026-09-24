package api

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// noDiskImports are packages the API layer must not import: all file
// access goes through files.Service (S01.3-T01). golangci-lint's depguard
// enforces the same list.
var noDiskImports = []string{"os", "path/filepath", "io/ioutil", "syscall", "golang.org/x/sys/"}

// TestHandlersUseOnlyTheService is the S01.3-T01 architecture test: the
// non-test, non-generated sources of internal/api import nothing that
// reaches the disk, and never build the concrete files service (cmd wires
// it in through Options.Files).
func TestHandlersUseOnlyTheService(t *testing.T) {
	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	checked := 0
	for _, src := range sources {
		if strings.HasSuffix(src, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, src, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		checked++
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			for _, bad := range noDiskImports {
				if p == bad || strings.HasSuffix(bad, "/") && strings.HasPrefix(p, bad) {
					t.Errorf("%s imports %s; the API reaches files only through files.Service", src, p)
				}
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "files" && (sel.Sel.Name == "NewLocal" || sel.Sel.Name == "Local") {
				t.Errorf("%s uses files.%s; the API gets the service through Options.Files", fset.Position(sel.Pos()), sel.Sel.Name)
			}
			return true
		})
	}
	if checked == 0 {
		t.Fatal("no sources checked")
	}
}
