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

func TestHandlerHealthy(t *testing.T) {
	h := Handler("1.2.3", Check{"database", func(context.Context) error { return nil }})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/system/health", nil))
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("got %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	var got Report
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := Report{Status: "ok", Version: "1.2.3", Checks: []Result{{Name: "database", Status: "ok"}}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("report (-want +got):\n%s", diff)
	}
}

func TestHandlerFailingCheck(t *testing.T) {
	h := Handler("dev",
		Check{"config", func(context.Context) error { return nil }},
		Check{"database", func(context.Context) error { return errors.New("database is locked") }},
	)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	var got Report
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := Report{Status: "fail", Version: "dev", Checks: []Result{
		{Name: "config", Status: "ok"},
		{Name: "database", Status: "fail", Error: "database is locked"},
	}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("report (-want +got):\n%s", diff)
	}
}

func TestCheckTimeout(t *testing.T) {
	start := time.Now()
	rep := Run(t.Context(), "dev", []Check{{"slow", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
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
