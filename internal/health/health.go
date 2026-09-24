// Package health runs the health checks of the server and its
// dependencies: at startup and for GET /api/v1/system/health, which the API
// layer serves (S01.1-T11, S01.2-T05).
package health

import (
	"context"
	"errors"
	"fmt"
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
