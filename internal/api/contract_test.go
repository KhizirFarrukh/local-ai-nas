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
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
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
	// DELETE /api/v1/files/items.
	{http.MethodDelete, "/api/v1/files/items", http.StatusBadRequest, "invalid_request"},
	{http.MethodDelete, "/api/v1/files/items?path=/docs&recursive=maybe", http.StatusBadRequest, "invalid_request"},
	{http.MethodDelete, "/api/v1/files/items?path=/", http.StatusBadRequest, "invalid_request"},
	{http.MethodDelete, "/api/v1/files/items?path=/docs/../../x", http.StatusBadRequest, "outside_root"},
	{http.MethodDelete, "/api/v1/files/items?path=/missing", http.StatusNotFound, "not_found"},
	{http.MethodDelete, "/api/v1/files/items?path=/docs", http.StatusConflict, "conflict"},
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
	// The tus endpoint without a tus server: 501 (the tus protocol and its
	// problem bodies are tested in internal/uploads).
	{http.MethodPost, "/api/v1/files/uploads/", http.StatusNotImplemented, "not_available"},
	{http.MethodPost, "/api/v1/files/uploads", http.StatusNotImplemented, "not_available"},
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

// TestErrorStatusesAreInTheSpec is part of S01.5-T05 (the spec is
// complete for S01): every status that a contract case gets is declared
// in the spec for its operation; a 405 for the path. A new error that the
// spec does not document fails here.
func TestErrorStatusesAreInTheSpec(t *testing.T) {
	doc := loadSpec(t)
	routeOf := http.NewServeMux()
	for _, r := range Routes(Options{}) {
		routeOf.Handle(r.Pattern, r.Handler)
	}
	check := func(method, target string, status int) {
		t.Helper()
		req := httptest.NewRequest(method, target, nil)
		_, pattern := routeOf.Handler(req)
		if pattern == "" {
			_, pattern = routeOf.Handler(httptest.NewRequest(http.MethodGet, target, nil))
		}
		if pattern == "" {
			return // no endpoint at all: the 404 of an unknown path
		}
		if !specDeclares(doc, pattern, method, status) {
			t.Errorf("%s %s answers %d, which the spec does not declare for %q", method, target, status, pattern)
		}
	}
	for _, c := range errorCases {
		check(c.method, c.path, c.status)
	}
	for _, c := range bodyErrorCases {
		check(c.method, c.path, c.status)
	}
}

// specDeclares reports whether the spec declares status for the route
// pattern: for its method, or for any operation of the path when the
// pattern has no method (hand-written catch-alls) or the status is 405.
func specDeclares(doc *openapi3.T, pattern, method string, status int) bool {
	m, p, found := strings.Cut(pattern, " ")
	if !found {
		m, p = "", pattern
	}
	p = strings.TrimPrefix(p, "/api/v1")
	item := doc.Paths.Find(p)
	if item == nil {
		item = doc.Paths.Find(strings.TrimSuffix(p, "/"))
	}
	if item == nil {
		item = doc.Paths.Find(p + "/")
	}
	if item == nil {
		return false
	}
	for opMethod, op := range item.Operations() {
		if (m == "" || status == http.StatusMethodNotAllowed || strings.EqualFold(opMethod, method)) && op.Responses.Status(status) != nil {
			return true
		}
	}
	return false
}

// TestTusStatusesAreInTheSpec runs tus requests against a real tus server
// and checks that the spec's uploads operations declare every status they
// get, and that each error is a problem.
func TestTusStatusesAreInTheSpec(t *testing.T) {
	doc := loadSpec(t)
	svc, _ := testFilesDir(t)
	c := client{t, testutil.NewServer(t, New(Options{Files: svc, Uploads: testUploads(t, svc)})).URL}
	tus := map[string]string{"Tus-Resumable": "1.0.0"}
	with := func(extra map[string]string) map[string]string {
		h := map[string]string{"Tus-Resumable": "1.0.0"}
		for k, v := range extra {
			h[k] = v
		}
		return h
	}
	st, hdr, b := c.do("POST", UploadsPath, "", nil, with(map[string]string{"Upload-Length": "4", "Upload-Metadata": "target_path " + b64("/t.bin")}))
	if st != http.StatusCreated {
		t.Fatalf("create: %d %s", st, b)
	}
	upload := strings.TrimPrefix(hdr.Get("Location"), c.url)
	const create, one = "/files/uploads/", "/files/uploads/{id}"
	tests := []struct {
		name, method, specPath, target string
		headers                        map[string]string
		body                           string
		status                         int
	}{
		{"no Tus-Resumable", "POST", create, UploadsPath, nil, "", 412},
		{"no target", "POST", create, UploadsPath, with(map[string]string{"Upload-Length": "4"}), "", 400},
		{"deferred length", "POST", create, UploadsPath, with(map[string]string{"Upload-Defer-Length": "1", "Upload-Metadata": "target_path " + b64("/d.bin")}), "", 411},
		{"missing parent", "POST", create, UploadsPath, with(map[string]string{"Upload-Length": "4", "Upload-Metadata": "target_path " + b64("/nope/x.bin")}), "", 404},
		{"target exists", "POST", create, UploadsPath, with(map[string]string{"Upload-Length": "4", "Upload-Metadata": "target_path " + b64("/docs/a.txt")}), "", 409},
		{"wrong method", "GET", create, UploadsPath, tus, "", 405},
		{"unknown upload", "HEAD", one, UploadsPath + "nope", tus, "", 404},
		{"unknown upload", "PATCH", one, UploadsPath + "nope", with(map[string]string{"Upload-Offset": "0", "Content-Type": "application/offset+octet-stream"}), "ab", 404},
		{"wrong offset", "PATCH", one, upload, with(map[string]string{"Upload-Offset": "2", "Content-Type": "application/offset+octet-stream"}), "ab", 409},
		{"wrong content type", "PATCH", one, upload, with(map[string]string{"Upload-Offset": "0", "Content-Type": "text/plain"}), "ab", 400},
		{"no Tus-Resumable", "PATCH", one, upload, nil, "ab", 412},
		{"unknown upload", "DELETE", one, UploadsPath + "nope", tus, "", 404},
	}
	for _, tt := range tests {
		var body []byte
		if tt.body != "" {
			body = []byte(tt.body)
		}
		st, hdr, b := c.do(tt.method, tt.target, "", body, tt.headers)
		if st != tt.status {
			t.Errorf("%s %s (%s): %d %s, want %d", tt.method, tt.target, tt.name, st, b, tt.status)
			continue
		}
		if tt.method != "HEAD" && hdr.Get("Content-Type") != apperr.ContentType {
			t.Errorf("%s %s (%s): the error is not a problem: %s", tt.method, tt.target, tt.name, b)
		}
		item := doc.Paths.Find(tt.specPath)
		declared := false
		for m, op := range item.Operations() {
			if (strings.EqualFold(m, tt.method) || st == http.StatusMethodNotAllowed) && op.Responses.Status(st) != nil {
				declared = true
			}
		}
		if !declared {
			t.Errorf("%s %s (%s) answers %d, which the spec does not declare", tt.method, tt.specPath, tt.name, st)
		}
	}
}
