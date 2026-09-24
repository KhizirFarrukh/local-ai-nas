package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// putContent calls PUT /api/v1/files/content with body.
func putContent(t *testing.T, h http.Handler, query string, body []byte) (*httptest.ResponseRecorder, gen.FileItem) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/files/content?"+query, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/octet-stream")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var it gen.FileItem
	if rec.Code == http.StatusCreated || rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &it); err != nil {
			t.Fatalf("response is not a FileItem: %v\n%s", err, rec.Body)
		}
	}
	return rec, it
}

func randomBytes(n int) []byte {
	rng := rand.New(rand.NewPCG(uint64(n), 7))
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(rng.Uint32())
	}
	return b
}

func TestUploadEndpoint(t *testing.T) {
	doc := loadSpec(t)
	svc, dir := testFilesDir(t)
	h := New(Options{Files: svc})
	big := randomBytes(3<<20 + 5) // over the 1 MiB limit of JSON bodies
	tests := []struct {
		query    string
		body     []byte
		wantCode int
		wantPath string
		wantDisk string // the file on disk that must hold body
	}{
		{"path=/docs/big.bin", big, http.StatusCreated, "/docs/big.bin", "docs/big.bin"},
		{"path=/docs/big.bin&on_conflict=rename", []byte("second"), http.StatusCreated, "/docs/big (1).bin", "docs/big (1).bin"},
		{"path=/docs/a.txt&on_conflict=overwrite", []byte("replaced"), http.StatusOK, "/docs/a.txt", "docs/a.txt"},
		{"path=/empty.txt", nil, http.StatusCreated, "/empty.txt", "empty.txt"},
		{"path=" + url.QueryEscape("/docs/résumé & notes.txt"), []byte("x"), http.StatusCreated, "/docs/résumé & notes.txt", "docs/résumé & notes.txt"},
	}
	for _, tt := range tests {
		rec, it := putContent(t, h, tt.query, tt.body)
		if rec.Code != tt.wantCode || it.Path != tt.wantPath || it.Kind != "file" || it.Size != int64(len(tt.body)) {
			t.Errorf("%s → %d %+v, want %d %s", tt.query, rec.Code, it, tt.wantCode, tt.wantPath)
			continue
		}
		var body any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if err := doc.Components.Schemas["FileItem"].Value.VisitJSON(body); err != nil {
			t.Errorf("%s: the response does not match FileItem: %v", tt.query, err)
		}
		if it.Etag == nil || it.Mime == nil {
			t.Errorf("%s: no ETag or MIME type in %+v", tt.query, it)
		}
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(tt.wantDisk)))
		if err != nil || sha256.Sum256(got) != sha256.Sum256(tt.body) {
			t.Errorf("%s: the file on disk differs from the upload (%v)", tt.query, err)
		}
	}
}

func TestUploadInvalidInputNeverReachesService(t *testing.T) {
	svc := &recordingFiles{}
	h := New(Options{Files: svc, MaxUploadBytes: 10})
	for _, tt := range []struct {
		query       string
		length      int64
		wantStatus  int
		wantProblem string
	}{
		{"", 3, http.StatusBadRequest, "invalid_request"},
		{"path=/a.txt&on_conflict=merge", 3, http.StatusBadRequest, "invalid_request"},
		{"path=/a.txt", -1, http.StatusLengthRequired, "length_required"},
		{"path=/a.txt", 11, http.StatusRequestEntityTooLarge, "too_large"},
	} {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/files/content?"+tt.query, strings.NewReader("abc"))
		req.ContentLength = tt.length
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		var p apperr.Problem
		if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil || rec.Code != tt.wantStatus || p.Code != tt.wantProblem {
			t.Errorf("%q (length %d) → %d %s, want %d %s", tt.query, tt.length, rec.Code, rec.Body, tt.wantStatus, tt.wantProblem)
		}
	}
	if svc.calls != 0 {
		t.Errorf("the service was called %d times with invalid input", svc.calls)
	}
}

// countingBody counts what the client's transport reads from a body.
type countingBody struct {
	r io.Reader
	n int
}

func (c *countingBody) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += n
	return n, err
}

// TestUploadOverHTTP runs uploads through a real server and client: the
// size checks answer before the body is sent (Expect: 100-continue), a
// body of unknown size gets 411, and a large body streams through.
func TestUploadOverHTTP(t *testing.T) {
	svc, dir := testFilesDir(t)
	srv := testutil.NewServer(t, New(Options{Files: svc, MaxUploadBytes: 4 << 20}))
	client := &http.Client{Transport: &http.Transport{ExpectContinueTimeout: 10 * time.Second}}
	put := func(query string, body io.Reader, length int64, expect bool) *http.Response {
		t.Helper()
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, srv.URL+"/api/v1/files/content?"+query, body)
		if err != nil {
			t.Fatal(err)
		}
		req.ContentLength = length
		if expect {
			req.Header.Set("Expect", "100-continue")
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = resp.Body.Close() })
		return resp
	}

	// Refused before the body: over the limit, and a conflict.
	for _, tt := range []struct {
		query  string
		length int64
		want   int
	}{
		{"path=/huge.bin", 5 << 20, http.StatusRequestEntityTooLarge},
		{"path=/docs/a.txt", 1 << 20, http.StatusConflict},
	} {
		body := &countingBody{r: bytes.NewReader(make([]byte, tt.length))}
		resp := put(tt.query, body, tt.length, true)
		if resp.StatusCode != tt.want || body.n != 0 {
			t.Errorf("%s: %d after sending %d bytes, want %d before any byte", tt.query, resp.StatusCode, body.n, tt.want)
		}
	}

	// Unknown size (chunked).
	if resp := put("path=/chunked.bin", io.MultiReader(strings.NewReader("abc")), -1, false); resp.StatusCode != http.StatusLengthRequired {
		t.Errorf("chunked upload → %d, want 411", resp.StatusCode)
	}

	// A body of several MiB, with and without Expect.
	for _, expect := range []bool{true, false} {
		data := randomBytes(4<<20 - 1)
		name := "/stream-" + map[bool]string{true: "expect", false: "plain"}[expect] + ".bin"
		resp := put("path="+name, bytes.NewReader(data), int64(len(data)), expect)
		if resp.StatusCode != http.StatusCreated {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("%s → %d %s", name, resp.StatusCode, b)
		}
		got, err := os.ReadFile(filepath.Join(dir, name[1:]))
		if err != nil || sha256.Sum256(got) != sha256.Sum256(data) {
			t.Errorf("%s: the file on disk differs from the upload (%v)", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "chunked.bin")); !os.IsNotExist(err) {
		t.Errorf("a refused upload left a file: %v", err)
	}
}
