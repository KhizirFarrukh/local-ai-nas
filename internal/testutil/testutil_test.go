package testutil

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStorageRoot(t *testing.T) {
	dir := StorageRoot(t)
	if !filepath.IsAbs(dir) {
		t.Errorf("StorageRoot() = %q, want an absolute path", dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read storage root: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("storage root has %d entries, want 0", len(entries))
	}
	if other := StorageRoot(t); other == dir {
		t.Errorf("two calls returned the same directory %q", dir)
	}
}

func TestNewServer(t *testing.T) {
	srv := NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "path="+r.URL.Path)
	}))

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	if got := u.Hostname(); got != "127.0.0.1" {
		t.Errorf("server host = %q, want 127.0.0.1", got)
	}

	resp, err := srv.Client().Get(srv.URL + "/api/v1/ping")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK || string(body) != "path=/api/v1/ping" {
		t.Errorf("GET = %d %q, want 200 %q", resp.StatusCode, body, "path=/api/v1/ping")
	}
}

// TestWriteReadFiles is an integration test on a real temporary storage
// root: a fixture tree written with WriteFiles reads back unchanged.
func TestWriteReadFiles(t *testing.T) {
	dir := StorageRoot(t)
	want := map[string]string{
		"a.txt":               "alpha",
		"empty.bin":           "",
		"docs/2026/notes.md":  "# notes\n",
		"docs/2026/space d.x": "names with spaces work",
		"unicode/ñandú.txt":   "non-ASCII names work",
	}
	if err := WriteFiles(dir, want); err != nil {
		t.Fatalf("WriteFiles: %v", err)
	}
	got, err := ReadFiles(dir)
	if err != nil {
		t.Fatalf("ReadFiles: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ReadFiles mismatch (-want +got):\n%s", diff)
	}
}

func TestWriteFilesRejectsEscapes(t *testing.T) {
	names := []string{
		"../outside.txt",
		"a/../../outside.txt",
		"/abs.txt",
	}
	if filepath.Separator == '\\' {
		names = append(names, `C:\abs.txt`, `..\outside.txt`, `\\host\share\x.txt`)
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			dir := filepath.Join(parent, "root")
			if err := os.Mkdir(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := WriteFiles(dir, map[string]string{name: "x"}); err == nil {
				t.Errorf("WriteFiles(%q) succeeded, want an error", name)
			}
			assertOnlyRoot(t, parent)
		})
	}
}

// assertOnlyRoot fails the test if parent contains anything besides the
// directory "root".
func assertOnlyRoot(t *testing.T, parent string) {
	t.Helper()
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if diff := cmp.Diff([]string{"root"}, names); diff != "" {
		t.Errorf("something was written outside the root (-want +got):\n%s", diff)
	}
}

func TestSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	Symlink(t, target, link) // skips where links are not allowed
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("Lstat(link) = %v, %v; want a symbolic link", info, err)
	}
}

func TestTrySymlink(t *testing.T) {
	dir := t.TempDir()
	if TrySymlink(t, dir, filepath.Join(dir, "link")) {
		if _, err := os.Lstat(filepath.Join(dir, "link")); err != nil {
			t.Error(err)
		}
	}
}
