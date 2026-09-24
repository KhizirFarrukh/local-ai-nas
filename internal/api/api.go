// Package api is the HTTP layer of the server: the routes under /api/v1
// and the middleware around them (ADR-0002, S01.5). The routes are listed
// in one table, so tests can check all of them (for example that every
// route is under /api/v1 or /api/docs, S01.5-T02).
package api

import (
	"log/slog"
	"net/http"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/health"
	"github.com/KhizirFarrukh/local-ai-nas/internal/logging"
)

// DefaultMaxBodyBytes limits request bodies unless a route sets its own
// limit (uploads do, from S01.3).
const DefaultMaxBodyBytes = 1 << 20

// Options configures New.
type Options struct {
	Logger  *slog.Logger
	Version string         // shown by the health endpoint
	Checks  []health.Check // the health checks
	// MaxBodyBytes limits request bodies; 0 means DefaultMaxBodyBytes.
	MaxBodyBytes int64
}

// Route is one entry of the route table.
type Route struct {
	// Pattern is an http.ServeMux pattern, such as
	// "GET /api/v1/system/health". A pattern without a method matches all
	// methods.
	Pattern string
	Handler http.Handler
}

// Routes returns the route table.
func Routes(o Options) []Route {
	// The photos area exists on disk from S01, but its API is reserved
	// until the media stages (S01.2-T06).
	photos := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apperr.Write(w, r, o.Logger, apperr.New(apperr.NotAvailable, "the photos API is not available yet; it arrives with the media stages (S04)"))
	})
	return []Route{
		{"GET /api/v1/system/health", health.Handler(o.Version, o.Checks...)},
		{"/api/v1/photos", photos},
		{"/api/v1/photos/", photos},
	}
}

// New builds the complete HTTP handler. From the outside in: request ID,
// body limit, access log, panic recovery, routes. The body limit sits
// outside the access log because http.MaxBytesHandler passes a copy of the
// request on, and the access log must see the request the mux fills in
// (its route pattern).
func New(o Options) http.Handler {
	if o.Logger == nil {
		o.Logger = slog.New(slog.DiscardHandler)
	}
	if o.MaxBodyBytes <= 0 {
		o.MaxBodyBytes = DefaultMaxBodyBytes
	}
	mux := http.NewServeMux()
	for _, r := range Routes(o) {
		mux.Handle(r.Pattern, r.Handler)
	}
	var h = problemsForUnmatched(mux, o.Logger)
	h = apperr.Recover(o.Logger)(h)
	h = logging.AccessLog(o.Logger)(h)
	h = http.MaxBytesHandler(h, o.MaxBodyBytes)
	return logging.RequestID(h)
}
