package files

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// zipContents reads an archive written to a buffer: each entry's name and
// content ("/" for folders).
func zipContents(t *testing.T, b []byte) map[string]string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, f := range r.File {
		if f.Method != zip.Store {
			t.Errorf("%s is compressed; entries are stored", f.Name)
		}
		if f.FileInfo().IsDir() {
			out[f.Name] = "/"
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		out[f.Name] = string(data)
	}
	return out
}

func TestPlanAndWriteArchive(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{
		"docs/a.txt":       "alpha",
		"docs/sub/b.txt":   "beta",
		"docs/empty/.keep": "",
		"other/docs/c.txt": "gamma",
		"top.txt":          "top",
		"été/ü.txt":        "unicode",
	})
	plan, err := s.PlanArchive(t.Context(), owner, []string{"/docs", "/other/docs", "/top.txt", "/été"})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Owner != owner || plan.Size != int64(len("alphabetagammatopunicode")) {
		t.Errorf("plan owner %q size %d", plan.Owner, plan.Size)
	}
	var buf bytes.Buffer
	if err := s.WriteArchive(t.Context(), plan, &buf); err != nil {
		t.Fatal(err)
	}
	// Two items named "docs" at the top: the second gets a number.
	want := map[string]string{
		"docs/": "/", "docs/a.txt": "alpha", "docs/empty/": "/", "docs/empty/.keep": "", "docs/sub/": "/", "docs/sub/b.txt": "beta",
		"docs (1)/": "/", "docs (1)/c.txt": "gamma",
		"top.txt": "top",
		"été/":    "/", "été/ü.txt": "unicode",
	}
	if diff := cmp.Diff(want, zipContents(t, buf.Bytes())); diff != "" {
		t.Errorf("archive (-want +got):\n%s", diff)
	}
}

func TestPlanArchiveOfTheRoot(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{"a.txt": "a"})
	plan, err := s.PlanArchive(t.Context(), owner, []string{"/"})
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range plan.Entries {
		names = append(names, e.Name)
	}
	if diff := cmp.Diff([]string{"files", "files/a.txt"}, names); diff != "" {
		t.Errorf("entries (-want +got):\n%s", diff)
	}
}

func TestPlanArchiveRefusals(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"docs/a.txt": "a"})
	for _, tc := range []struct {
		paths []string
		kind  apperr.Kind
	}{
		{nil, apperr.InvalidRequest},
		{[]string{"/missing"}, apperr.NotFound},
		{[]string{"/../u0002"}, apperr.OutsideRoot},
	} {
		if _, err := s.PlanArchive(t.Context(), owner, tc.paths); apperr.KindOf(err) != tc.kind {
			t.Errorf("PlanArchive(%v) = %v, want %s", tc.paths, err, tc.kind.Code())
		}
	}
	// A link anywhere inside is refused, never followed (B09's regression
	// test asks that links inside a large tree stay refused).
	area := l.Area(storage.FilesArea, owner)
	if !testutil.TrySymlink(t, filepath.Join(area, "docs", "a.txt"), filepath.Join(area, "docs", "link")) {
		t.Skip("links cannot be made here")
	}
	_, err := s.PlanArchive(t.Context(), owner, []string{"/docs"})
	if apperr.KindOf(err) != apperr.InvalidRequest {
		t.Errorf("a folder with a link: %v, want invalid_request", err)
	}
}

func TestPlanArchiveLimit(t *testing.T) {
	s, l := newService(t, nil, nil)
	area := l.Area(storage.FilesArea, owner)
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(filepath.Join(area, fmt.Sprintf("f%d", i)), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	defer func(old int) { archiveLimit = old }(archiveLimit)
	archiveLimit = 3
	_, err := s.PlanArchive(t.Context(), owner, []string{"/"})
	if apperr.KindOf(err) != apperr.TooLargeForSync {
		t.Errorf("over the limit: %v, want too_large_for_sync", err)
	}
}

// TestPlanArchiveLargeTreeIsFast is the regression test of bug S02-B09: the
// plan of a folder of many files reads each folder once instead of one
// Lstat per file, which took 50 s for 49,992 files on Windows.
func TestPlanArchiveLargeTreeIsFast(t *testing.T) {
	if testing.Short() {
		t.Skip("writes 20,000 files")
	}
	s, l := newService(t, nil, nil)
	dir := filepath.Join(l.Area(storage.FilesArea, owner), "big")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	const n = 20_000
	for i := 0; i < n; i++ {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("f%05d", i)), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	start := time.Now()
	plan, err := s.PlanArchive(t.Context(), owner, []string{"/big"})
	took := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Entries) != n+1 {
		t.Errorf("%d entries, want %d", len(plan.Entries), n+1)
	}
	// One Lstat per file costs about 1 ms on Windows (20 s here); reading
	// the folder once takes well under a second.
	if took > 5*time.Second {
		t.Errorf("planning %d files took %v; the bound is 5 s", n, took)
	}
	t.Logf("planned %d files in %v", n, took)
}

func TestWriteArchiveSkipsFilesRemovedSinceThePlan(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"d/a.txt": "a", "d/b.txt": "b"})
	plan, err := s.PlanArchive(t.Context(), owner, []string{"/d"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(l.Area(storage.FilesArea, owner), "d", "a.txt")); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := s.WriteArchive(t.Context(), plan, &buf); err != nil {
		t.Fatal(err)
	}
	got := zipContents(t, buf.Bytes())
	names := make([]string, 0, len(got))
	for n := range got {
		names = append(names, n)
	}
	sort.Strings(names)
	if diff := cmp.Diff([]string{"d/", "d/b.txt"}, names); diff != "" {
		t.Errorf("entries (-want +got):\n%s", diff)
	}
}

func TestWriteArchiveStopsWhenCancelled(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{"d/a.txt": "a"})
	plan, err := s.PlanArchive(t.Context(), owner, []string{"/d"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := s.WriteArchive(ctx, plan, io.Discard); err == nil {
		t.Error("a cancelled archive was written")
	}
}

func TestArchiveName(t *testing.T) {
	day := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		paths []string
		want  string
	}{
		{[]string{"/docs"}, "docs.zip"},
		{[]string{"/docs/report.pdf"}, "report.pdf.zip"},
		{[]string{"/"}, "download-2026-09-28.zip"},
		{[]string{"/a", "/b"}, "download-2026-09-28.zip"},
	} {
		if got := ArchiveName(tc.paths, day); got != tc.want {
			t.Errorf("ArchiveName(%v) = %q, want %q", tc.paths, got, tc.want)
		}
	}
}
