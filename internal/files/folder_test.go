package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

func TestCreateFolder(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{"f.txt": "x"})
	it, created, err := s.CreateFolder(t.Context(), owner, "/a", FolderOptions{})
	if err != nil || !created || it.RelPath != "a" || it.Kind != KindDir || it.OwnerID != owner {
		t.Fatalf("CreateFolder(/a) = %+v, %v, %v", it, created, err)
	}

	tests := []struct {
		name        string
		path        string
		opts        FolderOptions
		wantRel     string
		wantCreated bool
		wantErr     apperr.Kind
	}{
		{"fail on an existing folder", "/a", FolderOptions{}, "", false, apperr.Conflict},
		{"explicit fail", "/a", FolderOptions{OnConflict: ConflictFail}, "", false, apperr.Conflict},
		{"rename", "/a", FolderOptions{OnConflict: ConflictRename}, "a (1)", true, 0},
		{"rename again", "/a", FolderOptions{OnConflict: ConflictRename}, "a (2)", true, 0},
		{"overwrite keeps the folder", "/a", FolderOptions{OnConflict: ConflictOverwrite}, "a", false, 0},
		{"a file is never replaced", "/f.txt", FolderOptions{OnConflict: ConflictOverwrite}, "", false, apperr.Conflict},
		{"rename next to a file", "/f.txt", FolderOptions{OnConflict: ConflictRename}, "f.txt (1)", true, 0},
		{"unknown policy", "/b", FolderOptions{OnConflict: "merge"}, "", false, apperr.InvalidRequest},
	}
	for _, tt := range tests {
		it, created, err := s.CreateFolder(t.Context(), owner, tt.path, tt.opts)
		if tt.wantErr != 0 {
			if k := apperr.KindOf(err); err == nil || k != tt.wantErr {
				t.Errorf("%s: err = %v (kind %s), want %s", tt.name, err, k, tt.wantErr)
			}
			continue
		}
		if err != nil || it.RelPath != tt.wantRel || created != tt.wantCreated || it.Kind != KindDir {
			t.Errorf("%s: got %q created=%v err=%v; want %q created=%v", tt.name, it.RelPath, created, err, tt.wantRel, tt.wantCreated)
		}
	}
}

func TestCreateFolderParents(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"f.txt": "x"})
	if _, _, err := s.CreateFolder(t.Context(), owner, "/x/y/z", FolderOptions{}); apperr.KindOf(err) != apperr.NotFound {
		t.Errorf("missing parents without the option: %v, want not_found", err)
	}
	it, created, err := s.CreateFolder(t.Context(), owner, "/x/y/z", FolderOptions{Parents: true})
	if err != nil || !created || it.RelPath != "x/y/z" {
		t.Fatalf("with parents: %+v, %v, %v", it, created, err)
	}
	for _, d := range []string{"x", "x/y", "x/y/z"} {
		if info, err := os.Stat(filepath.Join(l.Area(storage.FilesArea, owner), filepath.FromSlash(d))); err != nil || !info.IsDir() {
			t.Errorf("%s was not created: %v", d, err)
		}
	}
	if _, _, err := s.CreateFolder(t.Context(), owner, "/f.txt/sub", FolderOptions{Parents: true}); apperr.KindOf(err) != apperr.Conflict {
		t.Errorf("a file as a parent (parents): %v, want conflict", err)
	}
	if _, _, err := s.CreateFolder(t.Context(), owner, "/f.txt/sub", FolderOptions{}); apperr.KindOf(err) != apperr.Conflict {
		t.Errorf("a file as a parent: %v, want conflict", err)
	}
}

func TestCreateFolderNameRules(t *testing.T) {
	s, l := newService(t, nil, nil)
	tests := []struct {
		path string
		opts FolderOptions
		rule string
	}{
		{"/", FolderOptions{}, storage.RuleDotName},
		{"/con", FolderOptions{}, storage.RuleReservedName},
		{"/Aux.backup", FolderOptions{}, storage.RuleReservedName},
		{"/draft.", FolderOptions{}, storage.RuleTrailingChar},
		{"/a|b", FolderOptions{}, storage.RuleForbiddenChar},
		{"/ok/bad./x", FolderOptions{Parents: true}, storage.RuleTrailingChar}, // a missing parent
	}
	for _, tt := range tests {
		_, _, err := s.CreateFolder(t.Context(), owner, tt.path, tt.opts)
		var e *apperr.Error
		if !errors.As(err, &e) || e.Kind != apperr.InvalidName || e.Rule != tt.rule {
			t.Errorf("CreateFolder(%q) = %v, want invalid_name / %s", tt.path, err, tt.rule)
		}
	}
	// Refused requests create nothing, not even the valid parent "ok" of
	// the last case: all new names are checked before any is created.
	entries, err := os.ReadDir(l.Area(storage.FilesArea, owner))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("refused requests left %d entries, first %q", len(entries), entries[0].Name())
	}

	// An existing parent with a name the rules would refuse (made outside
	// the API) can still hold new folders. Only Linux can create such a
	// name on disk.
	if runtime.GOOS != "windows" {
		if err := os.Mkdir(filepath.Join(l.Area(storage.FilesArea, owner), "odd:dir"), 0o750); err != nil {
			t.Fatal(err)
		}
		if _, _, err := s.CreateFolder(t.Context(), owner, "/odd:dir/new", FolderOptions{}); err != nil {
			t.Errorf("inside an existing odd folder: %v", err)
		}
	}
}

// TestCreateFolderRenameRace: concurrent requests for one name with
// ConflictRename all succeed, each with its own folder.
func TestCreateFolderRenameRace(t *testing.T) {
	s, _ := newService(t, nil, nil)
	const n = 10
	var wg sync.WaitGroup
	var mu sync.Mutex
	got := map[string]bool{}
	for range n {
		wg.Go(func() {
			it, created, err := s.CreateFolder(t.Context(), owner, "/r", FolderOptions{OnConflict: ConflictRename})
			if err != nil || !created {
				t.Errorf("CreateFolder: %v, created=%v", err, created)
				return
			}
			mu.Lock()
			got[it.RelPath] = true
			mu.Unlock()
		})
	}
	wg.Wait()
	want := map[string]bool{"r": true}
	for i := 1; i < n; i++ {
		want[fmt.Sprintf("r (%d)", i)] = true
	}
	if len(got) != n {
		t.Errorf("got %d distinct folders %v, want %d", len(got), got, n)
	}
	for name := range want {
		if !got[name] {
			t.Errorf("missing %q in %v", name, got)
		}
	}
}

func TestCreateFolderHook(t *testing.T) {
	h := &recordingHooks{}
	s, _ := newService(t, h, nil)
	if _, _, err := s.CreateFolder(t.Context(), owner, "/docs/new", FolderOptions{Parents: true}); err != nil {
		t.Fatal(err)
	}
	if len(h.events) != 1 || h.events[0].Op != OpCreateFolder || h.events[0].Path != "docs/new" {
		t.Errorf("events = %+v", h.events)
	}
}
