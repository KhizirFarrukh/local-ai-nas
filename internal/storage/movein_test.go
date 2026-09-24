package storage

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestMoveInto(t *testing.T) {
	r, l := newTestResolver(t)
	src := filepath.Join(l.TmpUploads, "data")
	if err := os.WriteFile(src, []byte("upload"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := r.MoveInto(FilesArea, DefaultNamespace, TempPrefix+"1.part", src); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(l.Area(FilesArea, DefaultNamespace), TempPrefix+"1.part"))
	if err != nil || string(got) != "upload" {
		t.Errorf("moved file: %q, %v", got, err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("the source is still there: %v", err)
	}
	for _, bad := range []string{".", "../x", "/abs", ""} {
		if err := r.MoveInto(FilesArea, DefaultNamespace, bad, src); err == nil {
			t.Errorf("MoveInto(%q) succeeded", bad)
		}
	}
}

func TestIsCrossDevice(t *testing.T) {
	if isCrossDevice(errors.New("other")) || isCrossDevice(&os.LinkError{Err: syscall.ENOENT}) {
		t.Error("an unrelated error counts as cross-device")
	}
	if !isCrossDevice(&os.LinkError{Op: "rename", Err: crossDeviceErrno}) {
		t.Error("the cross-device error is not recognized")
	}
}
