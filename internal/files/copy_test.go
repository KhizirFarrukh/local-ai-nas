package files

import (
	"context"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// modTimes returns the modification time of every file and folder below
// dir on disk, keyed by the slash path relative to dir.
func modTimes(t *testing.T, dir string) map[string]time.Time {
	t.Helper()
	times := map[string]time.Time{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || p == dir {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		times[filepath.ToSlash(rel)] = info.ModTime()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return times
}

// TestCopyTreeIsByteIdentical: a copied tree has the same files with the
// same bytes and modification times, and the source is unchanged.
func TestCopyTreeIsByteIdentical(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	tree := map[string]string{"src/empty.txt": "", "src/emptydir/.keep": ""}
	for i, size := range []int{1, 100, copyBufferSize + 3, 2*copyBufferSize - 1} {
		b := make([]byte, size)
		for j := range b {
			b[j] = byte(rng.Uint32())
		}
		tree[fmt.Sprintf("src/d%d/sub/f%d.bin", i%2, i)] = string(b)
	}
	s, l := newService(t, nil, tree)
	area := l.Area(storage.FilesArea, owner)
	// Old, distinct times, so keeping them is visible.
	old := time.Date(2020, 5, 17, 10, 30, 0, 0, time.UTC)
	i := 0
	for p := range modTimes(t, filepath.Join(area, "src")) {
		i++
		if err := os.Chtimes(filepath.Join(area, "src", filepath.FromSlash(p)), time.Time{}, old.Add(time.Duration(i)*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	// Set the folder times after the files, which would change them.
	for _, d := range []string{"src/d0/sub", "src/d1/sub", "src/d0", "src/d1", "src/emptydir"} {
		if err := os.Chtimes(filepath.Join(area, filepath.FromSlash(d)), time.Time{}, old); err != nil {
			t.Fatal(err)
		}
	}
	srcTimes := modTimes(t, filepath.Join(area, "src"))

	it, created, err := s.Copy(t.Context(), owner, "/src", "/dst", CopyOptions{})
	if err != nil || !created || it.RelPath != "dst" || it.Kind != KindDir {
		t.Fatalf("Copy = %+v, %v, %v", it, created, err)
	}
	disk := onDisk(t, l)
	for name, content := range tree {
		copied := "dst" + strings.TrimPrefix(name, "src")
		if disk[copied] != content || disk[name] != content {
			t.Errorf("%s: the copy or the source differs (%d, %d bytes, want %d)", name, len(disk[copied]), len(disk[name]), len(content))
		}
	}
	if len(disk) != 2*len(tree) {
		t.Errorf("%d files on disk, want %d (no temporary files)", len(disk), 2*len(tree))
	}
	if diff := cmp.Diff(srcTimes, modTimes(t, filepath.Join(area, "dst"))); diff != "" {
		t.Errorf("modification times of the copy (-source +copy):\n%s", diff)
	}
}

func TestCopyConflictPolicies(t *testing.T) {
	s, l := newService(t, nil, map[string]string{
		"a.txt": "a", "b.txt": "b", "dir/x": "x", "other/.keep": "",
	})
	tests := []struct {
		from, to    string
		policy      OnConflict
		wantErr     apperr.Kind
		want        string
		wantCreated bool
	}{
		{"/a.txt", "/c.txt", "", 0, "c.txt", true},
		{"/a.txt", "/b.txt", "", apperr.Conflict, "", false},
		{"/a.txt", "/b.txt", ConflictRename, 0, "b (1).txt", true},
		{"/a.txt", "/b.txt", ConflictOverwrite, 0, "b.txt", false},
		{"/a.txt", "/other", ConflictOverwrite, apperr.Conflict, "", false},
		{"/dir", "/other", "", apperr.Conflict, "", false},
		{"/dir", "/other", ConflictOverwrite, apperr.Conflict, "", false},
		{"/dir", "/a.txt", ConflictOverwrite, apperr.Conflict, "", false},
		{"/dir", "/other", ConflictRename, 0, "other (1)", true},
		{"/dir", "/other/dir", "", 0, "other/dir", true},
		{"/a.txt", "/a.txt", ConflictRename, 0, "a (1).txt", true}, // a copy next to the original
	}
	for _, tt := range tests {
		it, created, err := s.Copy(t.Context(), owner, tt.from, tt.to, CopyOptions{OnConflict: tt.policy})
		if tt.wantErr != 0 {
			if err == nil || apperr.KindOf(err) != tt.wantErr {
				t.Errorf("%s → %s %q: %v, want %v", tt.from, tt.to, tt.policy, err, tt.wantErr)
			}
			continue
		}
		if err != nil || it.RelPath != tt.want || created != tt.wantCreated {
			t.Errorf("%s → %s %q: %s created=%v %v; want %s created=%v", tt.from, tt.to, tt.policy, it.RelPath, created, err, tt.want, tt.wantCreated)
		}
	}
	want := map[string]string{
		"a.txt": "a", "b.txt": "a", "c.txt": "a", "b (1).txt": "a", "a (1).txt": "a",
		"dir/x": "x", "other/.keep": "", "other/dir/x": "x", "other (1)/x": "x",
	}
	if diff := cmp.Diff(want, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk (-want +got):\n%s", diff)
	}
}

// TestCopyRefusalsLeaveNothing: refused copies write nothing, not even a
// temporary file, and the checks run before any byte is copied.
func TestCopyRefusalsLeaveNothing(t *testing.T) {
	fixture := map[string]string{
		"src/a.txt": "12345", "src/sub/b.txt": "678", "src/" + storage.TempPrefix + "1.part": "partial", "f.txt": "f",
	}
	lowSpace := func(string) (uint64, error) { return 1 << 20, nil }
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name     string
		from, to string
		limits   CopyLimits
		space    storage.FreeFunc
		ctx      context.Context
		want     apperr.Kind
	}{
		{"into itself", "/src", "/src/sub/copy", CopyLimits{}, nil, nil, apperr.InvalidRequest},
		{"the root", "/", "/src/root", CopyLimits{}, nil, nil, apperr.InvalidRequest},
		{"onto the root", "/src", "/", CopyLimits{}, nil, nil, apperr.InvalidName},
		{"missing source", "/nope", "/x", CopyLimits{}, nil, nil, apperr.NotFound},
		{"missing parent", "/src", "/nope/x", CopyLimits{}, nil, nil, apperr.NotFound},
		{"file as parent", "/src", "/f.txt/x", CopyLimits{}, nil, nil, apperr.Conflict},
		{"reserved target", "/src", "/aux", CopyLimits{}, nil, nil, apperr.InvalidName},
		{"temporary source", "/src/" + storage.TempPrefix + "1.part", "/x", CopyLimits{}, nil, nil, apperr.NotFound},
		{"too many items", "/src", "/dst", CopyLimits{MaxItems: 3}, nil, nil, apperr.TooLargeForSync},
		{"too many bytes", "/src", "/dst", CopyLimits{MaxBytes: 7}, nil, nil, apperr.TooLargeForSync},
		{"no free space", "/src", "/dst", CopyLimits{}, lowSpace, nil, apperr.InsufficientStorage},
		{"canceled folder", "/src", "/dst", CopyLimits{}, nil, canceled, apperr.InvalidRequest},
		{"canceled file", "/f.txt", "/g.txt", CopyLimits{}, nil, canceled, apperr.InvalidRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := storage.NewLayout(testutil.StorageRoot(t), storage.Options{})
			if _, err := l.Init(); err != nil {
				t.Fatal(err)
			}
			if err := testutil.WriteFiles(l.Area(storage.FilesArea, owner), fixture); err != nil {
				t.Fatal(err)
			}
			o := Options{CopyLimits: tt.limits}
			if tt.space != nil {
				o.Space = storage.NewSpaceGuard(l.Root, 1<<30, tt.space)
			}
			s := NewLocal(storage.NewResolver(l), o)
			ctx := tt.ctx
			if ctx == nil {
				ctx = t.Context()
			}
			if _, _, err := s.Copy(ctx, owner, tt.from, tt.to, CopyOptions{}); err == nil || apperr.KindOf(err) != tt.want {
				t.Errorf("error %v, want %v", err, tt.want)
			}
			if diff := cmp.Diff(fixture, onDisk(t, l)); diff != "" {
				t.Errorf("files on disk changed (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCopyLimitsAreInclusive(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{"src/a.txt": "12345", "src/b.txt": "678"})
	s.copyLimits = CopyLimits{MaxItems: 3, MaxBytes: 8} // the folder and two files; 8 bytes
	if _, _, err := s.Copy(t.Context(), owner, "/src", "/dst", CopyOptions{}); err != nil {
		t.Errorf("a copy at the limits = %v", err)
	}
}

// TestCopyTreeRules: names in the tree follow the name rules, links are
// refused, and temporary files are not copied.
func TestCopyTreeRules(t *testing.T) {
	s, l := newService(t, nil, map[string]string{
		"src/a.txt": "a", "src/" + storage.TempPrefix + "1.part": "partial",
	})
	if _, _, err := s.Copy(t.Context(), owner, "/src", "/dst", CopyOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, ok := onDisk(t, l)["dst/"+storage.TempPrefix+"1.part"]; ok {
		t.Error("a temporary file was copied")
	}
	if runtime.GOOS == "windows" {
		t.Skip("the rest needs names and links that Windows cannot create")
	}
	area := l.Area(storage.FilesArea, owner)
	for name, want := range map[string]apperr.Kind{"bad:name.txt": apperr.InvalidName, "link": apperr.InvalidRequest} {
		dir := filepath.Join(area, "t-"+strings.TrimSuffix(name, ".txt"))
		if err := os.Mkdir(dir, 0o750); err != nil {
			t.Fatal(err)
		}
		if name == "link" {
			err := os.Symlink("../src/a.txt", filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
		} else if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := s.Copy(t.Context(), owner, "/"+filepath.Base(dir), "/copy-"+filepath.Base(dir), CopyOptions{}); err == nil || apperr.KindOf(err) != want {
			t.Errorf("copy of a folder with %q = %v, want %v", name, err, want)
		}
		if _, err := os.Lstat(filepath.Join(area, "copy-"+filepath.Base(dir))); !os.IsNotExist(err) {
			t.Errorf("a refused copy left %s: %v", "copy-"+filepath.Base(dir), err)
		}
	}
}

// TestCopyConcurrentRename: concurrent copies of a folder to one name with
// rename give complete, distinct copies.
func TestCopyConcurrentRename(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"src/a.txt": "a", "src/sub/b.txt": "b"})
	const n = 6
	paths := make([]string, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			var it Item
			it, _, errs[i] = s.Copy(t.Context(), owner, "/src", "/dst", CopyOptions{OnConflict: ConflictRename})
			paths[i] = it.RelPath
		})
	}
	wg.Wait()
	disk := onDisk(t, l)
	seen := map[string]bool{}
	for i := range n {
		if errs[i] != nil {
			t.Fatalf("copy %d: %v", i, errs[i])
		}
		if seen[paths[i]] || disk[paths[i]+"/a.txt"] != "a" || disk[paths[i]+"/sub/b.txt"] != "b" {
			t.Errorf("copy %d at %s is a duplicate or incomplete", i, paths[i])
		}
		seen[paths[i]] = true
	}
	if len(disk) != 2*(n+1) {
		t.Errorf("%d files on disk, want %d", len(disk), 2*(n+1))
	}
}

func TestCopyHook(t *testing.T) {
	h := &recordingHooks{}
	s, _ := newService(t, h, map[string]string{"a.txt": "a"})
	if _, _, err := s.Copy(t.Context(), owner, "/a.txt", "/b.txt", CopyOptions{}); err != nil {
		t.Fatal(err)
	}
	want := []Event{{Op: OpCopy, Owner: owner, Path: "a.txt", Target: "b.txt"}}
	if diff := cmp.Diff(want, h.events); diff != "" {
		t.Errorf("events (-want +got):\n%s", diff)
	}
}
