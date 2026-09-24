package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// postJSON posts a JSON body to path.
func postJSON(h http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRenameMoveEndpoints(t *testing.T) {
	doc := loadSpec(t)
	h := New(Options{Files: testFiles(t)})
	tests := []struct {
		path, body, wantPath string
	}{
		{"/api/v1/files/operations/rename", `{"path":"/docs/a.txt","new_name":"renamed.txt"}`, "/docs/renamed.txt"},
		{"/api/v1/files/operations/rename", `{"path":"/docs/renamed.txt","new_name":"b.txt","on_conflict":"rename"}`, "/docs/b (1).txt"},
		{"/api/v1/files/operations/move", `{"from":"/docs/b (1).txt","to":"/moved.txt"}`, "/moved.txt"},
		{"/api/v1/files/operations/move", `{"from":"/moved.txt","to":"/readme.md","on_conflict":"overwrite"}`, "/readme.md"},
		{"/api/v1/files/operations/move", `{"from":"/empty","to":"/docs/empty"}`, "/docs/empty"},
	}
	for _, tt := range tests {
		rec := postJSON(h, tt.path, tt.body)
		var it gen.FileItem
		if err := json.Unmarshal(rec.Body.Bytes(), &it); err != nil || rec.Code != http.StatusOK || it.Path != tt.wantPath {
			t.Errorf("%s → %d %s, want 200 %s", tt.body, rec.Code, rec.Body, tt.wantPath)
			continue
		}
		var body any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if err := doc.Components.Schemas["FileItem"].Value.VisitJSON(body); err != nil {
			t.Errorf("%s: the response does not match FileItem: %v", tt.body, err)
		}
	}
	_, root := getItems(t, h, url.Values{"path": {"/"}})
	_, docs := getItems(t, h, url.Values{"path": {"/docs"}})
	if got := strings.Join(names(root.Items), ","); got != "docs,readme.md" {
		t.Errorf("root after the moves = %s", got)
	}
	if got := strings.Join(names(docs.Items), ","); got != "b.txt,empty" {
		t.Errorf("/docs after the moves = %s", got)
	}
}

func TestOperationsInvalidInputNeverReachesService(t *testing.T) {
	svc := &recordingFiles{}
	h := New(Options{Files: svc})
	for _, path := range []string{"/api/v1/files/operations/rename", "/api/v1/files/operations/move"} {
		for _, body := range []string{
			``,
			`not json`,
			`{"path":"/x","new_name":"y","from":"/a","to":"/b","extra":true}`,
			`{"path":"/x","new_name":"y","on_conflict":"merge"}`,
			`{"from":"/a","to":"/b","on_conflict":"merge"}`,
			`{"from":"/a","to":7}`,
			`{} {}`,
		} {
			rec := postJSON(h, path, body)
			var p apperr.Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil || rec.Code != http.StatusBadRequest || p.Code != "invalid_request" {
				t.Errorf("%s %q → %d %s, want 400 invalid_request", path, body, rec.Code, rec.Body)
			}
		}
	}
	if svc.calls != 0 {
		t.Errorf("the service was called %d times with invalid input", svc.calls)
	}
}
