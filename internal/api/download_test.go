package api

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// getContent calls GET (or HEAD) /api/v1/files/content with headers.
func getContent(h http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/v1/files/content?path="+url.QueryEscape(path), nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestDownloadEndpoint(t *testing.T) {
	doc := loadSpec(t)
	h := New(Options{Files: testFiles(t)})
	if rec, _ := putContent(t, h, "path=/digits.txt", []byte("0123456789")); rec.Code != http.StatusCreated {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body)
	}

	full := getContent(h, http.MethodGet, "/digits.txt", nil)
	etag, lastMod := full.Header().Get("ETag"), full.Header().Get("Last-Modified")
	wantHeaders := map[string]string{
		"Content-Type":            "text/plain; charset=utf-8",
		"Content-Length":          "10",
		"Accept-Ranges":           "bytes",
		"Content-Disposition":     `attachment; filename="digits.txt"`,
		"Cache-Control":           "private, no-cache",
		"X-Content-Type-Options":  "nosniff",
		"Content-Security-Policy": "default-src 'none'; sandbox",
	}
	if full.Code != http.StatusOK || full.Body.String() != "0123456789" || etag == "" || lastMod == "" {
		t.Fatalf("GET → %d %q, ETag %q, Last-Modified %q", full.Code, full.Body, etag, lastMod)
	}
	for k, v := range wantHeaders {
		if got := full.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if _, rep := getItems(t, h, url.Values{"path": {"/digits.txt"}}); rep.Item.Etag == nil || *rep.Item.Etag != etag {
		t.Errorf("the download's ETag %s differs from the item's %v", etag, rep.Item.Etag)
	}
	longAgo := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Format(http.TimeFormat)

	tests := []struct {
		name         string
		method       string
		headers      map[string]string
		wantCode     int
		wantBody     string
		contentRange string
	}{
		{"single range", "GET", map[string]string{"Range": "bytes=0-4"}, 206, "01234", "bytes 0-4/10"},
		{"open-ended range", "GET", map[string]string{"Range": "bytes=5-"}, 206, "56789", "bytes 5-9/10"},
		{"suffix range", "GET", map[string]string{"Range": "bytes=-3"}, 206, "789", "bytes 7-9/10"},
		{"range past the end", "GET", map[string]string{"Range": "bytes=8-100"}, 206, "89", "bytes 8-9/10"},
		{"unsatisfiable range", "GET", map[string]string{"Range": "bytes=10-"}, 416, "", "bytes */10"},
		{"malformed range", "GET", map[string]string{"Range": "bytes=x-y"}, 416, "", ""},
		{"not modified (ETag)", "GET", map[string]string{"If-None-Match": etag}, 304, "", ""},
		{"modified (other ETag)", "GET", map[string]string{"If-None-Match": `"other"`}, 200, "0123456789", ""},
		{"not modified (time)", "GET", map[string]string{"If-Modified-Since": lastMod}, 304, "", ""},
		{"If-Range matches", "GET", map[string]string{"Range": "bytes=0-1", "If-Range": etag}, 206, "01", "bytes 0-1/10"},
		{"If-Range stale", "GET", map[string]string{"Range": "bytes=0-1", "If-Range": `"old"`}, 200, "0123456789", ""},
		{"If-Match holds", "GET", map[string]string{"If-Match": etag}, 200, "0123456789", ""},
		{"If-Match fails", "GET", map[string]string{"If-Match": `"other"`}, 412, "", ""},
		{"If-Unmodified-Since fails", "GET", map[string]string{"If-Unmodified-Since": longAgo}, 412, "", ""},
		{"HEAD", "HEAD", nil, 200, "", ""},
	}
	for _, tt := range tests {
		rec := getContent(h, tt.method, "/digits.txt", tt.headers)
		if rec.Code != tt.wantCode {
			t.Errorf("%s: status %d, want %d (%s)", tt.name, rec.Code, tt.wantCode, rec.Body)
			continue
		}
		if got := rec.Header().Get("Content-Range"); got != tt.contentRange {
			t.Errorf("%s: Content-Range %q, want %q", tt.name, got, tt.contentRange)
		}
		if rec.Code >= 400 {
			validateProblem(t, doc, tt.name, rec)
			if rec.Header().Get("Content-Disposition") != "" || rec.Header().Get("Cache-Control") != "no-store" {
				t.Errorf("%s: an error response keeps download headers: %v", tt.name, rec.Header())
			}
			continue
		}
		if rec.Body.String() != tt.wantBody {
			t.Errorf("%s: body %q, want %q", tt.name, rec.Body, tt.wantBody)
		}
	}

	// Several ranges: multipart/byteranges.
	rec := getContent(h, http.MethodGet, "/digits.txt", map[string]string{"Range": "bytes=0-1,4-5"})
	mediaType, params, err := mime.ParseMediaType(rec.Header().Get("Content-Type"))
	if rec.Code != http.StatusPartialContent || err != nil || mediaType != "multipart/byteranges" {
		t.Fatalf("multi-range → %d %q (%v)", rec.Code, rec.Header().Get("Content-Type"), err)
	}
	var parts []string
	mr := multipart.NewReader(rec.Body, params["boundary"])
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(p)
		parts = append(parts, p.Header.Get("Content-Range")+"="+string(b))
	}
	if strings.Join(parts, " ") != "bytes 0-1/10=01 bytes 4-5/10=45" {
		t.Errorf("multi-range parts = %v", parts)
	}
}

func TestDownloadErrors(t *testing.T) {
	h := New(Options{Files: testFiles(t)})
	for _, tt := range []struct {
		path string
		want int
	}{
		{"/docs", http.StatusBadRequest},
		{"/", http.StatusBadRequest},
		{"/missing.txt", http.StatusNotFound},
		{"/.local-ai-nas-tmp-x.part", http.StatusNotFound},
		{"/docs/../../x", http.StatusBadRequest},
	} {
		if rec := getContent(h, http.MethodGet, tt.path, nil); rec.Code != tt.want || rec.Header().Get("Content-Disposition") != "" {
			t.Errorf("%s → %d %v, want %d without Content-Disposition", tt.path, rec.Code, rec.Header(), tt.want)
		}
	}
}

func TestContentDisposition(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"a.txt", `attachment; filename="a.txt"`},
		{"my report (1).pdf", `attachment; filename="my report (1).pdf"`},
		{"résumé.pdf", `attachment; filename="r_sum_.pdf"; filename*=UTF-8''r%C3%A9sum%C3%A9.pdf`},
		{`say "hi".txt`, `attachment; filename="say _hi_.txt"; filename*=UTF-8''say%20%22hi%22.txt`},
		{"50%.txt", `attachment; filename="50_.txt"; filename*=UTF-8''50%25.txt`},
		{"中文 ✓.txt", `attachment; filename="__ _.txt"; filename*=UTF-8''%E4%B8%AD%E6%96%87%20%E2%9C%93.txt`},
		{"semi;colon,x.txt", `attachment; filename="semi;colon,x.txt"`},
	}
	for _, tt := range tests {
		got := attachment(tt.name)
		if got != tt.want {
			t.Errorf("attachment(%q) = %s, want %s", tt.name, got, tt.want)
		}
		// A standard parser gets the exact name back.
		disp, params, err := mime.ParseMediaType(got)
		if err != nil || disp != "attachment" || params["filename"] != tt.name {
			t.Errorf("parsing %s gives %q %q (%v), want the name %q", got, disp, params["filename"], err, tt.name)
		}
	}
}

// TestDownloadUnicodeName: a file with a non-ASCII name downloads with
// its exact name.
func TestDownloadUnicodeName(t *testing.T) {
	h := New(Options{Files: testFiles(t)})
	name := "/Übersicht «2026».txt"
	if rec, _ := putContent(t, h, "path="+url.QueryEscape(name), []byte("x")); rec.Code != http.StatusCreated {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body)
	}
	rec := getContent(h, http.MethodGet, name, nil)
	_, params, err := mime.ParseMediaType(rec.Header().Get("Content-Disposition"))
	if rec.Code != http.StatusOK || err != nil || params["filename"] != name[1:] {
		t.Errorf("→ %d, Content-Disposition %q (%v)", rec.Code, rec.Header().Get("Content-Disposition"), err)
	}
}
