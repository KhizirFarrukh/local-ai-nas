// Package testutil holds helpers shared by the tests of the other packages:
// a temporary storage root, a test HTTP server on a loopback port, and
// fixture file trees. It is imported only from _test.go files.
package testutil

import (
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
)

// StorageRoot returns a new, empty directory to use as a storage root. It
// is removed when the test ends. Symlinks in the path are resolved (for
// example /var -> /private/var on macOS), so paths built from it compare
// equal to the paths the server reports.
func StorageRoot(t testing.TB) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}
	return dir
}

// NewServer starts h on 127.0.0.1 with a free port and closes it when the
// test ends. The server never listens on other interfaces (NFR-020).
func NewServer(t testing.TB, h http.Handler) *httptest.Server {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on 127.0.0.1:0: %v", err)
	}
	srv := httptest.NewUnstartedServer(h)
	if err := srv.Listener.Close(); err != nil {
		t.Fatalf("close default listener: %v", err)
	}
	srv.Listener = l
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}

// WriteFiles creates each file in files under dir, with parent directories
// as needed. Keys are slash-separated paths relative to dir; values are the
// file contents. All writes go through os.Root, so a name such as
// "../x" or an absolute path returns an error and writes nothing outside dir.
func WriteFiles(dir string, files map[string]string) (err error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer closeWith(root, &err)
	for name, content := range files {
		osName := filepath.FromSlash(name)
		if parent := path.Dir(name); parent != "." {
			if err := root.MkdirAll(filepath.FromSlash(parent), 0o755); err != nil {
				return err
			}
		}
		if err := root.WriteFile(osName, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ReadFiles returns every regular file under dir, keyed by its
// slash-separated path relative to dir. Directories are not listed. Reads
// go through os.Root, like the writes in WriteFiles.
func ReadFiles(dir string) (files map[string]string, err error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer closeWith(root, &err)
	fsys := root.FS()
	files = make(map[string]string)
	err = fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || !d.Type().IsRegular() {
			return walkErr
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		files[name] = string(data)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// closeWith closes c and stores its error in *err unless *err already
// holds an error.
func closeWith(c io.Closer, err *error) {
	if cerr := c.Close(); cerr != nil && *err == nil {
		*err = cerr
	}
}

// Symlink creates the symbolic link link pointing to target. When the OS
// refuses (Windows without the symlink privilege or Developer Mode), the
// test is skipped with the reason, so link tests still run wherever links
// are allowed, such as the Windows CI runner.
func Symlink(t testing.TB, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("cannot create a symbolic link here: %v", err)
		}
		t.Fatalf("create the symbolic link %s: %v", link, err)
	}
}

// TrySymlink is Symlink for tests that also check other things: where the
// OS refuses links, it logs the reason and returns false instead of
// skipping the whole test.
func TrySymlink(t testing.TB, target, link string) bool {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		if runtime.GOOS == "windows" {
			t.Logf("skipping the symbolic link cases: %v", err)
			return false
		}
		t.Fatalf("create the symbolic link %s: %v", link, err)
	}
	return true
}
