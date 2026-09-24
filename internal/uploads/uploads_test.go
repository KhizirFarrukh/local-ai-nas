package uploads

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

const basePath = "/api/v1/files/uploads/"

// testServer starts the tus server on a real HTTP server and returns its
// URL, the upload directory, and the index.
func testServer(t *testing.T) (string, string, Index) {
	t.Helper()
	dir := t.TempDir()
	d, err := db.Open(t.Context(), filepath.Join(t.TempDir(), db.FileName))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	if _, err := d.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	s, err := New(Options{Dir: dir, DB: d, Namespace: "u0001", BasePath: basePath})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle(basePath, s)
	mux.Handle(strings.TrimSuffix(basePath, "/"), s)
	return testutil.NewServer(t, mux).URL, dir, s.index
}

// metadata encodes Upload-Metadata as tus defines it.
func metadata(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, pairs[i]+" "+base64.StdEncoding.EncodeToString([]byte(pairs[i+1])))
	}
	return strings.Join(parts, ",")
}

// tusRequest sends one tus request.
func tusRequest(t *testing.T, ctx context.Context, method, url string, headers map[string]string, body io.Reader, length int64) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Tus-Resumable", "1.0.0")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.ContentLength = length
	}
	return http.DefaultClient.Do(req)
}

// TestResumeInterruptedUpload is the S01.4-T01 acceptance test: a client
// creates an upload, is cut off part-way, asks for the offset, resumes
// from there, and the stored data equals the source.
func TestResumeInterruptedUpload(t *testing.T) {
	url, dir, index := testServer(t)
	data := make([]byte, 3<<20+123)
	rng := rand.New(rand.NewPCG(5, 6))
	for i := range data {
		data[i] = byte(rng.Uint32())
	}

	// Create.
	resp, err := tusRequest(t, t.Context(), http.MethodPost, url+basePath, map[string]string{
		"Upload-Length":   strconv.Itoa(len(data)),
		"Upload-Metadata": metadata(MetaTargetPath, "/docs/big.bin", MetaOnConflict, "rename"),
	}, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	location := resp.Header.Get("Location")
	if resp.StatusCode != http.StatusCreated || location == "" {
		t.Fatalf("create: %d, Location %q", resp.StatusCode, location)
	}
	id := location[strings.LastIndex(location, "/")+1:]

	// The session is recorded.
	s, err := index.Get(t.Context(), id)
	if err != nil || s.TargetPath != "/docs/big.bin" || s.OnConflict != "rename" || s.DeclaredSize != int64(len(data)) || s.Namespace != "u0001" {
		t.Fatalf("session %+v, %v", s, err)
	}
	if !s.ExpiresAt.After(s.CreatedAt) {
		t.Errorf("the session expires at %v, created at %v", s.ExpiresAt, s.CreatedAt)
	}

	// Send the first part, then cut the connection.
	part := 1<<20 + 7
	pr, pw := io.Pipe()
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		resp, err := tusRequest(t, ctx, http.MethodPatch, location, map[string]string{
			"Upload-Offset": "0", "Content-Type": "application/offset+octet-stream",
		}, pr, int64(len(data)))
		if err == nil {
			_ = resp.Body.Close()
		}
		done <- err
	}()
	if _, err := pw.Write(data[:part]); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond) // let the server store what it received
	cancel()
	_ = pw.CloseWithError(errors.New("cut off")) // the transport waits for its body writer
	if err := <-done; err == nil {
		t.Fatal("the interrupted PATCH succeeded")
	}

	// Ask where to resume. The server stores what it received once the
	// cut-off request has wound down.
	var offset int
	for deadline := time.Now().Add(15 * time.Second); ; {
		resp, err := tusRequest(t, t.Context(), http.MethodHead, location, nil, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		offset, _ = strconv.Atoi(resp.Header.Get("Upload-Offset"))
		if resp.StatusCode == http.StatusOK && offset > 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if offset <= 0 || offset > part {
		t.Fatalf("offset after the interruption = %d, want 1..%d", offset, part)
	}

	// Resume from the server's offset.
	resp, err = tusRequest(t, t.Context(), http.MethodPatch, location, map[string]string{
		"Upload-Offset": strconv.Itoa(offset), "Content-Type": "application/offset+octet-stream",
	}, bytes.NewReader(data[offset:]), int64(len(data)-offset))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent || resp.Header.Get("Upload-Offset") != strconv.Itoa(len(data)) {
		t.Fatalf("resume: %d, offset %s", resp.StatusCode, resp.Header.Get("Upload-Offset"))
	}
	stored, err := os.ReadFile(filepath.Join(dir, id))
	if err != nil || sha256.Sum256(stored) != sha256.Sum256(data) {
		t.Fatalf("the stored upload differs from the source (%d of %d bytes, %v)", len(stored), len(data), err)
	}
	t.Logf("resumed at %d of %d bytes", offset, len(data))
}

// TestCreateRefusals: bad metadata is refused with a problem body before
// anything is stored.
func TestCreateRefusals(t *testing.T) {
	url, dir, _ := testServer(t)
	tests := []struct {
		name    string
		headers map[string]string
		status  int
		code    string
	}{
		{"no target", map[string]string{"Upload-Length": "5"}, 400, "invalid_request"},
		{"relative target", map[string]string{"Upload-Length": "5", "Upload-Metadata": metadata(MetaTargetPath, "docs/a")}, 400, "invalid_request"},
		{"bad policy", map[string]string{"Upload-Length": "5", "Upload-Metadata": metadata(MetaTargetPath, "/a", MetaOnConflict, "merge")}, 400, "invalid_request"},
		{"bad checksum", map[string]string{"Upload-Length": "5", "Upload-Metadata": metadata(MetaTargetPath, "/a", MetaSHA256, "abc")}, 400, "invalid_request"},
		{"deferred length", map[string]string{"Upload-Defer-Length": "1", "Upload-Metadata": metadata(MetaTargetPath, "/a")}, 411, "length_required"},
	}
	for _, tt := range tests {
		resp, err := tusRequest(t, t.Context(), http.MethodPost, url+basePath, tt.headers, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		var p apperr.Problem
		err = json.NewDecoder(resp.Body).Decode(&p)
		_ = resp.Body.Close()
		if err != nil || resp.StatusCode != tt.status || p.Code != tt.code || resp.Header.Get("Content-Type") != apperr.ContentType {
			t.Errorf("%s: %d %+v (%v), want %d %s as a problem", tt.name, resp.StatusCode, p, err, tt.status, tt.code)
		}
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("refused uploads left %d entries in the upload directory", len(entries))
	}
}

// TestTerminate: a cancelled upload loses its data and its session.
func TestTerminate(t *testing.T) {
	url, dir, index := testServer(t)
	resp, err := tusRequest(t, t.Context(), http.MethodPost, url+basePath, map[string]string{
		"Upload-Length": "10", "Upload-Metadata": metadata(MetaTargetPath, "/x.bin"),
	}, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	location := resp.Header.Get("Location")
	id := location[strings.LastIndex(location, "/")+1:]
	resp, err = tusRequest(t, t.Context(), http.MethodDelete, location, nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("terminate: %d", resp.StatusCode)
	}
	if _, err := index.Get(t.Context(), id); !errors.Is(err, ErrNoSession) {
		t.Errorf("the session is still there: %v", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("the terminated upload left %d entries", len(entries))
	}
}

// TestProtocolErrorsAreProblems: errors of the tus protocol itself keep
// tusd's status (tus clients act on it) and get a problem body.
func TestProtocolErrorsAreProblems(t *testing.T) {
	url, _, _ := testServer(t)
	resp, err := tusRequest(t, t.Context(), http.MethodPost, url+basePath, map[string]string{
		"Upload-Length": "4", "Upload-Metadata": metadata(MetaTargetPath, "/p.bin"),
	}, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	location := resp.Header.Get("Location")

	tests := []struct {
		name, method, target string
		headers              map[string]string
		body                 string
		status               int
		code                 string
	}{
		{"no Tus-Resumable", http.MethodPost, url + basePath, map[string]string{"Tus-Resumable": ""}, "", 412, "precondition_failed"},
		{"unknown upload", http.MethodPatch, url + basePath + "nope", map[string]string{"Upload-Offset": "0", "Content-Type": "application/offset+octet-stream"}, "ab", 404, "not_found"},
		{"wrong offset", http.MethodPatch, location, map[string]string{"Upload-Offset": "2", "Content-Type": "application/offset+octet-stream"}, "ab", 409, "conflict"},
		{"wrong content type", http.MethodPatch, location, map[string]string{"Upload-Offset": "0", "Content-Type": "text/plain"}, "ab", 400, "invalid_request"},
		{"no length", http.MethodPost, url + basePath, map[string]string{"Upload-Metadata": metadata(MetaTargetPath, "/q")}, "", 400, "invalid_request"},
		{"wrong method", http.MethodGet, url + basePath, nil, "", 405, "method_not_allowed"},
	}
	for _, tt := range tests {
		var body io.Reader
		if tt.body != "" {
			body = strings.NewReader(tt.body)
		}
		resp, err := tusRequest(t, t.Context(), tt.method, tt.target, tt.headers, body, int64(len(tt.body)))
		if err != nil {
			t.Fatal(err)
		}
		var p apperr.Problem
		err = json.NewDecoder(resp.Body).Decode(&p)
		_ = resp.Body.Close()
		if err != nil || resp.StatusCode != tt.status || p.Code != tt.code || p.Status != tt.status ||
			resp.Header.Get("Content-Type") != apperr.ContentType || resp.Header.Get("Tus-Resumable") == "" && tt.status != 405 && tt.status != 412 {
			t.Errorf("%s: %d %+v (%v), headers %v; want %d %s as a problem", tt.name, resp.StatusCode, p, err, resp.Header, tt.status, tt.code)
		}
	}
}
