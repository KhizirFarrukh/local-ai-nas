package api

import (
	"log/slog"
	"net/http"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// problemsForUnmatched serves mux, but answers requests that match no
// route with problems instead of the mux's plain-text replies: 404
// not_found for an unknown path, and 405 method_not_allowed (with the
// Allow header) for a known path with another method (S01.5-T03).
func problemsForUnmatched(mux *http.ServeMux, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, pattern := mux.Handler(r)
		if pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}
		// Let the mux's own handler decide between 404 and 405 (and the
		// Allow header), then answer with a problem.
		probe := &probeWriter{header: http.Header{}}
		h.ServeHTTP(probe, r)
		if probe.status == http.StatusMethodNotAllowed {
			allow := probe.header.Get("Allow")
			w.Header().Set("Allow", allow)
			apperr.Write(w, r, log, apperr.Newf(apperr.MethodNotAllowed,
				"%s is not allowed for %s; allowed: %s", r.Method, r.URL.Path, allow))
			return
		}
		apperr.Write(w, r, log, apperr.Newf(apperr.NotFound, "there is no endpoint at %s", r.URL.Path))
	})
}

// probeWriter records the status and headers a handler sets and discards
// the body.
type probeWriter struct {
	header http.Header
	status int
}

func (p *probeWriter) Header() http.Header { return p.header }

func (p *probeWriter) Write(b []byte) (int, error) {
	if p.status == 0 {
		p.status = http.StatusOK
	}
	return len(b), nil
}

func (p *probeWriter) WriteHeader(code int) {
	if p.status == 0 {
		p.status = code
	}
}
