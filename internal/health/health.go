// Package health reports the health of the server and its dependencies for
// GET /api/v1/system/health (S01.1-T11). S01.2-T05 adds the storage checks
// (root writable, same filesystem, free space).
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Check is one named health check. Run returns nil when healthy.
type Check struct {
	Name string
	Run  func(ctx context.Context) error
}

// Report is the health response body.
type Report struct {
	// Status is "ok" when every check passes, otherwise "fail".
	Status  string   `json:"status"`
	Version string   `json:"version"`
	Checks  []Result `json:"checks"`
}

// Result is the outcome of one check.
type Result struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// checkTimeout bounds each check, so a stuck dependency cannot hang the
// health endpoint.
const checkTimeout = 2 * time.Second

// Run runs every check and returns the report.
func Run(ctx context.Context, version string, checks []Check) Report {
	rep := Report{Status: "ok", Version: version, Checks: make([]Result, 0, len(checks))}
	for _, c := range checks {
		cctx, cancel := context.WithTimeout(ctx, checkTimeout)
		err := c.Run(cctx)
		cancel()
		res := Result{Name: c.Name, Status: "ok"}
		if err != nil {
			res.Status, res.Error = "fail", err.Error()
			rep.Status = "fail"
		}
		rep.Checks = append(rep.Checks, res)
	}
	return rep
}

// Handler serves the report as JSON: 200 when healthy, 503 otherwise.
func Handler(version string, checks ...Check) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rep := Run(r.Context(), version, checks)
		body, _ := json.Marshal(rep) // Report of strings always marshals.
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if rep.Status != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_, _ = w.Write(append(body, '\n'))
	})
}
