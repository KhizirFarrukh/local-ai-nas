package files

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// Only Go's built-in media types are asserted by extension (.pdf, .png,
// .json): Linux and Windows add their own entries for other extensions.
func TestStatDetails(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{
		"doc.pdf":   "%PDF-1.7 not really a pdf",
		"image.PNG": "not really a png", // extension case does not matter
		"data.json": `{"a":1}`,
		"noext-pdf": "%PDF-1.4\n%binary",
		"plain":     "just some words",
		"blob.zzz9": "\x00\x01\x02\x03\xff\xfe",
		"empty":     "",
		"folder/x":  "",
	})
	tests := []struct {
		path, mime string
	}{
		{"/doc.pdf", "application/pdf"},
		{"/image.PNG", "image/png"},
		{"/data.json", "application/json"},
		{"/noext-pdf", "application/pdf"},          // sniffed
		{"/plain", "text/plain; charset=utf-8"},    // sniffed
		{"/blob.zzz9", "application/octet-stream"}, // unknown extension, sniffed
		{"/empty", "text/plain; charset=utf-8"},    // sniffing an empty file
		{"/folder", ""},
	}
	etagFormat := regexp.MustCompile(`^"[0-9a-f]{24}"$`)
	for _, tt := range tests {
		it, err := s.Stat(t.Context(), owner, tt.path)
		if err != nil {
			t.Fatalf("Stat(%q): %v", tt.path, err)
		}
		if it.MIME != tt.mime {
			t.Errorf("Stat(%q).MIME = %q, want %q", tt.path, it.MIME, tt.mime)
		}
		if it.Kind == KindFile && !etagFormat.MatchString(it.ETag) {
			t.Errorf("Stat(%q).ETag = %q, want a quoted 24-digit hex tag", tt.path, it.ETag)
		}
		if it.Kind == KindDir && it.ETag != "" {
			t.Errorf("a folder got an ETag: %q", it.ETag)
		}
	}
}

func TestETagFollowsContent(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"a.txt": "version 1"})
	p := filepath.Join(l.Area(storage.FilesArea, owner), "a.txt")
	etagOf := func() string {
		t.Helper()
		it, err := s.Stat(t.Context(), owner, "/a.txt")
		if err != nil {
			t.Fatal(err)
		}
		return it.ETag
	}
	first := etagOf()
	if again := etagOf(); again != first {
		t.Fatalf("the ETag is not stable: %q, then %q", first, again)
	}

	// Rewritten in place: the content and the modification time change.
	if err := os.WriteFile(p, []byte("version 2"), 0o600); err != nil {
		t.Fatal(err)
	}
	later := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(p, later, later); err != nil {
		t.Fatal(err)
	}
	second := etagOf()
	if second == first {
		t.Error("the ETag did not change after a rewrite")
	}

	// Replaced by another file of the same size and time (an atomic
	// rename): only the file ID differs, and the ETag must still change.
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, []byte("version 3"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(tmp, later, later); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmp, p); err != nil {
		t.Fatal(err)
	}
	if third := etagOf(); third == second {
		t.Error("the ETag did not change when the file was replaced with the same size and time")
	}
}

func TestListShowsMIMEByExtensionOnly(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{"doc.pdf": "x", "noext": "%PDF-1.4"})
	page, err := s.List(t.Context(), owner, "/", ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Item{}
	for _, it := range page.Items {
		got[it.Name] = it
	}
	if got["doc.pdf"].MIME != "application/pdf" || got["noext"].MIME != "" {
		t.Errorf("listing MIME types: doc.pdf %q, noext %q; want application/pdf and none (no sniffing in lists)", got["doc.pdf"].MIME, got["noext"].MIME)
	}
	if got["doc.pdf"].ETag != "" {
		t.Error("listings do not carry ETags")
	}
}
