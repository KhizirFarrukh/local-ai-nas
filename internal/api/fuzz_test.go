package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/health"
)

// FuzzAPI sends arbitrary methods, targets (path and query), and bodies to
// the whole API (S01.5-T04). Whatever the input:
//   - the server never answers 5xx, apart from the documented 501 of the
//     reserved photos routes;
//   - every error response is a problem.
//
// The seeds cover every route in the route table, so new endpoints are
// fuzzed as soon as they are added. Whether invalid input can reach the
// service layer is checked per endpoint with fake services (S01.3).
func FuzzAPI(f *testing.F) {
	for _, r := range Routes(Options{}) {
		method, path, found := strings.Cut(r.Pattern, " ")
		if !found {
			method, path = http.MethodGet, r.Pattern
		}
		f.Add(method, path, "")
		f.Add(method, path+"?path=/a&limit=-1&cursor=%zz&sort=nope", "{}")
		f.Add(http.MethodPost, path, `{"from":"/..","to":`)
	}
	for _, seed := range [][3]string{
		{"GET", "/", ""},
		{"PATCH", "/api/v1/system/health?x=%00", "\x00"},
		{"DELETE", "/api/v1/../etc/passwd", ""},
		{"GET", "/api/v1/system/health/..", ""},
		{"OPTIONS", "/api/v1/photos/%2e%2e/", ""},
	} {
		f.Add(seed[0], seed[1], seed[2])
	}

	ok := health.Check{Name: "config", Run: func(context.Context) (string, error) { return "valid", nil }}
	h := New(Options{Version: "fuzz", Checks: []health.Check{ok}, Files: testFiles(f)})

	f.Fuzz(func(t *testing.T, method, target, body string) {
		if method == "" || strings.ContainsAny(method, " \t\r\n/") || !strings.HasPrefix(target, "/") {
			return // not an HTTP request line a client can send
		}
		req, err := http.NewRequest(method, "http://nas.local"+target, strings.NewReader(body))
		if err != nil {
			return
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		photos := req.URL.Path == "/api/v1/photos" || strings.HasPrefix(req.URL.Path, "/api/v1/photos/")
		if rec.Code >= 500 && (rec.Code != http.StatusNotImplemented || !photos) {
			t.Fatalf("%s %s → %d: %s", method, target, rec.Code, rec.Body)
		}
		if rec.Code >= 400 && rec.Header().Get("Content-Type") != apperr.ContentType {
			t.Fatalf("%s %s → %d with Content-Type %q, want a problem", method, target, rec.Code, rec.Header().Get("Content-Type"))
		}
	})
}
