package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

func TestFileID(t *testing.T) {
	dir := testutil.StorageRoot(t)
	if err := testutil.WriteFiles(dir, map[string]string{"a": "1", "b": "1"}); err != nil {
		t.Fatal(err)
	}
	idOf := func(name string) string {
		t.Helper()
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		id, err := FileID(f)
		if err != nil || id == "" {
			t.Fatalf("FileID(%s) = %q, %v", name, id, err)
		}
		return id
	}
	a1, a2, b := idOf("a"), idOf("a"), idOf("b")
	if a1 != a2 {
		t.Errorf("two handles to one file: %q and %q", a1, a2)
	}
	if a1 == b {
		t.Errorf("two files share the ID %q", a1)
	}
	// Replacing a with b (a rename) gives the name b's identity.
	if err := os.Rename(filepath.Join(dir, "b"), filepath.Join(dir, "a")); err != nil {
		t.Fatal(err)
	}
	if got := idOf("a"); got != b {
		t.Errorf("after the rename, a has ID %q, want b's %q", got, b)
	}
	closed, err := os.Open(filepath.Join(dir, "a"))
	if err != nil {
		t.Fatal(err)
	}
	_ = closed.Close()
	if _, err := FileID(closed); err == nil {
		t.Error("FileID of a closed file succeeded")
	}
}
