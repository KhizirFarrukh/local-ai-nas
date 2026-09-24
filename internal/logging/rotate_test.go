package logging

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

func readDir(t *testing.T, dir string) map[string]string {
	t.Helper()
	files, err := testutil.ReadFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestRotatingFileRotates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nas.log")
	r, err := OpenRotatingFile(path, 10, 2)
	if err != nil {
		t.Fatal(err)
	}
	// Each line is 5 bytes, so two lines fill a 10-byte file.
	for _, line := range []string{"aaaa\n", "bbbb\n", "cccc\n", "dddd\n", "eeee\n", "ffff\n", "gggg\n"} {
		if _, err := r.Write([]byte(line)); err != nil {
			t.Fatalf("Write(%q): %v", line, err)
		}
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"nas.log":   "gggg\n",
		"nas.log.1": "eeee\nffff\n",
		"nas.log.2": "cccc\ndddd\n",
		// "aaaa bbbb" was the oldest and was deleted (max 2 rotated files).
	}
	if diff := cmp.Diff(want, readDir(t, dir)); diff != "" {
		t.Errorf("log files (-want +got):\n%s", diff)
	}
}

func TestRotatingFileAppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nas.log")
	if err := os.WriteFile(path, []byte("old1\nold2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	r, err := OpenRotatingFile(path, 12, 3)
	if err != nil {
		t.Fatal(err)
	}
	// 10 existing bytes + 5 > 12, so the first write rotates.
	if _, err := r.Write([]byte("new1\n")); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"nas.log": "new1\n", "nas.log.1": "old1\nold2\n"}
	if diff := cmp.Diff(want, readDir(t, dir)); diff != "" {
		t.Errorf("log files (-want +got):\n%s", diff)
	}
}

func TestRotatingFileOversizedWrite(t *testing.T) {
	dir := t.TempDir()
	r, err := OpenRotatingFile(filepath.Join(dir, "nas.log"), 4, 1)
	if err != nil {
		t.Fatal(err)
	}
	big := strings.Repeat("x", 10)
	for range 2 {
		if _, err := r.Write([]byte(big)); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"nas.log": big, "nas.log.1": big}
	if diff := cmp.Diff(want, readDir(t, dir)); diff != "" {
		t.Errorf("log files (-want +got):\n%s", diff)
	}
}

func TestRotatingFileConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	r, err := OpenRotatingFile(filepath.Join(dir, "nas.log"), 1000, 50)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 100 {
				if _, err := r.Write([]byte("0123456789\n")); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Wait()
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	total := 0
	for name, content := range readDir(t, dir) {
		if len(content) > 1000 {
			t.Errorf("%s has %d bytes, more than the 1000-byte limit", name, len(content))
		}
		total += strings.Count(content, "0123456789\n")
	}
	if total != 800 {
		t.Errorf("found %d complete lines, want 800", total)
	}
}

func TestRotatingFileKeepsLoggingWhenRotationFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nas.log")
	// A non-empty directory where the oldest rotated file should be removed
	// makes every rotation fail, on Linux and Windows alike.
	if err := testutil.WriteFiles(dir, map[string]string{"nas.log.1/blocker": ""}); err != nil {
		t.Fatal(err)
	}
	r, err := OpenRotatingFile(path, 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{"one\n", "two\n", "three\n"} {
		if _, err := r.Write([]byte(line)); err != nil {
			t.Fatalf("Write(%q) = %v, want the line kept in the current file", line, err)
		}
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "one\ntwo\nthree\n" {
		t.Errorf("current file = %q, want all three lines", got)
	}
}

func TestRotatingFileClosed(t *testing.T) {
	r, err := OpenRotatingFile(filepath.Join(t.TempDir(), "nas.log"), 100, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
	if _, err := r.Write([]byte("x")); !errors.Is(err, fs.ErrClosed) {
		t.Errorf("Write after Close = %v, want fs.ErrClosed", err)
	}
}

func TestOpenRotatingFileErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := OpenRotatingFile(filepath.Join(dir, "nas.log"), 0, 1); err == nil {
		t.Error("max size 0 accepted")
	}
	if _, err := OpenRotatingFile(filepath.Join(dir, "nas.log"), 10, 0); err == nil {
		t.Error("max files 0 accepted")
	}
	if _, err := OpenRotatingFile(filepath.Join(dir, "missing", "nas.log"), 10, 1); err == nil {
		t.Error("a path in a missing directory was opened")
	}
}
