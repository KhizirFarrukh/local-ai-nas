package storage

import (
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// wantDirs is the complete layout, relative to the root.
var wantDirs = []string{
	".local-ai-nas",
	".local-ai-nas/db",
	".local-ai-nas/logs",
	".local-ai-nas/tmp",
	".local-ai-nas/tmp/uploads",
	"files",
	"files/u0001",
	"photos",
	"photos/u0001",
}

// listDirs returns every directory under root with its modification time.
func listDirs(t *testing.T, root string) map[string]time.Time {
	t.Helper()
	dirs := map[string]time.Time{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || p == root || !d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		dirs[filepath.ToSlash(rel)] = info.ModTime()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return dirs
}

func names(m map[string]time.Time) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestInitCreatesLayout(t *testing.T) {
	root := filepath.Join(testutil.StorageRoot(t), "nas") // the root itself is created too
	l := NewLayout(root, Options{})
	unknown, err := l.Init()
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if len(unknown) != 0 {
		t.Errorf("unknown entries = %v, want none", unknown)
	}
	got := names(listDirs(t, root))
	slices.Sort(got)
	if diff := cmp.Diff(wantDirs, got); diff != "" {
		t.Errorf("layout (-want +got):\n%s", diff)
	}
	if files, err := testutil.ReadFiles(root); err != nil || len(files) != 0 {
		t.Errorf("Init left files behind: %v, %v", files, err)
	}
	if got, want := l.Area(FilesArea, DefaultNamespace), filepath.Join(root, "files", "u0001"); got != want {
		t.Errorf("Area = %q, want %q", got, want)
	}
}

func TestInitTwiceChangesNothing(t *testing.T) {
	root := testutil.StorageRoot(t)
	l := NewLayout(root, Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	if err := testutil.WriteFiles(l.Area(FilesArea, DefaultNamespace), map[string]string{"docs/a.txt": "keep me"}); err != nil {
		t.Fatal(err)
	}
	before := listDirs(t, root)
	time.Sleep(20 * time.Millisecond) // a changed time would differ now

	if _, err := l.Init(); err != nil {
		t.Fatalf("second Init: %v", err)
	}
	if diff := cmp.Diff(before, listDirs(t, root)); diff != "" {
		t.Errorf("the second Init changed directories or their times (-before +after):\n%s", diff)
	}
	files, err := testutil.ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(map[string]string{"files/u0001/docs/a.txt": "keep me"}, files); diff != "" {
		t.Errorf("files (-want +got):\n%s", diff)
	}
}

func TestInitReportsUnknownEntries(t *testing.T) {
	root := testutil.StorageRoot(t)
	if err := testutil.WriteFiles(root, map[string]string{"notes.txt": "mine", "Backups/2025.zip": "old"}); err != nil {
		t.Fatal(err)
	}
	unknown, err := NewLayout(root, Options{}).Init()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"Backups", "notes.txt"}, unknown); diff != "" {
		t.Errorf("unknown entries (-want +got):\n%s", diff)
	}
	files, err := testutil.ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if files["notes.txt"] != "mine" || files["Backups/2025.zip"] != "old" {
		t.Errorf("unknown entries were changed: %v", files)
	}
}

func TestInitRejectsNonDirectories(t *testing.T) {
	for _, name := range []string{"files", "photos/u0001", ".local-ai-nas/db"} {
		t.Run(name, func(t *testing.T) {
			root := testutil.StorageRoot(t)
			if err := testutil.WriteFiles(root, map[string]string{name: "not a directory"}); err != nil {
				t.Fatal(err)
			}
			_, err := NewLayout(root, Options{}).Init()
			if err == nil || !strings.Contains(err.Error(), "not a directory") {
				t.Errorf("Init = %v, want a \"not a directory\" error for %s", err, name)
			}
		})
	}
}

func TestInitRejectsSymlinkedArea(t *testing.T) {
	root := testutil.StorageRoot(t)
	elsewhere := t.TempDir()
	if err := os.Symlink(elsewhere, filepath.Join(root, "files")); err != nil {
		t.Skipf("cannot create a symbolic link here (on Windows this needs developer mode or admin rights): %v", err)
	}
	_, err := NewLayout(root, Options{}).Init()
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Errorf("Init = %v, want a symbolic-link error", err)
	}
}

func TestInitRejectsRelativeRoot(t *testing.T) {
	if _, err := NewLayout(filepath.Join("relative", "nas"), Options{}).Init(); err == nil {
		t.Error("Init accepted a relative root")
	}
}

// denyWrites makes dir unwritable for the current user until the test ends:
// with file permissions on Unix, and with an explicit deny entry in the
// directory's access control list on Windows.
func denyWrites(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		u, err := user.Current()
		if err != nil {
			t.Fatal(err)
		}
		if out, err := exec.Command("icacls", dir, "/deny", u.Username+":(W)").CombinedOutput(); err != nil {
			t.Fatalf("icacls /deny: %v: %s", err, out)
		}
		t.Cleanup(func() {
			if out, err := exec.Command("icacls", dir, "/remove:d", u.Username).CombinedOutput(); err != nil {
				t.Errorf("icacls /remove:d: %v: %s", err, out)
			}
		})
		return
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root: file permissions do not stop root from writing")
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}

func TestInitFailsOnReadOnlyRoot(t *testing.T) {
	root := testutil.StorageRoot(t)
	denyWrites(t, root)
	_, err := NewLayout(root, Options{}).Init()
	if err == nil || !strings.Contains(err.Error(), "cannot create") {
		t.Errorf("Init = %v, want a clear \"cannot create\" error", err)
	}
}

func TestInitFailsOnReadOnlyArea(t *testing.T) {
	root := testutil.StorageRoot(t)
	l := NewLayout(root, Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	area := l.Area(FilesArea, DefaultNamespace)
	denyWrites(t, area)
	_, err := l.Init()
	if err == nil || !strings.Contains(err.Error(), "is not writable") || !strings.Contains(err.Error(), area) {
		t.Errorf("Init = %v, want \"%s is not writable\"", err, area)
	}
}
