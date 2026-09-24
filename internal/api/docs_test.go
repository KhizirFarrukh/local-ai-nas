package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	apispec "github.com/KhizirFarrukh/local-ai-nas/api"
)

// TestAPIDocs is the S01.5-T06 test: the documentation page references
// nothing outside this server, the bundle is the registered one, and the
// served spec is the contract.
func TestAPIDocs(t *testing.T) {
	h := New(Options{})
	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}

	page := get("/api/docs/")
	if page.Code != http.StatusOK || !strings.HasPrefix(page.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("the page: %d %s", page.Code, page.Header().Get("Content-Type"))
	}
	html := page.Body.String()
	if strings.Contains(html, "://") || strings.Contains(html, "//cdn") {
		t.Error("the page references an absolute URL")
	}
	for _, m := range regexp.MustCompile(`(?:src|href|spec-url)="([^"]*)"`).FindAllStringSubmatch(html, -1) {
		if m[1] != "redoc.standalone.js" && m[1] != "openapi.yaml" {
			t.Errorf("the page references %q", m[1])
		}
	}
	if !strings.Contains(page.Header().Get("Content-Security-Policy"), "default-src 'self'") {
		t.Errorf("Content-Security-Policy = %q", page.Header().Get("Content-Security-Policy"))
	}

	bundle := get("/api/docs/redoc.standalone.js")
	sum := sha256.Sum256(bundle.Body.Bytes())
	if bundle.Code != http.StatusOK || hex.EncodeToString(sum[:]) != RedocSHA256 {
		t.Errorf("the Redoc bundle: %d, SHA-256 %x, want %s", bundle.Code, sum, RedocSHA256)
	}

	spec := get("/api/docs/openapi.yaml")
	if spec.Code != http.StatusOK || !bytes.Equal(spec.Body.Bytes(), apispec.OpenAPI) {
		t.Fatalf("the spec: %d, %d bytes", spec.Code, spec.Body.Len())
	}
	doc, err := openapi3.NewLoader().LoadFromData(spec.Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Validate(t.Context()); err != nil {
		t.Errorf("the served spec is not valid OpenAPI: %v", err)
	}

	if rec := get("/api/docs"); rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != "/api/docs/" {
		t.Errorf("/api/docs: %d → %q", rec.Code, rec.Header().Get("Location"))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodHead, "/api/docs/", nil))
	if rec.Code != http.StatusOK || rec.Body.Len() != 0 {
		t.Errorf("HEAD /api/docs/: %d with %d bytes", rec.Code, rec.Body.Len())
	}
}
