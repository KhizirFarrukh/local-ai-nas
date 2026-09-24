package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/logging"
)

// loadSpec loads and validates api/openapi.yaml.
func loadSpec(t *testing.T) *openapi3.T {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "api", "openapi.yaml"))
	if err != nil {
		t.Fatalf("load the spec: %v", err)
	}
	if err := doc.Validate(t.Context()); err != nil {
		t.Fatalf("the spec is not valid OpenAPI: %v", err)
	}
	return doc
}

// validateProblem checks a response against the spec's Problem schema and
// the rules the schema cannot express (status and correlation ID).
func validateProblem(t *testing.T, doc *openapi3.T, label string, rec *httptest.ResponseRecorder) {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != apperr.ContentType {
		t.Errorf("%s: Content-Type %q, want %q", label, ct, apperr.ContentType)
	}
	var body any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("%s: body is not JSON: %v", label, err)
	}
	if err := doc.Components.Schemas["Problem"].Value.VisitJSON(body); err != nil {
		t.Errorf("%s: the body does not match the Problem schema: %v\n%s", label, err, rec.Body)
	}
	m, _ := body.(map[string]any)
	if status, _ := m["status"].(float64); int(status) != rec.Code {
		t.Errorf("%s: problem status %v, response status %d", label, m["status"], rec.Code)
	}
	if m["correlation_id"] != rec.Header().Get(logging.RequestIDHeader) {
		t.Errorf("%s: correlation_id %v, X-Request-ID %q", label, m["correlation_id"], rec.Header().Get(logging.RequestIDHeader))
	}
}

// errorCases are requests that must produce an error, with the expected
// status and code. Every route in the route table must be covered by at
// least one case (TestErrorResponsesMatchSchema checks this), so a new
// endpoint cannot skip the contract test.
var errorCases = []struct {
	method, path string
	status       int
	code         string
}{
	// No route at all.
	{http.MethodGet, "/", http.StatusNotFound, "not_found"},
	{http.MethodGet, "/api/v1/nothing-here", http.StatusNotFound, "not_found"},
	{http.MethodGet, "/api/v2/system/health", http.StatusNotFound, "not_found"},
	// GET /api/v1/system/health: other methods.
	{http.MethodPost, "/api/v1/system/health", http.StatusMethodNotAllowed, "method_not_allowed"},
	{http.MethodDelete, "/api/v1/system/health", http.StatusMethodNotAllowed, "method_not_allowed"},
	// GET /api/v1/files/items.
	{http.MethodGet, "/api/v1/files/items", http.StatusBadRequest, "invalid_request"},
	{http.MethodGet, "/api/v1/files/items?path=/&limit=0", http.StatusBadRequest, "invalid_request"},
	{http.MethodGet, "/api/v1/files/items?path=/&limit=abc", http.StatusBadRequest, "invalid_request"},
	{http.MethodGet, "/api/v1/files/items?path=/&sort=color", http.StatusBadRequest, "invalid_request"},
	{http.MethodGet, "/api/v1/files/items?path=/&cursor=garbage", http.StatusBadRequest, "invalid_request"},
	{http.MethodGet, "/api/v1/files/items?path=/docs/../../x", http.StatusBadRequest, "outside_root"},
	{http.MethodGet, "/api/v1/files/items?path=/docs%5Ca.txt", http.StatusBadRequest, "invalid_name"},
	{http.MethodGet, "/api/v1/files/items?path=/missing", http.StatusNotFound, "not_found"},
	{http.MethodPost, "/api/v1/files/items?path=/", http.StatusMethodNotAllowed, "method_not_allowed"},
	// POST /api/v1/files/folders: other methods (the body cases are below).
	{http.MethodGet, "/api/v1/files/folders", http.StatusMethodNotAllowed, "method_not_allowed"},
	// POST /api/v1/files/operations/{rename,move}: other methods (the body
	// cases are below).
	{http.MethodGet, "/api/v1/files/operations/rename", http.StatusMethodNotAllowed, "method_not_allowed"},
	{http.MethodPut, "/api/v1/files/operations/move", http.StatusMethodNotAllowed, "method_not_allowed"},
	{http.MethodDelete, "/api/v1/files/operations/copy", http.StatusMethodNotAllowed, "method_not_allowed"},
	// GET /api/v1/files/content (412 and 416 need request headers; the
	// download tests validate them against the schema).
	{http.MethodGet, "/api/v1/files/content", http.StatusBadRequest, "invalid_request"},
	{http.MethodGet, "/api/v1/files/content?path=/docs", http.StatusBadRequest, "invalid_request"},
	{http.MethodGet, "/api/v1/files/content?path=/docs/../../x", http.StatusBadRequest, "outside_root"},
	{http.MethodGet, "/api/v1/files/content?path=/missing.txt", http.StatusNotFound, "not_found"},
	// PUT /api/v1/files/content: other methods (the body cases are below).
	{http.MethodPost, "/api/v1/files/content", http.StatusMethodNotAllowed, "method_not_allowed"},
	// The reserved photos routes, with any method.
	{http.MethodGet, "/api/v1/photos", http.StatusNotImplemented, "not_available"},
	{http.MethodPost, "/api/v1/photos", http.StatusNotImplemented, "not_available"},
	{http.MethodGet, "/api/v1/photos/timeline", http.StatusNotImplemented, "not_available"},
	{http.MethodPut, "/api/v1/photos/items/1", http.StatusNotImplemented, "not_available"},
	{http.MethodDelete, "/api/v1/photos/items/1", http.StatusNotImplemented, "not_available"},
}

// bodyErrorCases are requests with a body that must produce an error. An
// empty contentType sends no Content-Type; a non-zero length replaces the
// declared Content-Length (-1: unknown, as with chunked encoding). They
// count for route coverage like errorCases.
var bodyErrorCases = []struct {
	method, path, contentType, body string
	length                          int64
	status                          int
	code                            string
}{
	// POST /api/v1/files/folders.
	{http.MethodPost, "/api/v1/files/folders", "", "", 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/folders", jsonType, `{"path":"/new","extra":1}`, 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/folders", jsonType, `{"path":"/new"}{}`, 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/folders", jsonType, `{"path":"/new","on_conflict":"merge"}`, 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/folders", jsonType, `{"path":"/con"}`, 0, http.StatusBadRequest, "invalid_name"},
	{http.MethodPost, "/api/v1/files/folders", jsonType, `{"path":"/../x"}`, 0, http.StatusBadRequest, "outside_root"},
	{http.MethodPost, "/api/v1/files/folders", jsonType, `{"path":"/x/y/z"}`, 0, http.StatusNotFound, "not_found"},
	{http.MethodPost, "/api/v1/files/folders", jsonType, `{"path":"/docs"}`, 0, http.StatusConflict, "conflict"},
	{http.MethodPost, "/api/v1/files/folders", jsonType, `{"path":"/` + strings.Repeat("a", 2<<20) + `"}`, 0, http.StatusRequestEntityTooLarge, "too_large"},
	// POST /api/v1/files/operations/rename.
	{http.MethodPost, "/api/v1/files/operations/rename", "", "", 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/operations/rename", jsonType, `{"path":"/docs/a.txt","new_name":"x","extra":1}`, 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/operations/rename", jsonType, `{"path":"/docs/a.txt","new_name":"x","on_conflict":"merge"}`, 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/operations/rename", jsonType, `{"path":"/","new_name":"x"}`, 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/operations/rename", jsonType, `{"path":"/docs/a.txt","new_name":"a/b"}`, 0, http.StatusBadRequest, "invalid_name"},
	{http.MethodPost, "/api/v1/files/operations/rename", jsonType, `{"path":"/docs/a.txt","new_name":"con"}`, 0, http.StatusBadRequest, "invalid_name"},
	{http.MethodPost, "/api/v1/files/operations/rename", jsonType, `{"path":"/docs/../../a","new_name":"x"}`, 0, http.StatusBadRequest, "outside_root"},
	{http.MethodPost, "/api/v1/files/operations/rename", jsonType, `{"path":"/missing","new_name":"x"}`, 0, http.StatusNotFound, "not_found"},
	{http.MethodPost, "/api/v1/files/operations/rename", jsonType, `{"path":"/docs/a.txt","new_name":"b.txt"}`, 0, http.StatusConflict, "conflict"},
	// POST /api/v1/files/operations/move.
	{http.MethodPost, "/api/v1/files/operations/move", jsonType, `{"from":"/docs","to":"/docs/sub"}`, 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/operations/move", jsonType, `{"from":"/docs/a.txt","to":"/"}`, 0, http.StatusBadRequest, "invalid_name"},
	{http.MethodPost, "/api/v1/files/operations/move", jsonType, `{"from":"/docs/a.txt","to":"/../x"}`, 0, http.StatusBadRequest, "outside_root"},
	{http.MethodPost, "/api/v1/files/operations/move", jsonType, `{"from":"/missing","to":"/x"}`, 0, http.StatusNotFound, "not_found"},
	{http.MethodPost, "/api/v1/files/operations/move", jsonType, `{"from":"/docs/a.txt","to":"/nope/a.txt"}`, 0, http.StatusNotFound, "not_found"},
	{http.MethodPost, "/api/v1/files/operations/move", jsonType, `{"from":"/docs/a.txt","to":"/readme.md"}`, 0, http.StatusConflict, "conflict"},
	{http.MethodPost, "/api/v1/files/operations/move", jsonType, `{"from":"/docs/a.txt","to":"/readme.md/a.txt"}`, 0, http.StatusConflict, "conflict"},
	{http.MethodPost, "/api/v1/files/operations/move", jsonType, `{"from":"/docs","to":"/empty","on_conflict":"overwrite"}`, 0, http.StatusConflict, "conflict"},
	// POST /api/v1/files/operations/copy (422 too_large_for_sync is in
	// TestCopyEndpoint, which needs a server with small limits).
	{http.MethodPost, "/api/v1/files/operations/copy", "", "", 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/operations/copy", jsonType, `{"from":"/docs","to":"/x","extra":1}`, 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/operations/copy", jsonType, `{"from":"/docs","to":"/x","on_conflict":"merge"}`, 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/operations/copy", jsonType, `{"from":"/docs","to":"/docs/in"}`, 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPost, "/api/v1/files/operations/copy", jsonType, `{"from":"/docs","to":"/"}`, 0, http.StatusBadRequest, "invalid_name"},
	{http.MethodPost, "/api/v1/files/operations/copy", jsonType, `{"from":"/docs","to":"/../x"}`, 0, http.StatusBadRequest, "outside_root"},
	{http.MethodPost, "/api/v1/files/operations/copy", jsonType, `{"from":"/missing","to":"/x"}`, 0, http.StatusNotFound, "not_found"},
	{http.MethodPost, "/api/v1/files/operations/copy", jsonType, `{"from":"/docs","to":"/empty"}`, 0, http.StatusConflict, "conflict"},
	// PUT /api/v1/files/content.
	{http.MethodPut, "/api/v1/files/content", octetType, "abc", 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPut, "/api/v1/files/content?path=/n.txt&on_conflict=merge", octetType, "abc", 0, http.StatusBadRequest, "invalid_request"},
	{http.MethodPut, "/api/v1/files/content?path=/n.txt", octetType, "abc", 10, http.StatusBadRequest, "invalid_request"}, // shorter than declared
	{http.MethodPut, "/api/v1/files/content?path=/n.txt", octetType, "abc", 1, http.StatusBadRequest, "invalid_request"},  // longer than declared
	{http.MethodPut, "/api/v1/files/content?path=/", octetType, "abc", 0, http.StatusBadRequest, "invalid_name"},
	{http.MethodPut, "/api/v1/files/content?path=/aux.txt", octetType, "abc", 0, http.StatusBadRequest, "invalid_name"},
	{http.MethodPut, "/api/v1/files/content?path=/.local-ai-nas-tmp-x.part", octetType, "abc", 0, http.StatusBadRequest, "invalid_name"},
	{http.MethodPut, "/api/v1/files/content?path=/docs/../../x", octetType, "abc", 0, http.StatusBadRequest, "outside_root"},
	{http.MethodPut, "/api/v1/files/content?path=/missing/n.txt", octetType, "abc", 0, http.StatusNotFound, "not_found"},
	{http.MethodPut, "/api/v1/files/content?path=/docs/a.txt", octetType, "abc", 0, http.StatusConflict, "conflict"},
	{http.MethodPut, "/api/v1/files/content?path=/docs&on_conflict=overwrite", octetType, "abc", 0, http.StatusConflict, "conflict"},
	{http.MethodPut, "/api/v1/files/content?path=/readme.md/n.txt", octetType, "abc", 0, http.StatusConflict, "conflict"},
	{http.MethodPut, "/api/v1/files/content?path=/n.txt", "", "abc", -1, http.StatusLengthRequired, "length_required"},
	{http.MethodPut, "/api/v1/files/content?path=/n.txt", octetType, "abc", DefaultMaxUploadBytes + 1, http.StatusRequestEntityTooLarge, "too_large"},
}

const (
	jsonType  = "application/json"
	octetType = "application/octet-stream"
)

// TestErrorResponsesMatchSchema is the S01.5-T03 contract test: every
// error response the API produces, across all endpoints, validates against
// the Problem schema in the spec.
func TestErrorResponsesMatchSchema(t *testing.T) {
	doc := loadSpec(t)
	h := New(Options{Files: testFiles(t)})

	// A mux with the same routes tells which route each case exercises.
	routeOf := http.NewServeMux()
	for _, r := range Routes(Options{}) {
		routeOf.Handle(r.Pattern, r.Handler)
	}
	covered := map[string]bool{}

	for _, c := range errorCases {
		label := c.method + " " + c.path
		req := httptest.NewRequest(c.method, c.path, nil)
		if _, pattern := routeOf.Handler(req); pattern != "" {
			covered[pattern] = true
		} else if _, p := routeOf.Handler(httptest.NewRequest(http.MethodGet, c.path, nil)); p != "" {
			covered[p] = true // a wrong-method case covers the route of its path
		}

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.status {
			t.Errorf("%s: status %d, want %d (%s)", label, rec.Code, c.status, rec.Body)
			continue
		}
		validateProblem(t, doc, label, rec)
		var p apperr.Problem
		_ = json.Unmarshal(rec.Body.Bytes(), &p)
		if p.Code != c.code {
			t.Errorf("%s: code %q, want %q", label, p.Code, c.code)
		}
		if c.status == http.StatusMethodNotAllowed && rec.Header().Get("Allow") == "" {
			t.Errorf("%s: 405 without an Allow header", label)
		}
	}

	for _, c := range bodyErrorCases {
		label := c.method + " " + c.path + " " + c.body
		if len(label) > 120 {
			label = label[:120] + "…"
		}
		req := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
		if c.contentType != "" {
			req.Header.Set("Content-Type", c.contentType)
		}
		if c.length != 0 {
			req.ContentLength = c.length
		}
		if _, pattern := routeOf.Handler(req); pattern != "" {
			covered[pattern] = true
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.status {
			t.Errorf("%s: status %d, want %d (%s)", label, rec.Code, c.status, rec.Body)
			continue
		}
		validateProblem(t, doc, label, rec)
		var p apperr.Problem
		_ = json.Unmarshal(rec.Body.Bytes(), &p)
		if p.Code != c.code {
			t.Errorf("%s: code %q, want %q", label, p.Code, c.code)
		}
	}

	for _, r := range Routes(Options{}) {
		if !covered[r.Pattern] {
			t.Errorf("route %q has no error case in errorCases or bodyErrorCases", r.Pattern)
		}
	}
}

// TestPanicResponseMatchesSchema covers the 500 of the recovery middleware.
func TestPanicResponseMatchesSchema(t *testing.T) {
	doc := loadSpec(t)
	h := logging.RequestID(apperr.Recover(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/x", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
	validateProblem(t, doc, "panic", rec)
}

// TestEveryCodeIsInTheSpec keeps the spec's list of codes and the codes of
// internal/apperr the same.
func TestEveryCodeIsInTheSpec(t *testing.T) {
	doc := loadSpec(t)
	var specCodes, codes []string
	for _, v := range doc.Components.Schemas["Problem"].Value.Properties["code"].Value.Enum {
		specCodes = append(specCodes, v.(string))
	}
	for _, k := range apperr.Kinds() {
		codes = append(codes, k.Code())
	}
	slices.Sort(specCodes)
	slices.Sort(codes)
	if !slices.Equal(specCodes, codes) {
		t.Errorf("the spec lists codes %v, internal/apperr defines %v", specCodes, codes)
	}
}

// TestHealthMatchesSchema checks the health body against HealthReport.
func TestHealthMatchesSchema(t *testing.T) {
	doc := loadSpec(t)
	rec := httptest.NewRecorder()
	New(Options{Version: "dev"}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/system/health", nil))
	var body any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if err := doc.Components.Schemas["HealthReport"].Value.VisitJSON(body); err != nil {
		t.Errorf("health body does not match HealthReport: %v\n%s", err, rec.Body)
	}
}
