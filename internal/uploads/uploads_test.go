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
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

const basePath = "/api/v1/files/uploads/"

// fixture is the files area behind the test servers.
var fixture = map[string]string{"docs/.keep": "", "taken.txt": "x", "f.txt": "f", "folder/.keep": ""}

// testServer starts the tus server on a real HTTP server, finishing
// uploads into a real files area holding fixture, and returns its URL,
// the upload directory, the index, and the files area's directory. opts
// adjust the files service.
func testServer(t *testing.T, opts ...func(*files.Options)) (string, string, Index, string) {
	t.Helper()
	l := storage.NewLayout(testutil.StorageRoot(t), storage.Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	area := l.Area(storage.FilesArea, storage.DefaultNamespace)
	if err := testutil.WriteFiles(area, fixture); err != nil {
		t.Fatal(err)
	}
	fo := files.Options{}
	for _, o := range opts {
		o(&fo)
	}
	d, err := db.Open(t.Context(), filepath.Join(t.TempDir(), db.FileName))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	if _, err := d.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	s, err := New(Options{
		Dir: l.TmpUploads, DB: d, Files: files.NewLocal(storage.NewResolver(l), fo),
		Namespace: storage.DefaultNamespace, BasePath: basePath, MaxSize: 1 << 30,
	})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle(basePath, s)
	mux.Handle(strings.TrimSuffix(basePath, "/"), s)
	return testutil.NewServer(t, mux).URL, l.TmpUploads, s.index, area
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
	url, dir, index, _ := testServer(t)
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
	url, dir, _, _ := testServer(t)
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
	url, dir, index, _ := testServer(t)
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
	url, _, _, _ := testServer(t)
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

// TestCreateChecksTarget is the S01.4-T02 acceptance test: an upload whose
// target the files service would refuse is refused when it is created,
// with a problem, and leaves no data and no session behind.
func TestCreateChecksTarget(t *testing.T) {
	lowSpace := func(string) (uint64, error) { return 1 << 20, nil }
	tests := []struct {
		name   string
		meta   []string
		length string
		space  storage.FreeFunc
		status int
		code   string
	}{
		{"outside the root", []string{MetaTargetPath, "/../u0002/x"}, "5", nil, 400, "outside_root"},
		{"encoded separator", []string{MetaTargetPath, "/a%2Fb"}, "5", nil, 400, "invalid_name"},
		{"reserved name", []string{MetaTargetPath, "/docs/aux.txt"}, "5", nil, 400, "invalid_name"},
		{"temporary prefix", []string{MetaTargetPath, "/" + storage.TempPrefix + "x"}, "5", nil, 400, "invalid_name"},
		{"the root", []string{MetaTargetPath, "/"}, "5", nil, 400, "invalid_name"},
		{"missing parent", []string{MetaTargetPath, "/nope/x.bin"}, "5", nil, 404, "not_found"},
		{"file as parent", []string{MetaTargetPath, "/f.txt/x.bin"}, "5", nil, 409, "conflict"},
		{"target exists", []string{MetaTargetPath, "/taken.txt"}, "5", nil, 409, "conflict"},
		{"overwrite a folder", []string{MetaTargetPath, "/folder", MetaOnConflict, "overwrite"}, "5", nil, 409, "conflict"},
		{"over the size limit", []string{MetaTargetPath, "/big.bin"}, strconv.Itoa(1<<30 + 1), nil, 413, "too_large"},
		{"no free space", []string{MetaTargetPath, "/x.bin"}, "5", lowSpace, 507, "insufficient_storage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []func(*files.Options)
			if tt.space != nil {
				opts = append(opts, func(o *files.Options) {
					o.Space = storage.NewSpaceGuard(t.TempDir(), 1<<30, tt.space)
				})
			}
			url, dir, index, _ := testServer(t, opts...)
			resp, err := tusRequest(t, t.Context(), http.MethodPost, url+basePath, map[string]string{
				"Upload-Length": tt.length, "Upload-Metadata": metadata(tt.meta...),
			}, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			var p apperr.Problem
			err = json.NewDecoder(resp.Body).Decode(&p)
			_ = resp.Body.Close()
			if err != nil || resp.StatusCode != tt.status || p.Code != tt.code {
				t.Errorf("%d %+v (%v), want %d %s", resp.StatusCode, p, err, tt.status, tt.code)
			}
			if entries, _ := os.ReadDir(dir); len(entries) != 0 {
				t.Errorf("a refused upload left %d entries in the upload directory", len(entries))
			}
			var n int
			if err := index.db.Read.QueryRowContext(t.Context(), `SELECT count(*) FROM uploads`).Scan(&n); err != nil || n != 0 {
				t.Errorf("%d session rows (%v), want none", n, err)
			}
		})
	}
	// Links among the parents are refused too (where links can be made).
	url, _, _, area := testServer(t)
	if testutil.TrySymlink(t, filepath.Join(area, "docs"), filepath.Join(area, "link")) {
		resp, err := tusRequest(t, t.Context(), http.MethodPost, url+basePath, map[string]string{
			"Upload-Length": "5", "Upload-Metadata": metadata(MetaTargetPath, "/link/x.bin"),
		}, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("an upload through a link: %d, want 400", resp.StatusCode)
		}
	}
}
