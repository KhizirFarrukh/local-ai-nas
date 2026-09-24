package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

func TestIsNotFound(t *testing.T) {
	dir := testutil.StorageRoot(t)
	if err := testutil.WriteFiles(dir, map[string]string{"a.txt": "x"}); err != nil {
		t.Fatal(err)
	}
	_, missing := os.Lstat(filepath.Join(dir, "missing"))
	_, throughFile := os.Lstat(filepath.Join(dir, "a.txt", "inside"))
	_, tooLong := os.Lstat(filepath.Join(dir, strings.Repeat("n", 300)))
	for name, err := range map[string]error{"missing": missing, "through a file": throughFile, "a name too long": tooLong} {
		if !IsNotFound(err) {
			t.Errorf("%s: IsNotFound(%v) = false", name, err)
		}
	}
	for _, err := range []error{nil, errors.New("disk on fire"), os.ErrPermission, os.ErrExist} {
		if IsNotFound(err) {
			t.Errorf("IsNotFound(%v) = true", err)
		}
	}
}
