package storage

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

func newTestResolver(t *testing.T) (*Resolver, Layout) {
	t.Helper()
	l := NewLayout(testutil.StorageRoot(t), Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	return NewResolver(l), l
}

func TestResolve(t *testing.T) {
	r, _ := newTestResolver(t)
	tests := []struct {
		in, want string
	}{
		{"/", "."},
		{"/a.txt", "a.txt"},
		{"/docs/a.txt", "docs/a.txt"},
		{"/docs/", "docs"},
		{"/docs//a.txt", "docs/a.txt"},
		{"/./docs/./a.txt", "docs/a.txt"},
		{"/ñandú/report 2026.pdf", "ñandú/report 2026.pdf"},
		{"/.hidden", ".hidden"},
		{"/a..b/c...", "a..b/c..."},
	}
	for _, tt := range tests {
		got, err := r.Resolve(FilesArea, DefaultNamespace, tt.in)
		if err != nil || got != tt.want {
			t.Errorf("Resolve(%q) = %q, %v; want %q", tt.in, got, err, tt.want)
		}
	}
}

func TestResolveRejects(t *testing.T) {
	r, _ := newTestResolver(t)
	tests := []struct {
		area, ns, in string
		want         apperr.Kind
	}{
		{FilesArea, DefaultNamespace, "", apperr.InvalidRequest},
		{FilesArea, DefaultNamespace, "docs/a.txt", apperr.InvalidRequest},
		{FilesArea, DefaultNamespace, "/..", apperr.OutsideRoot},
		{FilesArea, DefaultNamespace, "/../u0002/a.txt", apperr.OutsideRoot},
		{FilesArea, DefaultNamespace, "/docs/../../x", apperr.OutsideRoot},
		{FilesArea, DefaultNamespace, "/docs/..", apperr.OutsideRoot}, // even when it would stay inside
		// The photos area cannot be reached through the files area (S01.2-T06).
		{FilesArea, DefaultNamespace, "/../../photos/u0001/a.jpg", apperr.OutsideRoot},
		{FilesArea, DefaultNamespace, `/docs\a.txt`, apperr.InvalidName},
		{FilesArea, DefaultNamespace, `/..\..\photos`, apperr.InvalidName},
		{FilesArea, DefaultNamespace, "/a\x00b", apperr.InvalidName},
		{PhotosArea, DefaultNamespace, "/a.jpg", apperr.NotAvailable},
		{"music", DefaultNamespace, "/a.mp3", apperr.InvalidRequest},
		{FilesArea, "u1", "/a", apperr.InvalidRequest},
		{FilesArea, "U0001", "/a", apperr.InvalidRequest},
		{FilesArea, "../u0001", "/a", apperr.InvalidRequest},
		{FilesArea, "", "/a", apperr.InvalidRequest},
	}
	if runtime.GOOS == "windows" {
		tests = append(tests,
			struct {
				area, ns, in string
				want         apperr.Kind
			}{FilesArea, DefaultNamespace, "/NUL", apperr.InvalidName},
			struct {
				area, ns, in string
				want         apperr.Kind
			}{FilesArea, DefaultNamespace, "/C:/Windows", apperr.InvalidName},
		)
	}
	for _, tt := range tests {
		got, err := r.Resolve(tt.area, tt.ns, tt.in)
		if err == nil {
			t.Errorf("Resolve(%q, %q, %q) = %q, want a %s error", tt.area, tt.ns, tt.in, got, tt.want)
			continue
		}
		if k := apperr.KindOf(err); k != tt.want {
			t.Errorf("Resolve(%q, %q, %q) error %v has kind %s, want %s", tt.area, tt.ns, tt.in, err, k, tt.want)
		}
	}
}

func TestOpenRoot(t *testing.T) {
	r, l := newTestResolver(t)
	if err := testutil.WriteFiles(l.Area(FilesArea, DefaultNamespace), map[string]string{"docs/a.txt": "hello"}); err != nil {
		t.Fatal(err)
	}
	root, err := r.OpenRoot(FilesArea, DefaultNamespace)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()

	rel, err := r.Resolve(FilesArea, DefaultNamespace, "/docs/a.txt")
	if err != nil {
		t.Fatal(err)
	}
	data, err := root.ReadFile(filepath.FromSlash(rel))
	if err != nil || string(data) != "hello" {
		t.Errorf("ReadFile(%q) = %q, %v", rel, data, err)
	}

	if _, err := r.OpenRoot(PhotosArea, DefaultNamespace); apperr.KindOf(err) != apperr.NotAvailable {
		t.Errorf("OpenRoot(photos) = %v, want not_available", err)
	}
	if _, err := r.OpenRoot(FilesArea, "u0002"); apperr.KindOf(err) != apperr.Internal {
		t.Errorf("OpenRoot of a missing namespace = %v, want an internal error", err)
	}
}

// TestRootIsSecondLayer bypasses Resolve and shows that the os.Root from
// OpenRoot still refuses to leave the namespace (the full attack corpus is
// S01.6-T01).
func TestRootIsSecondLayer(t *testing.T) {
	r, l := newTestResolver(t)
	if err := testutil.WriteFiles(l.Area(PhotosArea, DefaultNamespace), map[string]string{"secret.jpg": "private"}); err != nil {
		t.Fatal(err)
	}
	root, err := r.OpenRoot(FilesArea, DefaultNamespace)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()

	escape := filepath.Join("..", "..", "photos", "u0001", "secret.jpg")
	if _, err := root.ReadFile(escape); err == nil {
		t.Error("os.Root read a file in the photos area")
	}
	if err := root.WriteFile(filepath.Join("..", "outside.txt"), []byte("x"), 0o600); err == nil {
		t.Error("os.Root wrote outside the namespace")
	}
	if _, err := os.Stat(filepath.Join(l.Root, FilesArea, "outside.txt")); !os.IsNotExist(err) {
		t.Error("a file appeared outside the namespace")
	}

	// A symbolic link that points out of the namespace is not followed.
	link := filepath.Join(l.Area(FilesArea, DefaultNamespace), "to-photos")
	if err := os.Symlink(l.Area(PhotosArea, DefaultNamespace), link); err != nil {
		t.Skipf("cannot create a symbolic link here: %v", err)
	}
	if _, err := root.ReadFile(filepath.Join("to-photos", "secret.jpg")); err == nil {
		t.Error("os.Root followed a symbolic link out of the namespace")
	}
}

// TestResolveReservedRule: on Windows the resolver refuses device names
// itself; it reports them with the same rule as the name check, so clients
// get one answer on every OS.
func TestResolveReservedRule(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only Windows refuses device names while resolving")
	}
	r, _ := newTestResolver(t)
	for _, p := range []string{"/NUL", "/docs/con", "/COM1/x"} {
		_, err := r.Resolve(FilesArea, DefaultNamespace, p)
		var e *apperr.Error
		if !errors.As(err, &e) || e.Kind != apperr.InvalidName || e.Rule != RuleReservedName {
			t.Errorf("Resolve(%q) = %v, want invalid_name / reserved_name", p, err)
		}
	}
}
