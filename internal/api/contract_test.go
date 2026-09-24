package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
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
	// The reserved photos routes, with any method.
	{http.MethodGet, "/api/v1/photos", http.StatusNotImplemented, "not_available"},
	{http.MethodPost, "/api/v1/photos", http.StatusNotImplemented, "not_available"},
	{http.MethodGet, "/api/v1/photos/timeline", http.StatusNotImplemented, "not_available"},
	{http.MethodPut, "/api/v1/photos/items/1", http.StatusNotImplemented, "not_available"},
	{http.MethodDelete, "/api/v1/photos/items/1", http.StatusNotImplemented, "not_available"},
}

// TestErrorResponsesMatchSchema is the S01.5-T03 contract test: every
// error response the API produces, across all endpoints, validates against
// the Problem schema in the spec.
func TestErrorResponsesMatchSchema(t *testing.T) {
	doc := loadSpec(t)
	h := New(Options{})

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

	for _, r := range Routes(Options{}) {
		if !covered[r.Pattern] {
			t.Errorf("route %q has no error case in errorCases", r.Pattern)
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
