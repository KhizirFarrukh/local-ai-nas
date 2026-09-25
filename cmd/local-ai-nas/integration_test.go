package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// liveServer is the program running as a separate process.
type liveServer struct {
	url  string
	root string // the storage root
}

// startLive runs the program (this test binary acting as main) with a new
// storage root and args, waits until it answers, and stops it at the end
// of the test.
func startLive(t *testing.T, args ...string) liveServer {
	t.Helper()
	root := testutil.StorageRoot(t)
	addr := freeAddr(t)
	cmd := exec.Command(os.Args[0], append([]string{"serve", "--storage-root", root, "--server-bind", addr}, args...)...)
	cmd.Env = append(os.Environ(), runAsMainEnv+"=1")
	var stderr lockedBuffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		<-done
	})
	url := "http://" + addr
	for deadline := time.Now().Add(20 * time.Second); ; {
		resp, err := http.Get(url + "/api/v1/system/health")
		if err == nil {
			_ = resp.Body.Close()
			return liveServer{url: url, root: root}
		}
		select {
		case <-done:
			t.Fatalf("the server stopped: %s", stderr.String())
		default:
		}
		if time.Now().After(deadline) {
			t.Fatalf("the server did not answer: %s", stderr.String())
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// harness sends requests and records which status of which operation of
// the spec they reached.
type harness struct {
	t       *testing.T
	doc     *openapi3.T
	client  *http.Client
	covered map[string]bool // "POST /files/folders 409"
}

// req describes one request.
type req struct {
	method, path string
	headers      map[string]string
	body         []byte
	length       int64 // 0: len(body); -1: unknown (chunked)
}

// expect sends r to s and checks that the answer has status want. op is
// the spec operation ("POST /files/folders") the status belongs to; for
// 405, every operation of that spec path is covered. It returns the
// answer's headers and body.
func (h *harness) expect(s liveServer, op string, r req, want int) (http.Header, []byte) {
	h.t.Helper()
	httpReq, err := http.NewRequestWithContext(h.t.Context(), r.method, s.url+r.path, bytes.NewReader(r.body))
	if err != nil {
		h.t.Fatal(err)
	}
	switch {
	case r.length < 0:
		httpReq.ContentLength = -1
		httpReq.Body = io.NopCloser(bytes.NewReader(r.body)) // no length: sent chunked
	case r.length > 0:
		httpReq.ContentLength = r.length
	}
	for k, v := range r.headers {
		httpReq.Header.Set(k, v)
	}
	resp, err := h.client.Do(httpReq)
	if err != nil {
		h.t.Fatalf("%s %s: %v", r.method, r.path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.t.Fatalf("%s %s: reading the answer: %v", r.method, r.path, err)
	}
	if resp.StatusCode != want {
		h.t.Errorf("%s %s: %d %s, want %d", r.method, r.path, resp.StatusCode, body, want)
		return resp.Header, body
	}
	switch {
	case op == "GET /system/health" && want == http.StatusServiceUnavailable:
		// A failed check answers with the health report, not a problem.
		var rep any
		if err := json.Unmarshal(body, &rep); err != nil {
			h.t.Errorf("the 503 health answer is not JSON: %s", body)
		} else if err := h.doc.Components.Schemas["HealthReport"].Value.VisitJSON(rep); err != nil {
			h.t.Errorf("the 503 health answer does not match HealthReport: %v", err)
		}
	case want >= 400 && r.method != http.MethodHead:
		var p any
		if resp.Header.Get("Content-Type") != apperr.ContentType || json.Unmarshal(body, &p) != nil {
			h.t.Errorf("%s %s: the %d is not a problem: %s", r.method, r.path, want, body)
		} else if err := h.doc.Components.Schemas["Problem"].Value.VisitJSON(p); err != nil {
			h.t.Errorf("%s %s: the problem does not match the schema: %v", r.method, r.path, err)
		}
	}
	h.covered[op+" "+strconv.Itoa(want)] = true
	if want == http.StatusMethodNotAllowed {
		_, specPath, _ := strings.Cut(op, " ")
		for m := range h.doc.Paths.Find(specPath).Operations() {
			h.covered[m+" "+specPath+" 405"] = true
		}
	}
	return resp.Header, body
}

// notCaused lists the declared statuses that no request from outside can
// cause on purpose, with the reason and where they are tested instead.
var notCaused = map[string]string{
	"500":                            "internal errors: unit tests inject faults (internal/apperr, internal/files, internal/uploads)",
	"POST /files/uploads/ 501":       "only when the API runs without a tus server; the program always has one (internal/api contract test)",
	"HEAD /files/uploads/{id} 423":   "tus lock contention: a timing race between two requests for one upload",
	"PATCH /files/uploads/{id} 423":  "tus lock contention: a timing race between two requests for one upload",
	"DELETE /files/uploads/{id} 423": "tus lock contention: a timing race between two requests for one upload",
	"PATCH /files/uploads/{id} 507":  "the copy fallback of finalize needs the upload directory on another file system (internal/uploads TestFinalizeCopyFallback)",
	"GET /files/content 409":         "a file replaced between the lookup and the open, five times in a row (internal/files stress test)",
}

// testedInS02_8 lists statuses of operations added during stage 2, whose
// requests are written with the stage's tests in S02.8 (RULES R6, S006).
// S02.8-T02 reaches each of them here and empties this list.
var testedInS02_8 = map[string]bool{
	"POST /files/archives 201":     true,
	"POST /files/archives 400":     true,
	"POST /files/archives 404":     true,
	"POST /files/archives 405":     true,
	"POST /files/archives 413":     true,
	"POST /files/archives 422":     true,
	"GET /files/archives/{id} 200": true,
	"GET /files/archives/{id} 404": true,
	"GET /files/archives/{id} 405": true,
}

func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

// TestIntegration is the S01.7-T01 integration suite: the program runs
// as a separate process and every status of every operation in the spec
// is reached over real HTTP (except the few in notCaused), with every
// error a valid problem.
func TestIntegration(t *testing.T) {
	isolateEnv(t)
	doc, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, doc: doc, covered: map[string]bool{},
		client: &http.Client{Transport: &http.Transport{ExpectContinueTimeout: 10 * time.Second}}}
	main := startLive(t, "--uploads-max-file-size", "1MiB", "--uploads-max-chunk-size", "64KiB", "--copy-sync-max-bytes", "1MiB")
	full := startLive(t, "--storage-free-space-reserve", "1000000TiB")
	jsonType := map[string]string{"Content-Type": "application/json"}
	post := func(path, body string) req {
		return req{method: "POST", path: path, headers: jsonType, body: []byte(body)}
	}
	big := `{"path":"/` + strings.Repeat("a", 2<<20) + `"}`

	// System.
	h.expect(main, "GET /system/health", req{method: "GET", path: "/api/v1/system/health"}, 200)
	h.expect(full, "GET /system/health", req{method: "GET", path: "/api/v1/system/health"}, 503)
	h.expect(main, "GET /system/health", req{method: "POST", path: "/api/v1/system/health"}, 405)
	h.expect(main, "GET /photos", req{method: "GET", path: "/api/v1/photos"}, 501)

	// Folders.
	h.expect(main, "POST /files/folders", post("/api/v1/files/folders", `{"path":"/docs"}`), 201)
	h.expect(main, "POST /files/folders", post("/api/v1/files/folders", `{"path":"/docs","on_conflict":"overwrite"}`), 200)
	h.expect(main, "POST /files/folders", post("/api/v1/files/folders", `not json`), 400)
	h.expect(main, "POST /files/folders", post("/api/v1/files/folders", `{"path":"/nope/x"}`), 404)
	h.expect(main, "POST /files/folders", post("/api/v1/files/folders", `{"path":"/docs"}`), 409)
	h.expect(main, "POST /files/folders", post("/api/v1/files/folders", big), 413)
	h.expect(main, "POST /files/folders", req{method: "GET", path: "/api/v1/files/folders"}, 405)

	// Simple upload.
	put := func(query, body string) req {
		return req{method: "PUT", path: "/api/v1/files/content?" + query, body: []byte(body)}
	}
	h.expect(main, "PUT /files/content", put("path=/docs/a.txt", "0123456789"), 201)
	h.expect(main, "PUT /files/content", put("path=/docs/a.txt&on_conflict=overwrite", "0123456789"), 200)
	h.expect(main, "PUT /files/content", put("", "x"), 400)
	h.expect(main, "PUT /files/content", put("path=/nope/a.txt", "x"), 404)
	h.expect(main, "PUT /files/content", put("path=/docs/a.txt", "x"), 409)
	h.expect(main, "PUT /files/content", req{method: "PUT", path: "/api/v1/files/content?path=/c.txt", body: []byte("x"), length: -1}, 411)
	h.expect(main, "PUT /files/content", req{method: "PUT", path: "/api/v1/files/content?path=/huge.bin",
		headers: map[string]string{"Expect": "100-continue"}, body: make([]byte, 2<<20)}, 413)
	h.expect(full, "PUT /files/content", put("path=/x.txt", "x"), 507)
	h.expect(main, "PUT /files/content", req{method: "POST", path: "/api/v1/files/content"}, 405)

	// Download.
	get := func(path string, headers map[string]string) req {
		return req{method: "GET", path: "/api/v1/files/content?path=" + path, headers: headers}
	}
	hdr, _ := h.expect(main, "GET /files/content", get("/docs/a.txt", nil), 200)
	h.expect(main, "GET /files/content", get("/docs/a.txt", map[string]string{"Range": "bytes=0-1"}), 206)
	h.expect(main, "GET /files/content", get("/docs/a.txt", map[string]string{"If-None-Match": hdr.Get("ETag")}), 304)
	h.expect(main, "GET /files/content", get("/docs", nil), 400)
	h.expect(main, "GET /files/content", get("/missing", nil), 404)
	h.expect(main, "GET /files/content", get("/docs/a.txt", map[string]string{"If-Match": `"other"`}), 412)
	h.expect(main, "GET /files/content", get("/docs/a.txt", map[string]string{"Range": "bytes=100-"}), 416)

	// Items: list, details, delete.
	h.expect(main, "GET /files/items", req{method: "GET", path: "/api/v1/files/items?path=/"}, 200)
	h.expect(main, "GET /files/items", req{method: "GET", path: "/api/v1/files/items"}, 400)
	h.expect(main, "GET /files/items", req{method: "GET", path: "/api/v1/files/items?path=/missing"}, 404)
	h.expect(main, "GET /files/items", req{method: "POST", path: "/api/v1/files/items"}, 405)
	h.expect(main, "PUT /files/content", put("path=/docs/gone.txt", "x"), 201)
	h.expect(main, "DELETE /files/items", req{method: "DELETE", path: "/api/v1/files/items?path=/docs/gone.txt"}, 204)
	h.expect(main, "DELETE /files/items", req{method: "DELETE", path: "/api/v1/files/items?path=/"}, 400)
	h.expect(main, "DELETE /files/items", req{method: "DELETE", path: "/api/v1/files/items?path=/missing"}, 404)
	h.expect(main, "DELETE /files/items", req{method: "DELETE", path: "/api/v1/files/items?path=/docs"}, 409)

	// Rename, move, copy.
	h.expect(main, "PUT /files/content", put("path=/docs/c.txt", "c"), 201)
	for _, op := range []string{"rename", "move", "copy"} {
		path := "/api/v1/files/operations/" + op
		opKey := "POST /files/operations/" + op
		h.expect(main, opKey, post(path, `not json`), 400)
		h.expect(main, opKey, post(path, big), 413)
		h.expect(main, opKey, req{method: "GET", path: path}, 405)
	}
	h.expect(main, "POST /files/operations/rename", post("/api/v1/files/operations/rename", `{"path":"/docs/a.txt","new_name":"b.txt"}`), 200)
	h.expect(main, "POST /files/operations/rename", post("/api/v1/files/operations/rename", `{"path":"/missing","new_name":"x"}`), 404)
	h.expect(main, "POST /files/operations/rename", post("/api/v1/files/operations/rename", `{"path":"/docs/b.txt","new_name":"c.txt"}`), 409)
	h.expect(main, "POST /files/operations/move", post("/api/v1/files/operations/move", `{"from":"/docs/b.txt","to":"/b.txt"}`), 200)
	h.expect(main, "POST /files/operations/move", post("/api/v1/files/operations/move", `{"from":"/missing","to":"/x"}`), 404)
	h.expect(main, "POST /files/operations/move", post("/api/v1/files/operations/move", `{"from":"/b.txt","to":"/docs/c.txt"}`), 409)
	h.expect(main, "POST /files/operations/copy", post("/api/v1/files/operations/copy", `{"from":"/b.txt","to":"/b2.txt"}`), 201)
	h.expect(main, "POST /files/operations/copy", post("/api/v1/files/operations/copy", `{"from":"/b.txt","to":"/b2.txt","on_conflict":"overwrite"}`), 200)
	h.expect(main, "POST /files/operations/copy", post("/api/v1/files/operations/copy", `{"from":"/missing","to":"/x"}`), 404)
	h.expect(main, "POST /files/operations/copy", post("/api/v1/files/operations/copy", `{"from":"/b.txt","to":"/b2.txt"}`), 409)
	h.expect(main, "POST /files/folders", post("/api/v1/files/folders", `{"path":"/big"}`), 201)
	for _, name := range []string{"one.bin", "two.bin"} {
		h.expect(main, "PUT /files/content", req{method: "PUT", path: "/api/v1/files/content?path=/big/" + name, body: make([]byte, 600<<10)}, 201)
	}
	h.expect(main, "POST /files/operations/copy", post("/api/v1/files/operations/copy", `{"from":"/big","to":"/big2"}`), 422)
	if err := os.WriteFile(filepath.Join(full.root, "files", "u0001", "src.txt"), []byte("s"), 0o600); err != nil {
		t.Fatal(err)
	}
	h.expect(full, "POST /files/operations/copy", post("/api/v1/files/operations/copy", `{"from":"/src.txt","to":"/dst.txt"}`), 507)

	// Resumable uploads (tus).
	tus := func(extra map[string]string) map[string]string {
		m := map[string]string{"Tus-Resumable": "1.0.0"}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}
	create := func(s liveServer, meta map[string]string, want int) string {
		t.Helper()
		hdr, _ := h.expect(s, "POST /files/uploads/", req{method: "POST", path: "/api/v1/files/uploads/", headers: tus(meta)}, want)
		return strings.TrimPrefix(hdr.Get("Location"), s.url)
	}
	up := create(main, map[string]string{"Upload-Length": "5", "Upload-Metadata": "target_path " + b64("/docs/t.bin")}, 201)
	create(main, map[string]string{"Upload-Length": "5"}, 400)
	create(main, map[string]string{"Upload-Length": "5", "Upload-Metadata": "target_path " + b64("/nope/t.bin")}, 404)
	create(main, map[string]string{"Upload-Length": "5", "Upload-Metadata": "target_path " + b64("/docs/c.txt")}, 409)
	create(main, map[string]string{"Upload-Defer-Length": "1", "Upload-Metadata": "target_path " + b64("/d.bin")}, 411)
	create(main, map[string]string{"Upload-Length": strconv.Itoa(2 << 20), "Upload-Metadata": "target_path " + b64("/e.bin")}, 413)
	create(full, map[string]string{"Upload-Length": "5", "Upload-Metadata": "target_path " + b64("/f.bin")}, 507)
	h.expect(main, "POST /files/uploads/", req{method: "POST", path: "/api/v1/files/uploads/"}, 412)
	h.expect(main, "POST /files/uploads/", req{method: "GET", path: "/api/v1/files/uploads/", headers: tus(nil)}, 405)

	unknown := "/api/v1/files/uploads/aaaaaaaaaaaaaaaaaaaaaaaaaa"
	octet := func(offset string) map[string]string {
		return tus(map[string]string{"Upload-Offset": offset, "Content-Type": "application/offset+octet-stream"})
	}
	h.expect(main, "HEAD /files/uploads/{id}", req{method: "HEAD", path: up, headers: tus(nil)}, 200)
	h.expect(main, "HEAD /files/uploads/{id}", req{method: "HEAD", path: unknown, headers: tus(nil)}, 404)
	h.expect(main, "PATCH /files/uploads/{id}", req{method: "PATCH", path: up, headers: octet("2"), body: []byte("ab")}, 409)
	h.expect(main, "PATCH /files/uploads/{id}", req{method: "PATCH", path: up, headers: tus(map[string]string{"Upload-Offset": "0", "Content-Type": "text/plain"}), body: []byte("ab")}, 400)
	h.expect(main, "PATCH /files/uploads/{id}", req{method: "PATCH", path: unknown, headers: octet("0"), body: []byte("ab")}, 404)
	h.expect(main, "PATCH /files/uploads/{id}", req{method: "PATCH", path: up, body: []byte("ab")}, 412)
	h.expect(main, "PATCH /files/uploads/{id}", req{method: "PATCH", path: up, headers: octet("0"), body: make([]byte, 100<<10)}, 413)
	h.expect(main, "PATCH /files/uploads/{id}", req{method: "GET", path: up, headers: tus(nil)}, 405)
	hdr, _ = h.expect(main, "PATCH /files/uploads/{id}", req{method: "PATCH", path: up, headers: octet("0"), body: []byte("tus!!")}, 204)
	if hdr.Get("Item-Path") != "/docs/t.bin" {
		t.Errorf("the finished upload's Item-Path = %q", hdr.Get("Item-Path"))
	}
	cancelled := create(main, map[string]string{"Upload-Length": "5", "Upload-Metadata": "target_path " + b64("/docs/cancel.bin")}, 201)
	h.expect(main, "DELETE /files/uploads/{id}", req{method: "DELETE", path: cancelled}, 412)
	h.expect(main, "DELETE /files/uploads/{id}", req{method: "DELETE", path: cancelled, headers: tus(nil)}, 204)
	h.expect(main, "DELETE /files/uploads/{id}", req{method: "DELETE", path: unknown, headers: tus(nil)}, 404)

	// Every declared status of every operation was reached, or is listed
	// in notCaused.
	var missing []string
	for path, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			for code := range op.Responses.Map() {
				key := method + " " + path + " " + code
				_, why := notCaused[key]
				_, whyAny := notCaused[code]
				if !h.covered[key] && !why && !whyAny && !testedInS02_8[key] {
					missing = append(missing, key)
				}
			}
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("declared statuses no request reached (%d):\n%s", len(missing), strings.Join(missing, "\n"))
	}
	t.Logf("%d operation statuses reached over HTTP", len(h.covered))
}
