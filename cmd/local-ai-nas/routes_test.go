package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// TestPhotosAPINotAvailable: the photos area exists on disk from S01, but
// its API is reserved until S04 (S01.2-T06).
func TestPhotosAPINotAvailable(t *testing.T) {
	h := newHandler(slog.New(slog.DiscardHandler), nil)
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
		if p.CorrelationID == "" || p.CorrelationID != rec.Header().Get("X-Request-ID") {
			t.Errorf("%s %s: correlation ID %q, request ID %q", tc.method, tc.path, p.CorrelationID, rec.Header().Get("X-Request-ID"))
		}
	}
}
