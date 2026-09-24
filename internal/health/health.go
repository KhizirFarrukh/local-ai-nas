// Package health reports the health of the server and its dependencies:
// at startup and on GET /api/v1/system/health (S01.1-T11, S01.2-T05).
package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Check is one named health check. Run returns a short detail when
// healthy, a *Warning when the server works but in a degraded way, and any
// other error when the condition fails.
type Check struct {
	Name string
	Run  func(ctx context.Context) (detail string, err error)
}

// Warning is a degraded but working condition, such as uploads that must
// be copied instead of renamed. It does not fail the health report.
type Warning struct {
	Detail string
}

func (w *Warning) Error() string { return w.Detail }

// Warnf returns a *Warning with a formatted detail.
func Warnf(format string, args ...any) error {
	return &Warning{Detail: fmt.Sprintf(format, args...)}
}

// Statuses of checks and of the whole report.
const (
	StatusOK   = "ok"
	StatusWarn = "warn"
	StatusFail = "fail"
)

// Report is the health response body.
type Report struct {
	// Status is "fail" if any check failed, else "warn" if any check
	// warned, else "ok".
	Status  string   `json:"status"`
	Version string   `json:"version"`
	Checks  []Result `json:"checks"`
}

// Result is the outcome of one check.
type Result struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
	Error  string `json:"error,omitempty"`
}

// checkTimeout bounds each check, so a stuck dependency cannot hang the
// health endpoint.
const checkTimeout = 2 * time.Second

// Run runs every check and returns the report.
func Run(ctx context.Context, version string, checks []Check) Report {
	rep := Report{Status: StatusOK, Version: version, Checks: make([]Result, 0, len(checks))}
	for _, c := range checks {
		cctx, cancel := context.WithTimeout(ctx, checkTimeout)
		detail, err := c.Run(cctx)
		cancel()
		res := Result{Name: c.Name, Status: StatusOK, Detail: detail}
		var w *Warning
		switch {
		case err == nil:
		case errors.As(err, &w):
			res.Status, res.Detail = StatusWarn, w.Detail
			if rep.Status == StatusOK {
				rep.Status = StatusWarn
			}
		default:
			res.Status, res.Error = StatusFail, err.Error()
			rep.Status = StatusFail
		}
		rep.Checks = append(rep.Checks, res)
	}
	return rep
}

// Handler serves the report as JSON: 200 unless a check failed, then 503.
func Handler(version string, checks ...Check) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rep := Run(r.Context(), version, checks)
		body, _ := json.Marshal(rep) // A Report of strings always marshals.
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if rep.Status == StatusFail {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_, _ = w.Write(append(body, '\n'))
	})
}
