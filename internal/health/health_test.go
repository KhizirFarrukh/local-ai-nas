package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func ok(detail string) func(context.Context) (string, error) {
	return func(context.Context) (string, error) { return detail, nil }
}

func serve(t *testing.T, h http.Handler) (int, Report) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/system/health", nil))
	var rep Report
	if err := json.Unmarshal(rec.Body.Bytes(), &rep); err != nil {
		t.Fatal(err)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	return rec.Code, rep
}

func TestHandlerHealthy(t *testing.T) {
	code, got := serve(t, Handler("1.2.3", Check{"database", ok("open")}))
	want := Report{Status: "ok", Version: "1.2.3", Checks: []Result{{Name: "database", Status: "ok", Detail: "open"}}}
	if code != 200 {
		t.Errorf("status = %d, want 200", code)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("report (-want +got):\n%s", diff)
	}
}

func TestHandlerStatuses(t *testing.T) {
	warn := Check{"same_filesystem", func(context.Context) (string, error) {
		return "", Warnf("uploads are copied (upload_finalize_mode=%s)", "copy")
	}}
	fail := Check{"database", func(context.Context) (string, error) { return "", errors.New("database is locked") }}

	tests := []struct {
		name     string
		checks   []Check
		wantCode int
		want     Report
	}{
		{
			name:     "a warning keeps the server healthy",
			checks:   []Check{{"config", ok("valid")}, warn},
			wantCode: 200,
			want: Report{Status: "warn", Version: "dev", Checks: []Result{
				{Name: "config", Status: "ok", Detail: "valid"},
				{Name: "same_filesystem", Status: "warn", Detail: "uploads are copied (upload_finalize_mode=copy)"},
			}},
		},
		{
			name:     "a failure wins over a warning",
			checks:   []Check{warn, fail, {"config", ok("valid")}},
			wantCode: 503,
			want: Report{Status: "fail", Version: "dev", Checks: []Result{
				{Name: "same_filesystem", Status: "warn", Detail: "uploads are copied (upload_finalize_mode=copy)"},
				{Name: "database", Status: "fail", Error: "database is locked"},
				{Name: "config", Status: "ok", Detail: "valid"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, got := serve(t, Handler("dev", tt.checks...))
			if code != tt.wantCode {
				t.Errorf("status = %d, want %d", code, tt.wantCode)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("report (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWarning(t *testing.T) {
	err := Warnf("uploads are copied (mode=%s)", "copy")
	var w *Warning
	if !errors.As(err, &w) || err.Error() != "uploads are copied (mode=copy)" {
		t.Errorf("Warnf = %#v (%q)", err, err)
	}
}

func TestCheckTimeout(t *testing.T) {
	start := time.Now()
	rep := Run(t.Context(), "dev", []Check{{"slow", func(ctx context.Context) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	}}})
	if rep.Status != "fail" || rep.Checks[0].Error == "" {
		t.Errorf("a stuck check did not fail: %+v", rep)
	}
	if d := time.Since(start); d > checkTimeout+time.Second {
		t.Errorf("Run took %s, want about %s", d, checkTimeout)
	}
}

func TestNoChecks(t *testing.T) {
	rep := Run(t.Context(), "dev", nil)
	if rep.Status != "ok" || rep.Checks == nil || len(rep.Checks) != 0 {
		t.Errorf("report = %+v, want ok with an empty (not null) checks list", rep)
	}
}
