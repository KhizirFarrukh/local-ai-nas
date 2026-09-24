package api

import (
	"bytes"
	_ "embed" // for go:embed
	"log/slog"
	"net/http"
	"strings"
	"time"

	apispec "github.com/KhizirFarrukh/local-ai-nas/api"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// DocsPath is where the offline API documentation is served (S01.5-T06).
const DocsPath = "/api/docs/"

// RedocSHA256 is the checksum of the vendored Redoc 2.5.4 bundle, as
// recorded in code-agent-docs/dependencies.md.
const RedocSHA256 = "dcaf76612bc4a3fbcc923a8966dee2f6146a5f32e5ce1b6f02dd60cbbf89500b"

var (
	//go:embed docs/index.html
	docsPage []byte
	//go:embed docs/redoc.standalone.js
	redocBundle []byte
)

// docFile is one file of the documentation.
type docFile struct {
	body        []byte
	contentType string
}

// apiDocs serves the offline API documentation (S01.5-T06): the page, the
// vendored Redoc bundle, and the spec. Every URL in the page is relative
// and nothing is loaded from the network, so the documentation works
// without internet access.
func apiDocs(log *slog.Logger) http.Handler {
	docs := map[string]docFile{
		"":                    {docsPage, "text/html; charset=utf-8"},
		"redoc.standalone.js": {redocBundle, "text/javascript; charset=utf-8"},
		"openapi.yaml":        {apispec.OpenAPI, "application/yaml; charset=utf-8"},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			apperr.Write(w, r, log, apperr.Newf(apperr.MethodNotAllowed, "%s is not allowed for %s; allowed: GET, HEAD", r.Method, r.URL.Path))
			return
		}
		f, ok := docs[strings.TrimPrefix(r.URL.Path, DocsPath)]
		if !ok {
			apperr.Write(w, r, log, apperr.Newf(apperr.NotFound, "there is no documentation file at %s", r.URL.Path))
			return
		}
		h := w.Header()
		h.Set("Content-Type", f.contentType)
		h.Set("Cache-Control", "no-cache")
		h.Set("X-Content-Type-Options", "nosniff")
		// Redoc injects its styles and runs its search in a worker.
		h.Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; worker-src 'self' blob:")
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(f.body))
	})
}

// docsRedirect sends /api/docs to /api/docs/, where the page's relative
// URLs work.
func docsRedirect(log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			apperr.Write(w, r, log, apperr.Newf(apperr.MethodNotAllowed, "%s is not allowed for %s; allowed: GET, HEAD", r.Method, r.URL.Path))
			return
		}
		http.Redirect(w, r, DocsPath, http.StatusMovedPermanently)
	})
}
