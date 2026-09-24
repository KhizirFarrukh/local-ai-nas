package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// TestCleanTemp: old temporary files and folders go, new ones and items
// stay.
func TestCleanTemp(t *testing.T) {
	tmp := storage.TempPrefix
	s, l := newService(t, nil, map[string]string{
		"a.txt":                     "a",
		"docs/" + tmp + "old.part":  "old partial upload",
		"docs/" + tmp + "new.part":  "upload in progress",
		tmp + "tree.part/sub/x.txt": "old partial copy",
		"docs/keep.txt":             "k",
	})
	area := l.Area(storage.FilesArea, owner)
	now := time.Now()
	old := now.Add(-48 * time.Hour)
	for _, p := range []string{"docs/" + tmp + "old.part", tmp + "tree.part/sub/x.txt", tmp + "tree.part/sub", tmp + "tree.part"} {
		if err := os.Chtimes(filepath.Join(area, filepath.FromSlash(p)), time.Time{}, old); err != nil {
			t.Fatal(err)
		}
	}
	n, err := s.CleanTemp(t.Context(), owner, 24*time.Hour, now)
	if err != nil || n != 2 {
		t.Errorf("CleanTemp removed %d (%v), want 2", n, err)
	}
	want := map[string]string{"a.txt": "a", "docs/keep.txt": "k", "docs/" + tmp + "new.part": "upload in progress"}
	if diff := cmp.Diff(want, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk (-want +got):\n%s", diff)
	}
	if _, err := os.Stat(filepath.Join(area, tmp+"tree.part")); !os.IsNotExist(err) {
		t.Errorf("the old temporary folder is still there: %v", err)
	}
}
