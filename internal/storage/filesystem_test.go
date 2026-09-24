package storage

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

func TestDeviceID(t *testing.T) {
	root := testutil.StorageRoot(t)
	if err := testutil.WriteFiles(root, map[string]string{"a/x": "", "b/y": ""}); err != nil {
		t.Fatal(err)
	}
	a, err := DeviceID(filepath.Join(root, "a"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := DeviceID(filepath.Join(root, "b", "y")) // files work too
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Errorf("two paths in one temp directory have device IDs %d and %d", a, b)
	}
	if _, err := DeviceID(filepath.Join(root, "missing")); err == nil {
		t.Error("DeviceID of a missing path succeeded")
	}
}

func TestFinalizeMode(t *testing.T) {
	l := NewLayout(testutil.StorageRoot(t), Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	if mode, err := l.FinalizeMode(nil); err != nil || mode != FinalizeRename {
		t.Errorf("real devices: %q, %v; want rename", mode, err)
	}
	byPath := func(tmp, files uint64) DeviceFunc {
		return func(p string) (uint64, error) {
			if p == l.TmpUploads {
				return tmp, nil
			}
			return files, nil
		}
	}
	if mode, _ := l.FinalizeMode(byPath(7, 7)); mode != FinalizeRename {
		t.Errorf("same device: %q, want rename", mode)
	}
	if mode, _ := l.FinalizeMode(byPath(7, 8)); mode != FinalizeCopy {
		t.Errorf("different devices: %q, want copy", mode)
	}
	failAt := func(bad string) DeviceFunc {
		return func(p string) (uint64, error) {
			if p == bad {
				return 0, errors.New("stat failed")
			}
			return 1, nil
		}
	}
	for _, bad := range []string{l.TmpUploads, l.Area(FilesArea, DefaultNamespace)} {
		if _, err := l.FinalizeMode(failAt(bad)); err == nil {
			t.Errorf("an error for %s was not returned", bad)
		}
	}
}

func TestWritable(t *testing.T) {
	root := testutil.StorageRoot(t)
	l := NewLayout(root, Options{})
	if err := l.Writable(); err == nil {
		t.Error("Writable succeeded before Init")
	}
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	if err := l.Writable(); err != nil {
		t.Errorf("Writable after Init: %v", err)
	}
	if files, err := testutil.ReadFiles(root); err != nil || len(files) != 0 {
		t.Errorf("Writable left files behind: %v, %v", files, err)
	}
	denyWrites(t, l.Logs)
	if err := l.Writable(); err == nil {
		t.Error("Writable succeeded with a read-only logs directory")
	}
}
