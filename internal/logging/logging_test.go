package logging

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// lines decodes JSON log lines.
func lines(t *testing.T, data []byte) []map[string]any {
	t.Helper()
	var out []map[string]any
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatalf("log line is not JSON: %q: %v", sc.Text(), err)
		}
		out = append(out, m)
	}
	return out
}

func TestNewWritesStderrAndFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	var stderr bytes.Buffer
	logger, closer, err := New(Options{Level: "info", Stderr: &stderr, Dir: dir, FileMaxSize: 1 << 20, FileMaxFiles: 2})
	if err != nil {
		t.Fatal(err)
	}
	logger.Debug("hidden at info level")
	logger.Info("hello", "k", "v")
	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}
	file, err := os.ReadFile(filepath.Join(dir, "nas.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(file, stderr.Bytes()) {
		t.Errorf("file and stderr differ:\nfile:   %q\nstderr: %q", file, stderr.Bytes())
	}
	got := lines(t, file)
	if len(got) != 1 || got[0]["msg"] != "hello" || got[0]["k"] != "v" || got[0]["level"] != "INFO" {
		t.Errorf("log lines = %v, want one INFO line \"hello\" with k=v", got)
	}
}

func TestNewWithoutFile(t *testing.T) {
	var stderr bytes.Buffer
	logger, closer, err := New(Options{Level: "DEBUG", Stderr: &stderr})
	if err != nil {
		t.Fatal(err)
	}
	logger.Debug("shown")
	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}
	if got := lines(t, stderr.Bytes()); len(got) != 1 || got[0]["msg"] != "shown" {
		t.Errorf("log lines = %v", got)
	}
}

func TestNewErrors(t *testing.T) {
	if _, _, err := New(Options{Level: "loud"}); err == nil {
		t.Error("unknown level accepted")
	}
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := New(Options{Level: "info", Dir: filepath.Join(file, "logs"), FileMaxSize: 1, FileMaxFiles: 1}); err == nil {
		t.Error("a logs directory under a regular file was accepted")
	}
	if _, _, err := New(Options{Level: "info", Dir: t.TempDir()}); err == nil {
		t.Error("zero rotation settings accepted")
	}
}

func TestParseLevel(t *testing.T) {
	for in, want := range map[string]slog.Level{"debug": slog.LevelDebug, "INFO": slog.LevelInfo, "Warn": slog.LevelWarn, "error": slog.LevelError} {
		if got, err := ParseLevel(in); err != nil || got != want {
			t.Errorf("ParseLevel(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "trace", "info+2"} {
		if _, err := ParseLevel(bad); err == nil {
			t.Errorf("ParseLevel(%q) succeeded", bad)
		}
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestTeeWriterKeepsFileWhenStderrFails(t *testing.T) {
	var file bytes.Buffer
	tw := &teeWriter{primary: &file, secondary: failWriter{}}
	if _, err := tw.Write([]byte("line\n")); err == nil {
		t.Error("stderr failure not reported")
	}
	if file.String() != "line\n" {
		t.Errorf("file got %q, want the line despite the stderr failure", file.String())
	}
	tw = &teeWriter{primary: failWriter{}, secondary: &file}
	if _, err := tw.Write([]byte("two\n")); err == nil {
		t.Error("file failure not reported")
	}
}

// newTestStack returns a server running mux behind RequestID and AccessLog,
// and the buffer that receives the log lines.
func newTestStack(t *testing.T, level slog.Level, mux *http.ServeMux) (string, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: level}))
	srv := testutil.NewServer(t, RequestID(AccessLog(logger)(mux)))
	return srv.URL, &buf
}

func TestAccessLogFields(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/files/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = io.WriteString(w, "12345")
	})
	url, buf := newTestStack(t, slog.LevelInfo, mux)

	resp, err := http.Get(url + "/api/v1/files/items?path=/docs&limit=5")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	id := resp.Header.Get(RequestIDHeader)
	if len(id) != 32 {
		t.Errorf("generated request ID %q, want 32 hex characters", id)
	}

	got := lines(t, buf.Bytes())
	if len(got) != 1 {
		t.Fatalf("got %d log lines, want 1: %v", len(got), got)
	}
	l := got[0]
	checks := map[string]any{
		"msg": "request", "level": "INFO", "method": "GET", "route": "GET /api/v1/files/items",
		"status": float64(418), "bytes": float64(5), "request_id": id, "query": "limit=5&path=%2Fdocs",
	}
	for k, want := range checks {
		if l[k] != want {
			t.Errorf("field %q = %v, want %v", k, l[k], want)
		}
	}
	if d, ok := l["duration_ms"].(float64); !ok || d < 0 {
		t.Errorf("duration_ms = %v, want a non-negative number", l["duration_ms"])
	}
	if _, ok := l["headers"]; ok {
		t.Error("headers logged at info level")
	}
}

func TestAccessLogUnmatchedRouteAndDefaultStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /quiet", func(http.ResponseWriter, *http.Request) {})
	url, buf := newTestStack(t, slog.LevelInfo, mux)
	for _, p := range []string{"/quiet", "/nowhere"} {
		resp, err := http.Get(url + p)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
	}
	got := lines(t, buf.Bytes())
	if len(got) != 2 {
		t.Fatalf("got %d lines", len(got))
	}
	if got[0]["status"] != float64(200) || got[0]["bytes"] != float64(0) {
		t.Errorf("empty handler logged %v", got[0])
	}
	if got[1]["status"] != float64(404) || got[1]["route"] != "" {
		t.Errorf("unmatched route logged %v", got[1])
	}
}

func TestRequestIDIncoming(t *testing.T) {
	mux := http.NewServeMux()
	var seen string
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) { seen = RequestIDFrom(r.Context()) })
	url, _ := newTestStack(t, slog.LevelInfo, mux)

	tests := []struct {
		incoming string
		keep     bool
	}{
		{"3f2a9c1e-5b7d-4e8f-9a0b-1c2d3e4f5a6b", true},
		{"client.req:42_x", true},
		{"has space", false},
		{"line break", false},
		{strings.Repeat("a", 129), false},
		{"", false},
	}
	for _, tt := range tests {
		req, _ := http.NewRequest(http.MethodGet, url+"/", nil)
		if tt.incoming != "" {
			req.Header.Set(RequestIDHeader, tt.incoming)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		got := resp.Header.Get(RequestIDHeader)
		if got != seen {
			t.Errorf("response ID %q differs from context ID %q", got, seen)
		}
		if tt.keep && got != tt.incoming {
			t.Errorf("valid incoming ID %q replaced by %q", tt.incoming, got)
		}
		if !tt.keep && (got == tt.incoming || len(got) != 32) {
			t.Errorf("invalid incoming ID %q: got %q, want a new 32-character ID", tt.incoming, got)
		}
	}
}

func TestAccessLogRedaction(t *testing.T) {
	const secret = "s3cr3t-value"
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(http.ResponseWriter, *http.Request) {})
	url, buf := newTestStack(t, slog.LevelDebug, mux)

	req, _ := http.NewRequest(http.MethodGet, url+"/?name=ok&token="+secret+"&api_key="+secret+"&Password="+secret, nil)
	req.Header.Set("Authorization", "Bearer "+secret)
	req.Header.Set("Cookie", "session="+secret)
	req.Header.Set("X-Api-Key", secret)
	req.Header.Set("Accept", "text/plain")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	out := buf.String()
	if strings.Contains(out, secret) {
		t.Errorf("a secret reached the log:\n%s", out)
	}
	got := lines(t, buf.Bytes())
	if len(got) != 1 {
		t.Fatalf("got %d lines", len(got))
	}
	headers, _ := got[0]["headers"].(map[string]any)
	for _, h := range []string{"Authorization", "Cookie", "X-Api-Key"} {
		if v, _ := headers[h].([]any); len(v) != 1 || v[0] != Redacted {
			t.Errorf("header %s logged as %v, want [%s]", h, headers[h], Redacted)
		}
	}
	if v, _ := headers["Accept"].([]any); len(v) != 1 || v[0] != "text/plain" {
		t.Errorf("harmless header Accept logged as %v", headers["Accept"])
	}
	q, _ := got[0]["query"].(string)
	if !strings.Contains(q, "name=ok") || strings.Count(q, "%5BREDACTED%5D") != 3 {
		t.Errorf("query logged as %q, want name=ok and three redacted values", q)
	}
}

func TestRedactHeadersKeepsOriginal(t *testing.T) {
	h := http.Header{"Authorization": {"Bearer x"}}
	_ = RedactHeaders(h)
	if h.Get("Authorization") != "Bearer x" {
		t.Error("RedactHeaders changed its input")
	}
}

func TestRecorderFlush(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /stream", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "part")
		rc := http.NewResponseController(w)
		if err := rc.Flush(); err != nil {
			t.Errorf("ResponseController.Flush: %v", err)
		}
		// Deadlines are reached through recorder.Unwrap.
		if err := rc.SetWriteDeadline(time.Now().Add(time.Minute)); err != nil {
			t.Errorf("ResponseController.SetWriteDeadline: %v", err)
		}
		w.(http.Flusher).Flush()
	})
	url, buf := newTestStack(t, slog.LevelInfo, mux)
	resp, err := http.Get(url + "/stream")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) != "part" {
		t.Errorf("body = %q", body)
	}
	if got := lines(t, buf.Bytes()); len(got) != 1 || got[0]["bytes"] != float64(4) {
		t.Errorf("log = %v", got)
	}
}
