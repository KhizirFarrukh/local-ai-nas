package files

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// forbidden lists what reaches the file system by path. The files package
// must use the storage resolver and os.Root instead (S01.2-T03), so no
// request can touch a path outside its namespace. "*" forbids the whole
// package.
var forbidden = map[string]map[string]bool{
	"os": {
		"Chdir": true, "Chmod": true, "Chown": true, "Chtimes": true, "CopyFS": true,
		"Create": true, "CreateTemp": true, "DirFS": true, "Lchown": true, "Link": true,
		"Lstat": true, "Mkdir": true, "MkdirAll": true, "MkdirTemp": true, "Open": true,
		"OpenFile": true, "OpenInRoot": true, "OpenRoot": true, "ReadDir": true, "ReadFile": true,
		"Readlink": true, "Remove": true, "RemoveAll": true, "Rename": true, "Stat": true,
		"Symlink": true, "Truncate": true, "WriteFile": true,
	},
	"path/filepath":            {"Abs": true, "EvalSymlinks": true, "Glob": true, "Walk": true, "WalkDir": true},
	"io/ioutil":                {"*": true},
	"syscall":                  {"*": true},
	"golang.org/x/sys/unix":    {"*": true},
	"golang.org/x/sys/windows": {"*": true},
}

// violations returns every forbidden import or call in f.
func violations(fset *token.FileSet, f *ast.File) []string {
	var out []string
	imports := map[string]string{} // local name -> import path
	for _, imp := range f.Imports {
		p, _ := strconv.Unquote(imp.Path.Value)
		name := p[strings.LastIndex(p, "/")+1:]
		if imp.Name != nil {
			name = imp.Name.Name
		}
		imports[name] = p
		if forbidden[p]["*"] {
			out = append(out, "import "+p)
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if pkg, ok := sel.X.(*ast.Ident); ok && forbidden[imports[pkg.Name]][sel.Sel.Name] {
			out = append(out, imports[pkg.Name]+"."+sel.Sel.Name)
		}
		return true
	})
	return out
}

// TestNoDirectFileSystemAccess is the architecture test of S01.2-T03: every
// non-test file of this package does its file I/O only through the storage
// resolver and os.Root.
func TestNoDirectFileSystemAccess(t *testing.T) {
	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	fset := token.NewFileSet()
	for _, src := range sources {
		if strings.HasSuffix(src, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, src, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		checked++
		for _, v := range violations(fset, f) {
			t.Errorf("%s uses %s; file access goes through the storage resolver and os.Root", src, v)
		}
	}
	if checked == 0 {
		t.Fatal("no source files were checked")
	}
}

// TestViolationsDetected makes sure the architecture test would catch a
// breach.
func TestViolationsDetected(t *testing.T) {
	src := `package files

import (
	"os"
	fp "path/filepath"
	"syscall"
)

func leak() {
	_, _ = os.ReadFile("/etc/passwd")
	_ = fp.Walk
	_ = os.Getpid // not file access: allowed
	_ = syscall.Getpid
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "bad.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"import syscall", "os.ReadFile", "path/filepath.Walk"}
	if diff := cmp.Diff(want, violations(fset, f)); diff != "" {
		t.Errorf("violations (-want +got):\n%s", diff)
	}
}

// TestNoLinksCreated is part of the symlink policy (S01.6-T03): the API
// never creates symbolic links, so no non-test file of this package calls
// a Symlink function or method (os.Symlink is also on the forbidden list,
// and os.Root.Symlink would slip past it).
func TestNoLinksCreated(t *testing.T) {
	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, src := range sources {
		if strings.HasSuffix(src, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, src, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, pos := range symlinkCalls(fset, f) {
			t.Errorf("%s uses Symlink; the API never creates symbolic links", pos)
		}
	}
}

// symlinkCalls returns the positions of every use of a name Symlink, as
// a function or a method.
func symlinkCalls(fset *token.FileSet, f *ast.File) []string {
	var found []string
	ast.Inspect(f, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "Symlink" {
			found = append(found, fset.Position(sel.Pos()).String())
		}
		return true
	})
	return found
}

func TestSymlinkCallsDetected(t *testing.T) {
	src := `package files

import "os"

func bad(root *os.Root) {
	_ = root.Symlink("target", "link")
	_ = os.Symlink("target", "link")
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "bad.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := symlinkCalls(fset, f); len(got) != 2 {
		t.Errorf("symlinkCalls found %v, want both calls", got)
	}
}
