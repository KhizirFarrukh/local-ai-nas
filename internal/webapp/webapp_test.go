package webapp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// build is a small stand-in for web/build: the page with its <meta> policy,
// a hashed asset, a plain file, and a hidden one.
var build = fstest.MapFS{
	"index.html": {Data: []byte(`<!doctype html><html><head>` +
		`<meta http-equiv="content-security-policy" content="default-src 'self'; script-src 'self' 'sha256-abc='">` +
		`</head><body><div id="app"></div></body></html>`)},
	"_app/immutable/entry/start.A1b2.js": {Data: []byte("console.log('start')")},
	"_app/version.json":                  {Data: []byte(`{"version":"1"}`)},
	"favicon.png":                        {Data: []byte("\x89PNG")},
	"robots.txt":                         {Data: []byte("User-agent: *\n")},
	"pdfjs/wasm/openjpeg.wasm":           {Data: []byte("\x00asm")},
	"odd.xyz":                            {Data: []byte("?")},
	".gitkeep":                           {Data: []byte{}},
}

func serve(t *testing.T, h http.Handler, method, target string, header ...string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, target, nil)
	for i := 0; i+1 < len(header); i += 2 {
		r.Header.Set(header[i], header[i+1])
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func newHandler(t *testing.T) *Handler {
	t.Helper()
	h, err := New(build)
	if err != nil {
		t.Fatal(err)
	}
	if !h.Built() {
		t.Fatal("the stand-in build has index.html")
	}
	return h
}

func TestServesTheBuild(t *testing.T) {
	h := newHandler(t)
	for _, tc := range []struct {
		target, contentType, cache, bodyStart string
	}{
		{"/", "text/html; charset=utf-8", "no-cache", "<!doctype html>"},
		{"/_app/immutable/entry/start.A1b2.js", "text/javascript; charset=utf-8", "public, max-age=31536000, immutable", "console.log"},
		{"/_app/version.json", "application/json", "no-cache", `{"version"`},
		{"/favicon.png", "image/png", "no-cache", "\x89PNG"},
		{"/robots.txt", "text/plain; charset=utf-8", "no-cache", "User-agent"},
		{"/pdfjs/wasm/openjpeg.wasm", "application/wasm", "no-cache", "\x00asm"},
		{"/odd.xyz", "application/octet-stream", "no-cache", "?"},
	} {
		w := serve(t, h, http.MethodGet, tc.target)
		if w.Code != http.StatusOK {
			t.Errorf("%s: status %d", tc.target, w.Code)
			continue
		}
		if got := w.Header().Get("Content-Type"); got != tc.contentType {
			t.Errorf("%s: Content-Type %q, want %q", tc.target, got, tc.contentType)
		}
		if got := w.Header().Get("Cache-Control"); got != tc.cache {
			t.Errorf("%s: Cache-Control %q, want %q", tc.target, got, tc.cache)
		}
		if !strings.HasPrefix(w.Body.String(), tc.bodyStart) {
			t.Errorf("%s: body %q", tc.target, w.Body.String())
		}
	}
}

func TestAppRoutesGetThePage(t *testing.T) {
	h := newHandler(t)
	for _, target := range []string{"/files", "/files/docs/a%20b", "/settings", "/files/report.pdf", "/files/../settings"} {
		w := serve(t, h, http.MethodGet, target)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `<div id="app">`) {
			t.Errorf("%s: %d, the page was not served", target, w.Code)
		}
	}
}

func TestMissingAssetsAndHiddenFilesAre404(t *testing.T) {
	h := newHandler(t)
	for _, target := range []string{"/_app/immutable/gone.js", "/missing.js", "/.gitkeep", "/pdfjs/.secret", "/.env"} {
		w := serve(t, h, http.MethodGet, target)
		if w.Code != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404 (never the page)", target, w.Code)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	h := newHandler(t)
	w := serve(t, h, http.MethodGet, "/files")
	want := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"Referrer-Policy":         "same-origin",
		"X-Frame-Options":         "DENY",
		"Content-Security-Policy": "default-src 'self'; script-src 'self' 'sha256-abc='; frame-ancestors 'none'",
	}
	for k, v := range want {
		if got := w.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestConditionalRequestsAndMethods(t *testing.T) {
	h := newHandler(t)
	first := serve(t, h, http.MethodGet, "/robots.txt")
	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("no ETag")
	}
	if w := serve(t, h, http.MethodGet, "/robots.txt", "If-None-Match", etag); w.Code != http.StatusNotModified {
		t.Errorf("If-None-Match: status %d, want 304", w.Code)
	}
	if w := serve(t, h, http.MethodHead, "/robots.txt"); w.Code != http.StatusOK || w.Body.Len() != 0 {
		t.Errorf("HEAD: %d with %d bytes", w.Code, w.Body.Len())
	}
	w := serve(t, h, http.MethodPost, "/files")
	if w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != "GET, HEAD" {
		t.Errorf("POST: %d, Allow %q", w.Code, w.Header().Get("Allow"))
	}
}

func TestPageWithoutAPolicyGetsTheFallback(t *testing.T) {
	h, err := New(fstest.MapFS{"index.html": {Data: []byte("<html></html>")}})
	if err != nil {
		t.Fatal(err)
	}
	w := serve(t, h, http.MethodGet, "/")
	if got := w.Header().Get("Content-Security-Policy"); got != fallbackPolicy+"; "+frameAncestors {
		t.Errorf("policy %q", got)
	}
}

// TestNoticeWithoutABuild covers the binary built without the interface:
// web/build holds only its .gitkeep (see bug S02-B01), and the core
// explains how to build it instead of failing.
func TestNoticeWithoutABuild(t *testing.T) {
	h, err := New(fstest.MapFS{".gitkeep": {Data: []byte{}}})
	if err != nil {
		t.Fatal(err)
	}
	if h.Built() {
		t.Fatal("a build with only .gitkeep counts as built")
	}
	w := serve(t, h, http.MethodGet, "/files/anything")
	body, _ := io.ReadAll(w.Body)
	if w.Code != http.StatusOK || !strings.Contains(string(body), "The web interface is not part of this build.") {
		t.Errorf("notice: %d %q", w.Code, body)
	}
	if got := w.Header().Get("Content-Security-Policy"); got != noticePolicy {
		t.Errorf("notice policy %q", got)
	}
}
