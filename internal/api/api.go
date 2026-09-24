// Package api is the HTTP layer of the server: the routes under /api/v1
// and the middleware around them (ADR-0002, S01.5). The routes are listed
// in one table, so tests can check all of them (for example that every
// route is under /api/v1 or /api/docs, S01.5-T02).
package api

import (
	"log/slog"
	"net/http"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/health"
	"github.com/KhizirFarrukh/local-ai-nas/internal/logging"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// DefaultMaxBodyBytes limits request bodies unless a route sets its own
// limit (uploads do, from S01.3).
const DefaultMaxBodyBytes = 1 << 20

// DefaultMaxUploadBytes is the default largest simple upload, the default
// of the setting uploads.max_file_size.
const DefaultMaxUploadBytes = 100 << 30

// Options configures New.
type Options struct {
	Logger  *slog.Logger
	Version string         // shown by the health endpoint
	Checks  []health.Check // the health checks
	// Files is the files area. Handlers reach the disk only through it
	// (S01.3-T01); the file endpoints use it from S01.3-T02.
	Files files.Service
	// MaxBodyBytes limits request bodies; 0 means DefaultMaxBodyBytes.
	MaxBodyBytes int64
	// MaxUploadBytes is the largest file a simple upload accepts
	// (uploads.max_file_size); 0 means DefaultMaxUploadBytes.
	MaxUploadBytes int64
}

// Route is one entry of the route table.
type Route struct {
	// Pattern is an http.ServeMux pattern, such as
	// "GET /api/v1/system/health". A pattern without a method matches all
	// methods.
	Pattern string
	Handler http.Handler
	// MaxBody limits the request body; 0 means Options.MaxBodyBytes.
	MaxBody int64
}

// Routes returns the route table. Operations of the spec go through the
// generated handlers (spec-first); a test checks that the table and the
// spec list the same operations.
func Routes(o Options) []Route {
	if o.Logger == nil {
		o.Logger = slog.New(slog.DiscardHandler)
	}
	if o.Files == nil {
		o.Files = noFiles{}
	}
	if o.MaxUploadBytes <= 0 {
		o.MaxUploadBytes = DefaultMaxUploadBytes
	}
	g := generated(&server{version: o.Version, checks: o.Checks, files: o.Files, owner: storage.DefaultNamespace, log: o.Logger}, o.Logger)

	// The photos area exists on disk from S01, but its API is reserved
	// until the media stages (S01.2-T06). One hand-written route answers
	// every method and sub-path; the spec documents it under the photos tag.
	photos := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apperr.Write(w, r, o.Logger, apperr.New(apperr.NotAvailable, "the photos API is not available yet; it arrives with the media stages (S04)"))
	})
	return []Route{
		{Pattern: "GET /api/v1/system/health", Handler: http.HandlerFunc(g.GetHealth)},
		{Pattern: "GET /api/v1/files/items", Handler: http.HandlerFunc(g.GetItems)},
		{Pattern: "POST /api/v1/files/folders", Handler: strictJSON[gen.CreateFolderRequest](o.Logger, g.CreateFolder)},
		{Pattern: "POST /api/v1/files/operations/rename", Handler: strictJSON[gen.RenameRequest](o.Logger, g.RenameItem)},
		{Pattern: "POST /api/v1/files/operations/move", Handler: strictJSON[gen.MoveRequest](o.Logger, g.MoveItem)},
		{Pattern: "GET /api/v1/files/content", Handler: withRequest(g.DownloadFile)},
		{Pattern: "PUT /api/v1/files/content", Handler: declaredSize(o.MaxUploadBytes, o.Logger, g.UploadFile), MaxBody: o.MaxUploadBytes},
		{Pattern: "/api/v1/photos", Handler: photos},
		{Pattern: "/api/v1/photos/", Handler: photos},
	}
}

// noStore marks every API response as not cacheable by default: they all
// reflect the current state of the storage. A handler may override it.
func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// New builds the complete HTTP handler. From the outside in: request ID,
// access log, panic recovery, routes. Each route has its own body limit
// (uploads need far more than JSON), applied inside the mux: the access
// log keeps the request the mux fills in (its route pattern), while
// http.MaxBytesHandler passes a copy on.
func New(o Options) http.Handler {
	if o.Logger == nil {
		o.Logger = slog.New(slog.DiscardHandler)
	}
	if o.MaxBodyBytes <= 0 {
		o.MaxBodyBytes = DefaultMaxBodyBytes
	}
	mux := http.NewServeMux()
	for _, r := range Routes(o) {
		limit := r.MaxBody
		if limit <= 0 {
			limit = o.MaxBodyBytes
		}
		mux.Handle(r.Pattern, http.MaxBytesHandler(r.Handler, limit))
	}
	var h = problemsForUnmatched(mux, o.Logger)
	h = noStore(h)
	h = apperr.Recover(o.Logger)(h)
	h = logging.AccessLog(o.Logger)(h)
	return logging.RequestID(h)
}
