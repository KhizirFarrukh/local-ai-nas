package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/logging"
)

// routePath returns the path part of a ServeMux pattern
// ("[METHOD ][HOST]/PATH").
func routePath(pattern string) string {
	p := pattern
	if _, rest, ok := strings.Cut(p, " "); ok {
		p = rest
	}
	if i := strings.Index(p, "/"); i > 0 {
		p = p[i:] // drop a host part
	}
	return p
}

// TestRoutesAreVersioned is the S01.5-T02 check: every registered route is
// under /api/v1 or /api/docs.
func TestRoutesAreVersioned(t *testing.T) {
	routes := Routes(Options{})
	if len(routes) == 0 {
		t.Fatal("no routes")
	}
	for _, r := range routes {
		p := routePath(r.Pattern)
		ok := p == "/api/v1" || strings.HasPrefix(p, "/api/v1/") || p == "/api/docs" || strings.HasPrefix(p, "/api/docs/")
		if !ok {
			t.Errorf("route %q is outside /api/v1 and /api/docs", r.Pattern)
		}
	}
}

func TestRoutePath(t *testing.T) {
	for in, want := range map[string]string{
		"GET /api/v1/system/health": "/api/v1/system/health",
		"/api/v1/photos/":           "/api/v1/photos/",
		"POST example.com/api/v1/x": "/api/v1/x",
		"example.com/x":             "/x",
	} {
		if got := routePath(in); got != want {
			t.Errorf("routePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUnversionedPathsAreNotServed(t *testing.T) {
	h := New(Options{})
	for _, p := range []string{"/", "/health", "/api", "/api/v2/system/health", "/v1/system/health"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", p, rec.Code)
		}
	}
}

func TestHealthRoute(t *testing.T) {
	rec := httptest.NewRecorder()
	New(Options{Version: "1.2.3"}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/system/health", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"version":"1.2.3"`) {
		t.Errorf("health = %d %s", rec.Code, rec.Body)
	}
	if rec.Header().Get(logging.RequestIDHeader) == "" {
		t.Error("no request ID header")
	}
}

// TestPhotosAPINotAvailable: the photos area exists on disk from S01, but
// its API is reserved until S04 (S01.2-T06).
func TestPhotosAPINotAvailable(t *testing.T) {
	h := New(Options{})
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/photos"},
		{http.MethodGet, "/api/v1/photos/"},
		{http.MethodGet, "/api/v1/photos/timeline"},
		{http.MethodPost, "/api/v1/photos/items"},
		{http.MethodDelete, "/api/v1/photos/items/abc"},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		var p apperr.Problem
		if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
			t.Fatalf("%s %s: body %q is not a problem: %v", tc.method, tc.path, rec.Body, err)
		}
		if rec.Code != http.StatusNotImplemented || p.Code != "not_available" || rec.Header().Get("Content-Type") != apperr.ContentType {
			t.Errorf("%s %s = %d %q (%s), want 501 not_available", tc.method, tc.path, rec.Code, p.Code, rec.Header().Get("Content-Type"))
		}
		if p.CorrelationID == "" || p.CorrelationID != rec.Header().Get(logging.RequestIDHeader) {
			t.Errorf("%s %s: correlation ID %q, request ID %q", tc.method, tc.path, p.CorrelationID, rec.Header().Get(logging.RequestIDHeader))
		}
	}
}
