package files

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

func TestDownload(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{"docs/a.txt": "0123456789", "docs/blob": "%PDF-1.7 ..."})
	for _, tt := range []struct{ path, content, mime string }{
		{"/docs/a.txt", "0123456789", "text/plain; charset=utf-8"},
		{"/docs/blob", "%PDF-1.7 ...", "application/pdf"}, // sniffed
	} {
		it, body, err := s.Download(t.Context(), owner, tt.path)
		if err != nil {
			t.Fatalf("%s: %v", tt.path, err)
		}
		stat, err := s.Stat(t.Context(), owner, tt.path)
		if err != nil {
			t.Fatal(err)
		}
		if it.Size != int64(len(tt.content)) || it.MIME != tt.mime || it.ETag == "" || it.ETag != stat.ETag {
			t.Errorf("%s: item %+v, want size %d, MIME %q, and Stat's ETag %s", tt.path, it, len(tt.content), tt.mime, stat.ETag)
		}
		got, err := io.ReadAll(body) // the sniff did not move the read position
		if err != nil || string(got) != tt.content {
			t.Errorf("%s: read %q, %v", tt.path, got, err)
		}
		if _, err := body.Seek(4, io.SeekStart); err != nil {
			t.Fatal(err)
		}
		if rest, _ := io.ReadAll(body); string(rest) != tt.content[4:] {
			t.Errorf("%s: after a seek read %q", tt.path, rest)
		}
		if err := body.Close(); err != nil {
			t.Error(err)
		}
	}
}

// TestDownloadServesTheVersionItDescribes: a file replaced during a
// download does not change what the open download sends, and the next
// download has a new ETag.
func TestDownloadServesTheVersionItDescribes(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{"a.txt": "old version"})
	it, body, err := s.Download(t.Context(), owner, "/a.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = body.Close() }()
	if _, _, err := upload(s, "/a.txt", "the new version", ConflictOverwrite); err != nil {
		if runtime.GOOS == "windows" {
			// Windows cannot replace a file that is open without delete
			// sharing; the overwrite fails instead of mixing versions.
			t.Skipf("overwrite of an open file on Windows: %v", err)
		}
		t.Fatal(err)
	}
	if got, _ := io.ReadAll(body); string(got) != "old version" {
		t.Errorf("the open download read %q, want the old version", got)
	}
	next, body2, err := s.Download(t.Context(), owner, "/a.txt")
	if err != nil {
		t.Fatal(err)
	}
	_ = body2.Close()
	if next.ETag == it.ETag {
		t.Error("the replaced file has the same ETag")
	}
}

func TestDownloadRefuses(t *testing.T) {
	s, l := newService(t, nil, map[string]string{
		"docs/a.txt":                            "x",
		"docs/" + storage.TempPrefix + "1.part": "partial",
	})
	tests := []struct {
		path string
		want apperr.Kind
	}{
		{"/docs", apperr.InvalidRequest},
		{"/", apperr.InvalidRequest},
		{"/docs/missing.txt", apperr.NotFound},
		{"/docs/a.txt/x", apperr.NotFound},
		{"/docs/" + storage.TempPrefix + "1.part", apperr.NotFound},
		{"/docs/../../x", apperr.OutsideRoot},
	}
	if runtime.GOOS != "windows" { // symbolic links need privileges on Windows
		area := l.Area(storage.FilesArea, owner)
		if err := os.Symlink("a.txt", filepath.Join(area, "docs", "link.txt")); err != nil {
			t.Fatal(err)
		}
		tests = append(tests, struct {
			path string
			want apperr.Kind
		}{"/docs/link.txt", apperr.InvalidRequest})
	}
	for _, tt := range tests {
		_, body, err := s.Download(t.Context(), owner, tt.path)
		if apperr.KindOf(err) != tt.want || err == nil || body != nil {
			t.Errorf("%s: %v (body %v), want %v", tt.path, err, body, tt.want)
		}
	}
}

func TestDownloadHook(t *testing.T) {
	h := &recordingHooks{}
	s, _ := newService(t, h, map[string]string{"a.txt": "x"})
	_, body, err := s.Download(t.Context(), owner, "/a.txt")
	if err != nil {
		t.Fatal(err)
	}
	_ = body.Close()
	if len(h.events) != 1 || h.events[0].Op != OpDownload || h.events[0].Path != "a.txt" {
		t.Errorf("events = %+v", h.events)
	}
}
