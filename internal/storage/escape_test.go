package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

func TestIsEscape(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()

	_, err = root.Open("../outside")
	if !IsEscape(err) {
		t.Errorf("os.Root error for ../outside = %v; IsEscape does not recognize it", err)
	}
	if testutil.TrySymlink(t, t.TempDir(), filepath.Join(dir, "link")) {
		if _, err := root.Lstat("link/x"); !IsEscape(err) {
			t.Errorf("os.Root error through a link to the outside = %v; IsEscape does not recognize it", err)
		}
	}
	for _, other := range []error{nil, os.ErrNotExist, errors.New("permission denied")} {
		if IsEscape(other) {
			t.Errorf("IsEscape(%v) = true", other)
		}
	}
}
