package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

func TestNewItem(t *testing.T) {
	dir := testutil.StorageRoot(t)
	if err := testutil.WriteFiles(dir, map[string]string{"docs/report.pdf": "12345"}); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		rel      string
		wantKind Kind
		wantName string
		wantSize int64
	}{
		{".", KindDir, "", 0},
		{"docs", KindDir, "docs", 0},
		{"docs/report.pdf", KindFile, "report.pdf", 5},
	}
	for _, tt := range tests {
		info, err := os.Lstat(filepath.Join(dir, filepath.FromSlash(tt.rel)))
		if err != nil {
			t.Fatal(err)
		}
		it := NewItem("u0001", tt.rel, info)
		if it.OwnerID != "u0001" || it.Area != "files" || it.RelPath != tt.rel {
			t.Errorf("%s: owner %q, area %q, path %q; want u0001, files, %q", tt.rel, it.OwnerID, it.Area, it.RelPath, tt.rel)
		}
		if it.Kind != tt.wantKind || it.Name != tt.wantName || it.Size != tt.wantSize {
			t.Errorf("%s: kind %s, name %q, size %d; want %s, %q, %d", tt.rel, it.Kind, it.Name, it.Size, tt.wantKind, tt.wantName, tt.wantSize)
		}
		if !it.ModTime.Equal(info.ModTime()) {
			t.Errorf("%s: ModTime %v, want %v", tt.rel, it.ModTime, info.ModTime())
		}
	}
}

func TestNewItemSymlink(t *testing.T) {
	dir := testutil.StorageRoot(t)
	if err := os.Symlink(filepath.Join(dir, "missing"), filepath.Join(dir, "link")); err != nil {
		t.Skipf("cannot create a symbolic link here: %v", err)
	}
	info, err := os.Lstat(filepath.Join(dir, "link"))
	if err != nil {
		t.Fatal(err)
	}
	if it := NewItem("u0001", "link", info); it.Kind != KindSymlink || it.Size != 0 {
		t.Errorf("symlink item = %+v, want kind symlink and no size", it)
	}
}
