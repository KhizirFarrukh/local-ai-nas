package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"

	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// attackCorpus holds API paths that must be refused on every OS
// (S01.6-T01). It also seeds FuzzResolve.
var attackCorpus = []string{
	// Not an API path at all.
	"", "docs", "docs/../x", `\..\x`,
	// Parent segments.
	"/..", "/../", "/../x", "/../../etc/passwd", "/../../../../../../windows/win.ini",
	"/a/../../x", "/a/b/../../../x", "/a/..", "/./../x", "/a/./../../x", "/a/b/c/../../../../x",
	// Other namespaces, the photos area, internal data.
	"/files/../photos/u0001/a.jpg", "/../u0002/secret.txt", "/../../photos/u0001",
	"/../../.local-ai-nas/db/nas.db",
	// Dot runs and dots with spaces (Windows may read them as "..").
	"/...", "/..../x", "/....", "/.../.../x", "/.. ", "/. .", "/ ..", "/a/.. /b", "/a/... /b", "/a/ ../b",
	// Backslashes.
	`/..\x`, `/a\..\..\x`, `/a\b`, `\\server\share\x`,
	// UNC-like and device paths.
	"//server/share/x", "///x", "//./C:/x", "//?/C:/x",
	// Drive letters.
	"/C:", "/C:/Windows/System32", "/c:/x", "/a/D:/x", "/Z:",
	// Percent-encoded separators and NUL (the path was encoded twice).
	"/..%2Fx", "/..%2fx", "/%2F..%2Fx", "/a%5C..%5Cx", "/a%5c..%5cx", "/a%00b", "/x%00.txt",
	"/%2e%2e%2f", "/a/%2F/b", "/C:%5Cx",
	// NUL bytes.
	"/a\x00b", "/\x00", "/..\x00/x", "/docs/\x00../x",
	// Unicode look-alikes of dot segments (fullwidth full stop, one and
	// two dot leaders, ellipsis), which folding tools read as dots.
	"/\uFF0E\uFF0E/x", "/a/\uFF0E\uFF0E/\uFF0E\uFF0E/x", "/\u2025/x", "/\u2024\u2024/x", "/\uFF0E/\uFF0E\uFF0E", "/\u2026/x",
}

// windowsOnlyCorpus holds names that only Windows refuses at this layer
// (S01.6-T02 refuses them on every OS with its own error codes, including
// the forms with an extension such as "aux.txt", which Go's IsLocal accepts
// since current Windows versions allow them).
var windowsOnlyCorpus = []string{"/NUL", "/con", "/COM1", "/a/LPT1", "/ /x", "/docs./x", "/docs /x", "/a.", "/a "}

func TestAttackCorpusIsLargeEnough(t *testing.T) {
	if len(attackCorpus) < 50 {
		t.Errorf("the attack corpus has %d cases, the task asks for at least 50", len(attackCorpus))
	}
}

func TestResolveRejectsAttackCorpus(t *testing.T) {
	r, _ := newTestResolver(t)
	corpus := slices.Clone(attackCorpus)
	if runtime.GOOS == "windows" {
		corpus = append(corpus, windowsOnlyCorpus...)
	}
	for _, p := range corpus {
		if rel, err := r.Resolve(FilesArea, DefaultNamespace, p); err == nil {
			t.Errorf("Resolve(%q) = %q, want an error", p, rel)
		}
	}
}

func TestResolveNormalizesToNFC(t *testing.T) {
	r, _ := newTestResolver(t)
	nfd := "/cafe\u0301/r\u00e9sume\u0301.txt" // decomposed accents, as macOS sends them
	got, err := r.Resolve(FilesArea, DefaultNamespace, nfd)
	if err != nil {
		t.Fatal(err)
	}
	if want := "caf\u00e9/r\u00e9sum\u00e9.txt"; got != want {
		t.Errorf("Resolve(NFD) = %q, want the NFC form %q", got, want)
	}
	if !norm.NFC.IsNormalString(got) {
		t.Error("the result is not NFC")
	}
}

// TestRootBlocksCorpusWithoutResolver sends every corpus path straight to
// the namespace's os.Root, as if the resolver had a bug, and checks that
// nothing appears outside the namespace (S01.6-T01 acceptance).
func TestRootBlocksCorpusWithoutResolver(t *testing.T) {
	parent := testutil.StorageRoot(t)
	l := NewLayout(filepath.Join(parent, "nas"), Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	r := NewResolver(l)
	if err := testutil.WriteFiles(l.Area(PhotosArea, DefaultNamespace), map[string]string{"private.jpg": "photo"}); err != nil {
		t.Fatal(err)
	}
	internalBefore, err := testutil.ReadFiles(l.Internal)
	if err != nil {
		t.Fatal(err)
	}
	root, err := r.OpenRoot(FilesArea, DefaultNamespace)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()

	corpus := append(slices.Clone(attackCorpus), windowsOnlyCorpus...)
	for _, p := range corpus {
		raw := strings.TrimLeft(p, "/")
		if raw == "" {
			continue
		}
		if data, err := root.ReadFile(raw); err == nil && string(data) == "photo" {
			t.Errorf("os.Root read the photos area through %q", raw)
		}
		if f, err := root.Create(raw); err == nil {
			_ = f.Close()
		}
		_ = root.MkdirAll(raw, 0o750)
	}

	// Whatever happened inside files/u0001 (Windows can create odd entries
	// such as "..." there, which is why the resolver refuses them), nothing
	// may change outside it.
	entries := func(dir string) []string {
		t.Helper()
		list, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, e := range list {
			names = append(names, e.Name())
		}
		return names
	}
	for dir, want := range map[string][]string{
		parent:                               {"nas"},
		l.Root:                               {".local-ai-nas", "files", "photos"},
		filepath.Join(l.Root, FilesArea):     {DefaultNamespace},
		filepath.Join(l.Root, PhotosArea):    {DefaultNamespace},
		l.Area(PhotosArea, DefaultNamespace): {"private.jpg"},
	} {
		if got := entries(dir); !slices.Equal(got, want) {
			t.Errorf("%s now holds %q, want %q", dir, got, want)
		}
	}
	internalAfter, err := testutil.ReadFiles(l.Internal)
	if err != nil {
		t.Fatal(err)
	}
	if len(internalAfter) != len(internalBefore) {
		t.Errorf("internal data changed: %v", internalAfter)
	}
	if data, err := os.ReadFile(filepath.Join(l.Area(PhotosArea, DefaultNamespace), "private.jpg")); err != nil || string(data) != "photo" {
		t.Error("the photos area was changed")
	}
}

// FuzzResolve checks, for any input, that an accepted path is local,
// NFC, free of ".." segments, and stable when resolved again.
func FuzzResolve(f *testing.F) {
	for _, p := range attackCorpus {
		f.Add(p)
	}
	for _, p := range []string{"/", "/a.txt", "/docs/a.txt", "/caf\u0065\u0301", "/a..b/c...", "/.hidden"} {
		f.Add(p)
	}
	r := NewResolver(NewLayout(filepath.Join(f.TempDir(), "nas"), Options{}))
	f.Fuzz(func(t *testing.T, p string) {
		rel, err := r.Resolve(FilesArea, DefaultNamespace, p)
		if err != nil {
			return
		}
		if rel != "." && !filepath.IsLocal(filepath.FromSlash(rel)) {
			t.Fatalf("Resolve(%q) = %q, which is not local", p, rel)
		}
		if slices.Contains(strings.Split(rel, "/"), "..") {
			t.Fatalf("Resolve(%q) = %q contains ..", p, rel)
		}
		if !norm.NFC.IsNormalString(rel) {
			t.Fatalf("Resolve(%q) = %q is not NFC", p, rel)
		}
		apiPath := "/" + rel
		if rel == "." {
			apiPath = "/"
		}
		again, err := r.Resolve(FilesArea, DefaultNamespace, apiPath)
		if err != nil || again != rel {
			t.Fatalf("Resolve(%q) = %q, but resolving that again gives %q, %v", p, rel, again, err)
		}
	})
}
