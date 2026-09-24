package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// postFolder calls POST /api/v1/files/folders with a JSON body.
func postFolder(t *testing.T, h http.Handler, body string) (*httptest.ResponseRecorder, gen.FileItem) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/folders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var it gen.FileItem
	if rec.Code == http.StatusCreated || rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &it); err != nil {
			t.Fatalf("response is not a FileItem: %v\n%s", err, rec.Body)
		}
	}
	return rec, it
}

func TestCreateFolderEndpoint(t *testing.T) {
	doc := loadSpec(t)
	h := New(Options{Files: testFiles(t)})
	tests := []struct {
		body     string
		wantCode int
		wantPath string
	}{
		{`{"path":"/projects"}`, http.StatusCreated, "/projects"},
		{`{"path":"/projects","on_conflict":"rename"}`, http.StatusCreated, "/projects (1)"},
		{`{"path":"/projects","on_conflict":"overwrite"}`, http.StatusOK, "/projects"},
		{`{"path":"/a/b/c","parents":true}`, http.StatusCreated, "/a/b/c"},
		{`{"path":"/docs/new","parents":false,"on_conflict":"fail"}`, http.StatusCreated, "/docs/new"},
	}
	for _, tt := range tests {
		rec, it := postFolder(t, h, tt.body)
		if rec.Code != tt.wantCode || it.Path != tt.wantPath || it.Kind != "dir" {
			t.Errorf("%s → %d %+v, want %d %s", tt.body, rec.Code, it, tt.wantCode, tt.wantPath)
			continue
		}
		var body any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if err := doc.Components.Schemas["FileItem"].Value.VisitJSON(body); err != nil {
			t.Errorf("%s: the response does not match FileItem: %v", tt.body, err)
		}
	}
	// The new folder shows up in its parent's listing.
	_, list := getItems(t, h, map[string][]string{"path": {"/a/b"}})
	if names(list.Items)[0] != "c" {
		t.Errorf("listing of /a/b = %v", names(list.Items))
	}
}

func TestCreateFolderInvalidInputNeverReachesService(t *testing.T) {
	svc := &recordingFiles{}
	h := New(Options{Files: svc})
	for _, body := range []string{
		``,
		`not json`,
		`{"path":"/x","extra":true}`,
		`{"path":"/x","parents":"yes"}`,
		`{"path":"/x","on_conflict":"merge"}`,
		`{"path":"/x"} trailing`,
		`[]`,
	} {
		rec, _ := postFolder(t, h, body)
		var p apperr.Problem
		if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil || rec.Code != http.StatusBadRequest || p.Code != "invalid_request" {
			t.Errorf("%q → %d %s, want 400 invalid_request", body, rec.Code, rec.Body)
		}
	}
	// Without the JSON content type the body is not even read.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/folders", strings.NewReader(`{"path":"/x"}`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("text/plain body → %d, want 400", rec.Code)
	}
	if svc.calls != 0 {
		t.Errorf("the service was called %d times with invalid input", svc.calls)
	}
}
