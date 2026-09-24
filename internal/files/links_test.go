package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// linkTree returns a service whose namespace holds docs/a.txt and links:
// in-dir → docs and in-file → docs/a.txt (inside the area), out-dir → a
// folder outside the storage root holding secret.txt, and out-rel → the
// photos area by a relative target. It also returns the outside folder.
func linkTree(t *testing.T) (*Local, storage.Layout, string) {
	t.Helper()
	s, l := newService(t, nil, map[string]string{"docs/a.txt": "a", "docs/sub/b.txt": "b"})
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	area := l.Area(storage.FilesArea, owner)
	testutil.Symlink(t, filepath.Join(area, "docs"), filepath.Join(area, "in-dir"))
	testutil.Symlink(t, filepath.Join(area, "docs", "a.txt"), filepath.Join(area, "in-file"))
	testutil.Symlink(t, outside, filepath.Join(area, "out-dir"))
	testutil.Symlink(t, filepath.FromSlash("../../photos/u0001"), filepath.Join(area, "out-rel"))
	return s, l, outside
}

// TestLinksAreNeverFollowed is the S01.6-T03 acceptance test: no operation
// reads, writes, or traverses through a symbolic link, whether it points
// inside the area or outside.
func TestLinksAreNeverFollowed(t *testing.T) {
	s, l, outside := linkTree(t)
	ctx := t.Context()
	through := []string{
		"/in-dir/a.txt", "/in-dir/sub/b.txt", "/in-dir/new.txt", "/in-dir/sub",
		"/out-dir/secret.txt", "/out-dir/new.txt", "/out-rel/x.jpg",
	}
	ops := map[string]func(p string) error{
		"stat":     func(p string) error { _, err := s.Stat(ctx, owner, p); return err },
		"list":     func(p string) error { _, err := s.List(ctx, owner, p, ListOptions{}); return err },
		"download": func(p string) error { _, _, err := s.Download(ctx, owner, p); return err },
		"upload":   func(p string) error { _, _, err := upload(s, p, "x", ConflictOverwrite); return err },
		"create folder": func(p string) error {
			_, _, err := s.CreateFolder(ctx, owner, p, FolderOptions{Parents: true})
			return err
		},
		"rename":    func(p string) error { _, err := s.Rename(ctx, owner, p, "renamed", MoveOptions{}); return err },
		"move from": func(p string) error { _, err := s.Move(ctx, owner, p, "/moved", MoveOptions{}); return err },
		"move to": func(p string) error {
			_, err := s.Move(ctx, owner, "/docs/sub/b.txt", p, MoveOptions{OnConflict: ConflictOverwrite})
			return err
		},
		"copy from": func(p string) error { _, _, err := s.Copy(ctx, owner, p, "/copied", CopyOptions{}); return err },
		"copy to": func(p string) error {
			_, _, err := s.Copy(ctx, owner, "/docs/a.txt", p, CopyOptions{OnConflict: ConflictOverwrite})
			return err
		},
		"delete": func(p string) error { return s.Delete(ctx, owner, p, DeleteOptions{Recursive: true}) },
	}
	for name, op := range ops {
		for _, p := range through {
			if err := op(p); apperr.KindOf(err) != apperr.InvalidRequest || !strings.Contains(err.Error(), "symbolic link") {
				t.Errorf("%s %s = %v, want invalid_request about the link", name, p, err)
			}
		}
	}
	checkUntouched(t, l, outside)
}

// TestLinkItems: a link is an item of its own. It is listed without a
// target, is never read or written through, and can be renamed and
// deleted without touching what it points to.
func TestLinkItems(t *testing.T) {
	s, l, outside := linkTree(t)
	ctx := t.Context()
	page, err := s.List(ctx, owner, "/", ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range page.Items {
		if strings.HasPrefix(it.Name, "in-") || strings.HasPrefix(it.Name, "out-") {
			if it.Kind != KindSymlink || it.Size != 0 || it.MIME != "" || it.ETag != "" {
				t.Errorf("listed link %+v, want kind symlink and nothing about its target", it)
			}
		}
	}
	refusals := []struct {
		name string
		err  error
		want apperr.Kind
	}{
		{"download", func() error { _, _, err := s.Download(ctx, owner, "/in-file"); return err }(), apperr.InvalidRequest},
		{"list", func() error { _, err := s.List(ctx, owner, "/in-dir", ListOptions{}); return err }(), apperr.InvalidRequest},
		{"copy", func() error { _, _, err := s.Copy(ctx, owner, "/out-dir", "/c", CopyOptions{}); return err }(), apperr.InvalidRequest},
		{"overwrite", func() error { _, _, err := upload(s, "/in-file", "x", ConflictOverwrite); return err }(), apperr.Conflict},
		{"folder", func() error {
			_, _, err := s.CreateFolder(ctx, owner, "/in-dir", FolderOptions{OnConflict: ConflictOverwrite})
			return err
		}(), apperr.Conflict},
	}
	for _, r := range refusals {
		if apperr.KindOf(r.err) != r.want || r.err == nil {
			t.Errorf("%s of a link = %v, want %v", r.name, r.err, r.want)
		}
	}
	if it, err := s.Rename(ctx, owner, "/in-file", "renamed-link", MoveOptions{}); err != nil || it.Kind != KindSymlink {
		t.Errorf("rename of a link = %+v, %v", it, err)
	}
	for _, p := range []string{"/renamed-link", "/in-dir", "/out-dir", "/out-rel"} {
		if err := s.Delete(ctx, owner, p, DeleteOptions{Recursive: true}); err != nil {
			t.Errorf("delete %s: %v", p, err)
		}
	}
	checkUntouched(t, l, outside)
}

// checkUntouched: what the links point to is unchanged.
func checkUntouched(t *testing.T, l storage.Layout, outside string) {
	t.Helper()
	if b, err := os.ReadFile(filepath.Join(outside, "secret.txt")); err != nil || string(b) != "secret" {
		t.Errorf("the file outside the storage root changed: %q, %v", b, err)
	}
	if entries, _ := os.ReadDir(outside); len(entries) != 1 {
		t.Errorf("the folder outside the storage root has %d entries, want 1", len(entries))
	}
	if entries, _ := os.ReadDir(l.Area(storage.PhotosArea, owner)); len(entries) != 0 {
		t.Errorf("the photos area has %d entries, want 0", len(entries))
	}
	disk := onDisk(t, l)
	if disk["docs/a.txt"] != "a" || disk["docs/sub/b.txt"] != "b" {
		t.Errorf("the link targets inside the area changed: %v", disk)
	}
}
