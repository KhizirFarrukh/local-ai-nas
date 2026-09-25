// Package webapp serves the embedded web interface (S02.1-T02) at every
// path outside /api: the files of the SvelteKit build, index.html for the
// app's own routes (SPA fallback), long caching for hashed assets, and
// security headers (NFR-022). The API itself stays in internal/api.
package webapp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
)

// indexName is the app's page, the fallback for app routes.
const indexName = "index.html"

// immutablePrefix holds build files whose names carry a content hash, so
// they can be cached forever.
const immutablePrefix = "_app/immutable/"

// frameAncestors is added to the page's own policy: a <meta> policy
// cannot carry it (CSP 3), so only the header can keep other sites from
// framing the app.
const frameAncestors = "frame-ancestors 'none'"

// fallbackPolicy applies when the page has no <meta> policy of its own.
const fallbackPolicy = "default-src 'self'; object-src 'none'; base-uri 'self'"

// noticePolicy applies to the notice shown when the interface is not built.
const noticePolicy = "default-src 'none'; " + frameAncestors

// notice is the page served when the binary was built without the web
// interface (web/build holds only its .gitkeep).
const notice = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>local-ai-nas</title></head>
<body>
<h1>local-ai-nas</h1>
<p>The web interface is not part of this build.</p>
<p>To include it, run <code>pnpm install</code> and <code>pnpm build</code> in the <code>web</code> folder, then build or run the server again. The API is available under <code>/api/v1</code>, and its documentation at <a href="/api/docs/">/api/docs/</a>.</p>
</body>
</html>
`

// metaPolicy finds the policy SvelteKit writes into index.html
// (kit.csp in web/svelte.config.js).
var metaPolicy = regexp.MustCompile(`<meta http-equiv="content-security-policy" content="([^"]*)"`)

// contentTypes are the media types of the files a build contains. They are
// fixed here, rather than taken from the operating system, whose table can
// be wrong (Windows registry entries have mapped .js to text/plain, which
// nosniff then refuses to run).
var contentTypes = map[string]string{
	".html":        "text/html; charset=utf-8",
	".js":          "text/javascript; charset=utf-8",
	".mjs":         "text/javascript; charset=utf-8",
	".css":         "text/css; charset=utf-8",
	".json":        "application/json",
	".map":         "application/json",
	".svg":         "image/svg+xml",
	".png":         "image/png",
	".ico":         "image/x-icon",
	".webp":        "image/webp",
	".woff2":       "font/woff2",
	".wasm":        "application/wasm",
	".txt":         "text/plain; charset=utf-8",
	".webmanifest": "application/manifest+json",
}

// file is one embedded file, read once at startup.
type file struct {
	data        []byte
	etag        string
	contentType string
}

// Handler serves the web interface.
type Handler struct {
	files  map[string]file
	policy string // the Content-Security-Policy header for the app
	built  bool   // false when the build holds no index.html
}

// New reads the built interface from build (web.Build()).
func New(build fs.FS) (*Handler, error) {
	h := &Handler{files: map[string]file{}}
	err := fs.WalkDir(build, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || hidden(name) {
			return nil
		}
		data, err := fs.ReadFile(build, name)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		h.files[name] = file{
			data:        data,
			etag:        `"` + hex.EncodeToString(sum[:12]) + `"`,
			contentType: contentType(name),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	index, ok := h.files[indexName]
	h.built = ok
	h.policy = fallbackPolicy + "; " + frameAncestors
	if ok {
		if m := metaPolicy.FindSubmatch(index.data); m != nil {
			h.policy = string(m[1]) + "; " + frameAncestors
		}
	}
	return h, nil
}

// Built reports whether the interface is part of this binary.
func (h *Handler) Built() bool { return h.built }

// ServeHTTP answers GET and HEAD for every path outside /api.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Frame-Options", "DENY")
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		header.Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.built {
		header.Set("Content-Security-Policy", noticePolicy)
		header.Set("Cache-Control", "no-cache")
		header.Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(notice))
		return
	}
	header.Set("Content-Security-Policy", h.policy)
	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	f, ok := h.files[name]
	switch {
	case ok:
	case missingAsset(name):
		http.Error(w, "not found", http.StatusNotFound)
		return
	default:
		// An app route, such as /files/docs: the app reads the path itself.
		name, f = indexName, h.files[indexName]
	}
	if strings.HasPrefix(name, immutablePrefix) {
		header.Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		header.Set("Cache-Control", "no-cache")
	}
	header.Set("Content-Type", f.contentType)
	header.Set("ETag", f.etag)
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(f.data))
}

// missingAsset reports whether a path that is not in the build names a file
// of the build rather than an app route. Such a path gets 404, so a missing
// script is never answered with the HTML page.
func missingAsset(name string) bool {
	return strings.HasPrefix(name, "_app/") || hidden(name) || path.Ext(name) != "" && !strings.Contains(name, "/")
}

// hidden reports whether any segment of name starts with a dot, such as
// build/.gitkeep. Hidden files are never served.
func hidden(name string) bool {
	for seg := range strings.SplitSeq(name, "/") {
		if strings.HasPrefix(seg, ".") && seg != "." {
			return true
		}
	}
	return false
}

// contentType returns the media type for a file name.
func contentType(name string) string {
	ext := strings.ToLower(path.Ext(name))
	if t, ok := contentTypes[ext]; ok {
		return t
	}
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return "application/octet-stream"
}
