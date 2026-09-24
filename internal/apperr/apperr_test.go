package apperr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/logging"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

func TestKindMapping(t *testing.T) {
	tests := []struct {
		kind   Kind
		code   string
		status int
	}{
		{Internal, "internal", 500},
		{InvalidRequest, "invalid_request", 400},
		{InvalidName, "invalid_name", 400},
		{OutsideRoot, "outside_root", 400},
		{NotFound, "not_found", 404},
		{Conflict, "conflict", 409},
		{TooLarge, "too_large", 413},
		{InsufficientStorage, "insufficient_storage", 507},
		{NotAvailable, "not_available", 501},
		{Kind(999), "internal", 500}, // unknown kinds are treated as internal
	}
	if len(tests)-1 != len(kinds) {
		t.Fatalf("the test covers %d kinds, the package defines %d", len(tests)-1, len(kinds))
	}
	codes := map[string]Kind{}
	for _, tt := range tests {
		if tt.kind.Code() != tt.code || tt.kind.Status() != tt.status || tt.kind.String() != tt.code {
			t.Errorf("kind %d = %s/%d, want %s/%d", tt.kind, tt.kind.Code(), tt.kind.Status(), tt.code, tt.status)
		}
		if other, dup := codes[tt.code]; dup && tt.kind != Kind(999) {
			t.Errorf("code %q used by kinds %d and %d", tt.code, other, tt.kind)
		}
		codes[tt.code] = tt.kind
	}
}

func TestErrorChain(t *testing.T) {
	cause := errors.New("disk on fire")
	err := fmt.Errorf("finalize upload: %w", Wrap(InsufficientStorage, "not enough free space", cause))
	if KindOf(err) != InsufficientStorage {
		t.Errorf("KindOf = %v, want insufficient_storage", KindOf(err))
	}
	if !errors.Is(err, cause) {
		t.Error("the cause is not reachable with errors.Is")
	}
	if got, want := err.Error(), "finalize upload: insufficient_storage: not enough free space: disk on fire"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if KindOf(errors.New("plain")) != Internal || KindOf(nil) != Internal {
		t.Error("errors without an *Error must be internal")
	}
	if got := Newf(NotFound, "no item at %q", "/a").Error(); got != `not_found: no item at "/a"` {
		t.Errorf("Newf Error() = %q", got)
	}
	if got := New(Conflict, "").Error(); got != "conflict" {
		t.Errorf("New without detail = %q", got)
	}
}

func TestProblemFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Problem
	}{
		{
			name: "domain error shows its detail",
			err:  New(NotFound, "no item at /docs/a.txt"),
			want: Problem{Type: "about:blank", Title: "Not Found", Status: 404, Detail: "no item at /docs/a.txt", Code: "not_found", CorrelationID: "req1"},
		},
		{
			name: "wrapped cause stays hidden",
			err:  Wrap(Conflict, "the target exists", errors.New("EEXIST /srv/nas/files/u0001/a")),
			want: Problem{Type: "about:blank", Title: "Conflict", Status: 409, Detail: "the target exists", Code: "conflict", CorrelationID: "req1"},
		},
		{
			name: "plain error becomes a generic internal problem",
			err:  errors.New("open /srv/nas/.local-ai-nas/db/nas.db: permission denied"),
			want: Problem{Type: "about:blank", Title: "Internal Server Error", Status: 500, Detail: genericDetail, Code: "internal", CorrelationID: "req1"},
		},
		{
			name: "internal domain error hides its detail too",
			err:  New(Internal, "secret internal state"),
			want: Problem{Type: "about:blank", Title: "Internal Server Error", Status: 500, Detail: genericDetail, Code: "internal", CorrelationID: "req1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, ProblemFor(tt.err, "req1")); diff != "" {
				t.Errorf("problem (-want +got):\n%s", diff)
			}
		})
	}
}

// stack runs the handler made by mk behind the production middleware
// order and returns the server URL and the log buffer.
func stack(t *testing.T, mk func(logger *slog.Logger) http.HandlerFunc) (string, *bytes.Buffer) {
	t.Helper()
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	srv := testutil.NewServer(t, logging.RequestID(logging.AccessLog(logger)(Recover(logger)(mk(logger)))))
	return srv.URL, &logs
}

func get(t *testing.T, url string) (*http.Response, Problem) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var p Problem
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatalf("body is not a problem: %q: %v", body, err)
	}
	return resp, p
}

func TestWrite(t *testing.T) {
	url, logs := stack(t, func(logger *slog.Logger) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "999") // must not survive
			Write(w, r, logger, Wrap(TooLarge, "the file is larger than 1GiB", errors.New("declared 5GiB")))
		}
	})

	resp, p := get(t, url)
	if resp.StatusCode != 413 || resp.Header.Get("Content-Type") != ContentType || resp.Header.Get("Cache-Control") != "no-store" {
		t.Errorf("response %d %q %q", resp.StatusCode, resp.Header.Get("Content-Type"), resp.Header.Get("Cache-Control"))
	}
	id := resp.Header.Get(logging.RequestIDHeader)
	want := Problem{Type: "about:blank", Title: "Request Entity Too Large", Status: 413, Detail: "the file is larger than 1GiB", Code: "too_large", CorrelationID: id}
	if diff := cmp.Diff(want, p); diff != "" {
		t.Errorf("problem (-want +got):\n%s", diff)
	}
	if !strings.Contains(logs.String(), `"error":"too_large: the file is larger than 1GiB: declared 5GiB"`) {
		t.Errorf("the cause is not in the log:\n%s", logs)
	}
}

func TestWriteInternalLogsCause(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil)) // info level: 5xx still logged
	srv := testutil.NewServer(t, logging.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Write(w, r, logger, errors.New("sqlite: database is locked"))
	})))
	resp, p := get(t, srv.URL)
	if resp.StatusCode != 500 || p.Detail != genericDetail {
		t.Errorf("got %d %q", resp.StatusCode, p.Detail)
	}
	if !strings.Contains(logs.String(), "database is locked") || !strings.Contains(logs.String(), `"level":"ERROR"`) {
		t.Errorf("internal error not logged at error level:\n%s", logs.String())
	}
	if p.CorrelationID == "" || !strings.Contains(logs.String(), p.CorrelationID) {
		t.Errorf("correlation ID %q not found in the log", p.CorrelationID)
	}
}

func TestWriteNilError(t *testing.T) {
	rec := httptest.NewRecorder()
	Write(rec, httptest.NewRequest(http.MethodGet, "/", nil), slog.New(slog.DiscardHandler), nil)
	if rec.Code != 500 {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestRecoverPanic(t *testing.T) {
	url, logs := stack(t, func(*slog.Logger) http.HandlerFunc {
		return func(http.ResponseWriter, *http.Request) {
			panic("secret-panic-detail at /srv/nas")
		}
	})
	resp, p := get(t, url)
	if resp.StatusCode != 500 || p.Code != "internal" || p.Detail != genericDetail {
		t.Errorf("got %d %+v", resp.StatusCode, p)
	}
	if p.CorrelationID != resp.Header.Get(logging.RequestIDHeader) {
		t.Errorf("correlation ID %q, want the request ID", p.CorrelationID)
	}
	out := logs.String()
	if !strings.Contains(out, "secret-panic-detail") || !strings.Contains(out, "goroutine") {
		t.Errorf("panic value or stack missing from the log:\n%s", out)
	}
	if !strings.Contains(out, `"status":500`) {
		t.Errorf("the access log did not record the 500:\n%s", out)
	}
}

func TestRecoverPassesThrough(t *testing.T) {
	url, _ := stack(t, func(*slog.Logger) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, "fine")
		}
	})
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != 200 || string(body) != "fine" {
		t.Errorf("got %d %q", resp.StatusCode, body)
	}
}

func TestRecoverReraisesAbortHandler(t *testing.T) {
	h := Recover(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	defer func() {
		if v := recover(); v != http.ErrAbortHandler {
			t.Errorf("recovered %v, want http.ErrAbortHandler re-raised", v)
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	t.Error("the panic was swallowed")
}
