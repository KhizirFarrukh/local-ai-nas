package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzWriteFiles checks two properties for any file name and content:
// WriteFiles never creates anything outside its root, and when it succeeds
// the file reads back through the same name with the same content.
//
// The seeds below run with every "go test". To search for new failures:
//
//	go test -run '^$' -fuzz '^FuzzWriteFiles$' -fuzztime 30s ./internal/testutil
func FuzzWriteFiles(f *testing.F) {
	f.Add("a.txt", "hello")
	f.Add("dir/sub/b.txt", "")
	f.Add("../escape.txt", "x")
	f.Add("a/../../escape.txt", "x")
	f.Add("/abs.txt", "x")
	f.Add("a/./b.txt", "dot segment")
	f.Add("ñandú.txt", "unicode")
	f.Add("", "empty name")
	f.Fuzz(func(t *testing.T, name, content string) {
		parent := t.TempDir()
		dir := filepath.Join(parent, "root")
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}

		err := WriteFiles(dir, map[string]string{name: content})
		assertOnlyRoot(t, parent)
		if err != nil {
			return // Rejected names are fine; escaping is not.
		}

		root, err := os.OpenRoot(dir)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = root.Close() }()
		got, err := root.ReadFile(filepath.FromSlash(name))
		if err != nil {
			t.Fatalf("WriteFiles(%q) succeeded but reading it back failed: %v", name, err)
		}
		if string(got) != content {
			t.Errorf("read back %q = %q, want %q", name, got, content)
		}
	})
}
