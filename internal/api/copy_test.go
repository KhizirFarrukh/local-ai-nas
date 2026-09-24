package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

func TestCopyEndpoint(t *testing.T) {
	doc := loadSpec(t)
	h := New(Options{Files: testFiles(t)})
	tests := []struct {
		body     string
		wantCode int
		wantPath string
	}{
		{`{"from":"/docs","to":"/docs-copy"}`, http.StatusCreated, "/docs-copy"},
		{`{"from":"/docs/a.txt","to":"/docs/a.txt","on_conflict":"rename"}`, http.StatusCreated, "/docs/a (1).txt"},
		{`{"from":"/readme.md","to":"/docs/b.txt","on_conflict":"overwrite"}`, http.StatusOK, "/docs/b.txt"},
	}
	for _, tt := range tests {
		rec := postJSON(h, "/api/v1/files/operations/copy", tt.body)
		var it gen.FileItem
		if err := json.Unmarshal(rec.Body.Bytes(), &it); err != nil || rec.Code != tt.wantCode || it.Path != tt.wantPath {
			t.Errorf("%s → %d %s, want %d %s", tt.body, rec.Code, rec.Body, tt.wantCode, tt.wantPath)
			continue
		}
		var body any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if err := doc.Components.Schemas["FileItem"].Value.VisitJSON(body); err != nil {
			t.Errorf("%s: the response does not match FileItem: %v", tt.body, err)
		}
	}
	_, copied := getItems(t, h, url.Values{"path": {"/docs-copy"}})
	if got := strings.Join(names(copied.Items), ","); got != "a.txt,b.txt" {
		t.Errorf("/docs-copy = %s", got)
	}
	if rec := getContent(h, http.MethodGet, "/docs/b.txt", nil); rec.Body.String() != fixtureFiles["readme.md"] {
		t.Errorf("/docs/b.txt after the overwrite = %q", rec.Body)
	}
}

// TestCopyTooLargeForSync: a copy over the synchronous limits is a
// too_large_for_sync problem.
func TestCopyTooLargeForSync(t *testing.T) {
	doc := loadSpec(t)
	l := storage.NewLayout(testutil.StorageRoot(t), storage.Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	if err := testutil.WriteFiles(l.Area(storage.FilesArea, storage.DefaultNamespace), fixtureFiles); err != nil {
		t.Fatal(err)
	}
	svc := files.NewLocal(storage.NewResolver(l), files.Options{CopyLimits: files.CopyLimits{MaxItems: 2}})
	rec := postJSON(New(Options{Files: svc}), "/api/v1/files/operations/copy", `{"from":"/docs","to":"/docs-copy"}`)
	var p apperr.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil || rec.Code != http.StatusUnprocessableEntity || p.Code != "too_large_for_sync" {
		t.Fatalf("→ %d %s, want 422 too_large_for_sync", rec.Code, rec.Body)
	}
	validateProblem(t, doc, "too_large_for_sync", rec)
}

func TestCopyInvalidInputNeverReachesService(t *testing.T) {
	svc := &recordingFiles{}
	h := New(Options{Files: svc})
	for _, body := range []string{``, `[]`, `{"from":"/a","to":"/b","x":1}`, `{"from":"/a","to":"/b","on_conflict":"skip"}`, `{"from":1}`} {
		rec := postJSON(h, "/api/v1/files/operations/copy", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%q → %d %s, want 400", body, rec.Code, rec.Body)
		}
	}
	if svc.calls != 0 {
		t.Errorf("the service was called %d times with invalid input", svc.calls)
	}
}
