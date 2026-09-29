package files

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// birthKnown reports whether the file system of dir records creation
// times, as BirthTime reads them.
func birthKnown(t *testing.T, dir, name string) bool {
	t.Helper()
	d, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = d.Close() }()
	info, err := os.Lstat(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	_, ok := storage.BirthTime(d, name, info)
	return ok
}

// TestAddedTimeKeptByRenameAndEdit is the S02.3-T05 check of FR-215: the
// added date of an upload is when it arrived, and a rename or an edit
// keeps it, while the modification time follows the edit. Where the file
// system records no creation time, the added date is the modification
// time.
func TestAddedTimeKeptByRenameAndEdit(t *testing.T) {
	s, l := newService(t, nil, nil)
	area := l.Area(storage.FilesArea, owner)
	before := time.Now().Add(-2 * time.Second) // file system clocks are coarse
	up, _, err := s.Upload(t.Context(), owner, "/a.txt", strings.NewReader("hello"), 5, UploadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !birthKnown(t, area, "a.txt") {
		if !up.AddedTime.Equal(up.ModTime) {
			t.Errorf("without creation times, added %v, want the modification time %v", up.AddedTime, up.ModTime)
		}
		t.Skip("this file system records no creation times; the rest needs them")
	}
	if up.AddedTime.Before(before) || up.AddedTime.After(time.Now().Add(2*time.Second)) {
		t.Errorf("added %v, want about now", up.AddedTime)
	}
	// An edit from outside, a day later, and a rename.
	later := time.Now().Add(24 * time.Hour)
	if err := os.WriteFile(filepath.Join(area, "a.txt"), []byte("hello, again"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(area, "a.txt"), later, later); err != nil {
		t.Fatal(err)
	}
	moved, err := s.Rename(t.Context(), owner, "/a.txt", "b.txt", MoveOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Stat(t.Context(), owner, "/b.txt")
	if err != nil {
		t.Fatal(err)
	}
	page, err := s.List(t.Context(), owner, "/", ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for what, it := range map[string]Item{"rename result": moved, "details": got, "listing": page.Items[0]} {
		if !it.AddedTime.Equal(up.AddedTime) {
			t.Errorf("%s: added %v, want %v (kept)", what, it.AddedTime, up.AddedTime)
		}
		if it.ModTime.Before(later.Add(-time.Second)) {
			t.Errorf("%s: modified %v, want the edit's %v", what, it.ModTime, later)
		}
	}
}

// TestListSortsByAdded sorts by the added date, pages through it with a
// cursor, and keeps the name as the tie-break.
func TestListSortsByAdded(t *testing.T) {
	s, _ := newService(t, nil, nil)
	for _, name := range []string{"c.txt", "a.txt", "b.txt"} { // the order they arrive
		if _, _, err := s.Upload(t.Context(), owner, "/"+name, strings.NewReader(name), int64(len(name)), UploadOptions{}); err != nil {
			t.Fatal(err)
		}
		time.Sleep(30 * time.Millisecond) // distinct times on coarse clocks
	}
	for _, tt := range []struct {
		order Order
		want  []string
	}{
		{Asc, []string{"c.txt", "a.txt", "b.txt"}},
		{Desc, []string{"b.txt", "a.txt", "c.txt"}},
	} {
		got, pages := listAll(t, s, "/", ListOptions{Sort: SortAdded, Order: tt.order, Limit: 1})
		if diff := cmp.Diff(tt.want, got); diff != "" || pages != 3 {
			t.Errorf("by added %s (-want +got):\n%s pages=%d", tt.order, diff, pages)
		}
	}
	if _, err := s.List(t.Context(), owner, "/", ListOptions{Sort: "added"}); apperr.KindOf(err) != apperr.InvalidRequest {
		t.Errorf("unknown sort key: %v", err)
	}
}

// TestListLocate reports where a named item is in the sort order
// (S02.4-T05), without changing the page.
func TestListLocate(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{"a.txt": "1", "b.txt": "333", "c.txt": "22"})
	for _, tt := range []struct {
		opts ListOptions
		want int
	}{
		{ListOptions{Locate: "c.txt"}, 2},
		{ListOptions{Locate: "c.txt", Sort: SortSize}, 1},
		{ListOptions{Locate: "c.txt", Order: Desc, Limit: 1}, 0},
		{ListOptions{Locate: "missing.txt"}, -1},
		{ListOptions{}, -1},
	} {
		page, err := s.List(t.Context(), owner, "/", tt.opts)
		if err != nil {
			t.Fatal(err)
		}
		if page.Position != tt.want {
			t.Errorf("List(%+v).Position = %d, want %d", tt.opts, page.Position, tt.want)
		}
		if tt.opts.Limit == 1 && (len(page.Items) != 1 || page.Items[0].Name != "c.txt") {
			t.Errorf("locate changed the page: %+v", page.Items)
		}
	}
	if _, err := s.List(t.Context(), owner, "/", ListOptions{Locate: "a/b"}); apperr.KindOf(err) != apperr.InvalidRequest {
		t.Errorf("locate with a slash: %v", err)
	}
}

// TestUsage adds up a folder tree (S02.3-T05, FR-214): files at any depth,
// folders counted, links neither followed nor counted, temporary files
// of writes in progress left out.
func TestUsage(t *testing.T) {
	s, l := newService(t, nil, map[string]string{
		"top/a.txt":             "12345",
		"top/sub/b.txt":         "123",
		"top/sub/deeper/c.txt":  "1",
		"top/sub/deeper/d.txt":  "",
		"top/empty/.keep-empty": "", // replaced below by an empty folder
		"other/x.txt":           "not counted",
	})
	area := l.Area(storage.FilesArea, owner)
	if err := os.Remove(filepath.Join(area, "top", "empty", ".keep-empty")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(area, "top", storage.TempPrefix+"upload"), []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}
	u, err := s.Usage(t.Context(), owner, "/top")
	if err != nil {
		t.Fatal(err)
	}
	if u.Size != 9 || u.Files != 4 || u.Folders != 3 || u.Item.RelPath != "top" {
		t.Errorf("usage of /top = %+v, want 9 bytes, 4 files, 3 folders", u)
	}
	if u, err := s.Usage(t.Context(), owner, "/top/a.txt"); err != nil || u.Size != 5 || u.Files != 1 || u.Folders != 0 {
		t.Errorf("usage of a file = %+v, %v; want its own size", u, err)
	}
	if u, err := s.Usage(t.Context(), owner, "/"); err != nil || u.Size != 20 || u.Files != 5 {
		t.Errorf("usage of the root = %+v, %v; want 20 bytes in 5 files", u, err)
	}
	if _, err := s.Usage(t.Context(), owner, "/missing"); apperr.KindOf(err) != apperr.NotFound {
		t.Errorf("missing folder: %v", err)
	}
	if _, err := s.Usage(t.Context(), owner, "/../x"); apperr.KindOf(err) != apperr.OutsideRoot {
		t.Errorf("a path outside the area: %v", err)
	}
}

// TestUsageSkipsLinks: a link to a large file is neither followed nor
// counted.
func TestUsageSkipsLinks(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"big/data.bin": strings.Repeat("x", 1000), "top/a.txt": "1"})
	area := l.Area(storage.FilesArea, owner)
	if !testutil.TrySymlink(t, filepath.Join("..", "big"), filepath.Join(area, "top", "link")) {
		t.Skip("cannot create links here")
	}
	u, err := s.Usage(t.Context(), owner, "/top")
	if err != nil || u.Size != 1 || u.Files != 1 || u.Folders != 0 {
		t.Errorf("usage with a link = %+v, %v; want only a.txt", u, err)
	}
}

// TestUsageStopsWhenCancelled: a client that goes away stops the walk.
func TestUsageStopsWhenCancelled(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{"top/a.txt": "1"})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := s.Usage(ctx, owner, "/top"); err == nil {
		t.Error("a cancelled walk returned no error")
	}
}
