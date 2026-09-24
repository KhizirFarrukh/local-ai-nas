package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

func TestDelete(t *testing.T) {
	s, l := newService(t, nil, map[string]string{
		"a.txt": "a", "full/x.txt": "x", "full/sub/y.txt": "y", "keep/k.txt": "k",
	})
	area := l.Area(storage.FilesArea, owner)
	if err := os.Mkdir(filepath.Join(area, "empty"), 0o750); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		path      string
		recursive bool
		want      apperr.Kind // 0: deleted
	}{
		{"/a.txt", false, 0},
		{"/empty", false, 0},
		{"/full", false, apperr.Conflict}, // not empty: refused, nothing deleted
		{"/full", true, 0},
		{"/", true, apperr.InvalidRequest},
		{"/missing", false, apperr.NotFound},
		{"/keep/k.txt/x", false, apperr.NotFound},
		{"/keep/../../x", false, apperr.OutsideRoot},
	}
	for _, tt := range tests {
		err := s.Delete(t.Context(), owner, tt.path, DeleteOptions{Recursive: tt.recursive})
		if tt.want == 0 && err != nil || tt.want != 0 && (err == nil || apperr.KindOf(err) != tt.want) {
			t.Errorf("Delete(%s, recursive=%v) = %v, want %v", tt.path, tt.recursive, err, tt.want)
		}
		if tt.path == "/full" && !tt.recursive {
			if !strings.Contains(err.Error(), "set recursive") {
				t.Errorf("the refusal does not say how to delete the folder: %v", err)
			}
			if _, ok := onDisk(t, l)["full/sub/y.txt"]; !ok {
				t.Error("a refused delete removed files")
			}
		}
	}
	if diff := cmp.Diff(map[string]string{"keep/k.txt": "k"}, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk (-want +got):\n%s", diff)
	}
	for _, gone := range []string{"a.txt", "empty", "full"} {
		if _, err := os.Lstat(filepath.Join(area, gone)); !os.IsNotExist(err) {
			t.Errorf("%s still exists: %v", gone, err)
		}
	}
}

// TestDeleteFolderWithUnfinishedWrite: a folder that holds only a
// temporary file looks empty, so the refusal says why.
func TestDeleteFolderWithUnfinishedWrite(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"up/" + storage.TempPrefix + "1.part": "partial"})
	err := s.Delete(t.Context(), owner, "/up", DeleteOptions{})
	if apperr.KindOf(err) != apperr.Conflict || !strings.Contains(err.Error(), "unfinished") {
		t.Errorf("Delete = %v, want a conflict about an unfinished upload", err)
	}
	if err := s.Delete(t.Context(), owner, "/up/"+storage.TempPrefix+"1.part", DeleteOptions{}); apperr.KindOf(err) != apperr.NotFound {
		t.Errorf("deleting the temporary file = %v, want not_found", err)
	}
	if err := s.Delete(t.Context(), owner, "/up", DeleteOptions{Recursive: true}); err != nil {
		t.Fatal(err)
	}
	if n := len(onDisk(t, l)); n != 0 {
		t.Errorf("%d files left", n)
	}
}

// TestDeleteLinkNotTarget: deleting a link, or a folder holding one,
// never deletes what the link points to.
func TestDeleteLinkNotTarget(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"target/t.txt": "t", "links/.keep": ""})
	area := l.Area(storage.FilesArea, owner)
	for _, link := range []string{"links/to-dir", "direct"} {
		testutil.Symlink(t, filepath.Join(area, "target"), filepath.Join(area, filepath.FromSlash(link)))
	}
	if err := s.Delete(t.Context(), owner, "/direct", DeleteOptions{Recursive: true}); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(t.Context(), owner, "/links", DeleteOptions{Recursive: true}); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(map[string]string{"target/t.txt": "t"}, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk (-want +got):\n%s", diff)
	}
}

func TestDeleteHook(t *testing.T) {
	h := &recordingHooks{}
	s, l := newService(t, h, map[string]string{"a.txt": "a", "b.txt": "b"})
	if err := s.Delete(t.Context(), owner, "/a.txt", DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	h.vetoErr = apperr.New(apperr.Conflict, "vetoed")
	if err := s.Delete(t.Context(), owner, "/b.txt", DeleteOptions{}); apperr.KindOf(err) != apperr.Conflict {
		t.Errorf("vetoed delete = %v", err)
	}
	want := []Event{{Op: OpDelete, Owner: owner, Path: "a.txt"}, {Op: OpDelete, Owner: owner, Path: "b.txt"}}
	if diff := cmp.Diff(want, h.events); diff != "" {
		t.Errorf("events (-want +got):\n%s", diff)
	}
	if _, ok := onDisk(t, l)["b.txt"]; !ok {
		t.Error("a vetoed delete deleted the file")
	}
}
