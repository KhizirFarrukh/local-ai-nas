package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// client sends requests to a real test server.
type client struct {
	t   *testing.T
	url string
}

// do sends one request and returns the status, the headers, and the body.
func (c client) do(method, path, contentType string, body []byte, headers map[string]string) (int, http.Header, []byte) {
	c.t.Helper()
	req, err := http.NewRequestWithContext(c.t.Context(), method, c.url+path, bytes.NewReader(body))
	if err != nil {
		c.t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatal(err)
	}
	return resp.StatusCode, resp.Header, b
}

func q(p string) string { return url.QueryEscape(p) }

// TestFileOperationsOverHTTP is the S01.3 acceptance test (plan S01.3,
// criterion 1): every file operation, through a real server and client,
// on a real temporary file system.
func TestFileOperationsOverHTTP(t *testing.T) {
	svc, _ := testFilesDir(t)
	c := client{t, testutil.NewServer(t, New(Options{Files: svc})).URL}
	data := randomBytes(3<<20 + 11)
	step := func(name string, gotStatus, wantStatus int, body []byte) {
		t.Helper()
		if gotStatus != wantStatus {
			t.Fatalf("%s: status %d, want %d: %s", name, gotStatus, wantStatus, body)
		}
	}

	st, _, b := c.do("POST", "/api/v1/files/folders", "application/json", []byte(`{"path":"/work/sub","parents":true}`), nil)
	step("create folder", st, 201, b)
	st, _, b = c.do("PUT", "/api/v1/files/content?path="+q("/work/sub/data.bin"), "application/octet-stream", data, nil)
	step("upload", st, 201, b)

	st, _, b = c.do("GET", "/api/v1/files/items?path="+q("/work/sub"), "", nil, nil)
	step("list", st, 200, b)
	var list gen.ItemsResponse
	if err := json.Unmarshal(b, &list); err != nil || len(*list.Items) != 1 || (*list.Items)[0].Size != int64(len(data)) {
		t.Fatalf("listing: %s", b)
	}
	st, _, b = c.do("GET", "/api/v1/files/items?path="+q("/work/sub/data.bin"), "", nil, nil)
	step("details", st, 200, b)
	var details gen.ItemsResponse
	if err := json.Unmarshal(b, &details); err != nil || details.Item.Etag == nil {
		t.Fatalf("details: %s", b)
	}

	st, h, b := c.do("GET", "/api/v1/files/content?path="+q("/work/sub/data.bin"), "", nil, map[string]string{"Range": "bytes=1048570-1048589"})
	step("ranged download", st, 206, b)
	if !bytes.Equal(b, data[1048570:1048590]) || h.Get("Content-Range") != "bytes 1048570-1048589/3145739" {
		t.Errorf("range: %d bytes, Content-Range %q", len(b), h.Get("Content-Range"))
	}
	st, _, b = c.do("GET", "/api/v1/files/content?path="+q("/work/sub/data.bin"), "", nil, map[string]string{"If-None-Match": *details.Item.Etag})
	step("conditional download", st, 304, b)

	st, _, b = c.do("POST", "/api/v1/files/operations/rename", "application/json", []byte(`{"path":"/work/sub/data.bin","new_name":"renamed.bin"}`), nil)
	step("rename", st, 200, b)
	st, _, b = c.do("POST", "/api/v1/files/operations/move", "application/json", []byte(`{"from":"/work/sub/renamed.bin","to":"/work/moved.bin"}`), nil)
	step("move", st, 200, b)
	st, _, b = c.do("POST", "/api/v1/files/operations/copy", "application/json", []byte(`{"from":"/work","to":"/work-copy"}`), nil)
	step("copy", st, 201, b)
	st, _, b = c.do("GET", "/api/v1/files/content?path="+q("/work-copy/moved.bin"), "", nil, nil)
	step("download of the copy", st, 200, b[:0])
	if sha256.Sum256(b) != sha256.Sum256(data) {
		t.Error("the copied file differs from the upload")
	}

	st, _, b = c.do("DELETE", "/api/v1/files/items?path="+q("/work"), "", nil, nil)
	step("delete without recursive", st, 409, b)
	st, _, b = c.do("DELETE", "/api/v1/files/items?path="+q("/work")+"&recursive=true", "", nil, nil)
	step("delete", st, 204, b)
	st, _, b = c.do("GET", "/api/v1/files/items?path="+q("/work"), "", nil, nil)
	step("deleted", st, 404, b)

	st, _, b = c.do("GET", "/api/v1/files/items?path=/", "", nil, nil)
	step("final listing", st, 200, b)
	var root gen.ItemsResponse
	_ = json.Unmarshal(b, &root)
	if got := strings.Join(names(root.Items), ","); got != "docs,empty,readme.md,work-copy" {
		t.Errorf("the root at the end = %s", got)
	}
}

func TestDeleteInvalidInputNeverReachesService(t *testing.T) {
	svc := &recordingFiles{}
	h := New(Options{Files: svc})
	for _, query := range []string{"", "recursive=true", "path=/a&recursive=yes", "path=/a&recursive="} {
		st, _, b := client{t, testutil.NewServer(t, h).URL}.do("DELETE", "/api/v1/files/items?"+query, "", nil, nil)
		if st != http.StatusBadRequest {
			t.Errorf("%q → %d %s, want 400", query, st, b)
		}
	}
	if svc.calls != 0 {
		t.Errorf("the service was called %d times with invalid input", svc.calls)
	}
}
