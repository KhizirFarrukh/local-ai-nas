package api

import (
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// Protection against other websites before login (S03.5-T02, ADR-0042,
// FR-350): a web page the user visits can make the browser send requests
// to the NAS, directly (a cross-site form post) or through DNS rebinding,
// where the page's own host name resolves to this server so its requests
// look same-origin. Two checks stop both, and they stay after login:
//
//   - The Host header must name this server: a loopback address,
//     "localhost", or a name the admin configured. A rebinding page sends
//     its own host name, so it gets 421 misdirected_request.
//   - A state-changing request that carries an Origin must come from this
//     server's own origin; another website's (or "null") gets 403
//     csrf_failed. Browsers always send Origin on such requests, so a
//     request without one comes from a script or another program, which
//     the check lets through.
//
// The web app sends Referrer-Policy: same-origin (internal/webapp), under
// which browsers put the real origin, not "null", on its own requests.

// localOrigin returns the middleware. hosts are extra host names the
// server answers to, besides loopback addresses and "localhost".
func localOrigin(log *slog.Logger, hosts []string) func(http.Handler) http.Handler {
	names := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		names[normalizeHost(h)] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !hostAllowed(r.Host, names) {
				log.WarnContext(r.Context(), "request refused: the Host header does not name this server",
					"host", truncate(r.Host), "remote", r.RemoteAddr)
				apperr.Write(w, r, log, apperr.New(apperr.Misdirected,
					"the Host header does not name this server; open it by its own address"))
				return
			}
			if !safeMethod(r.Method) && !sameOrigin(r) {
				log.WarnContext(r.Context(), "request refused: it comes from another website",
					"origin", truncate(r.Header.Get("Origin")), "remote", r.RemoteAddr)
				apperr.Write(w, r, log, apperr.New(apperr.CSRFFailed,
					"the request comes from another website: its Origin is not this server's"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// hostAllowed reports whether a Host header (host or host:port) names
// this server. The port is not checked: a rebinding page cannot choose
// the host name the browser sends, whatever the port.
func hostAllowed(host string, names map[string]bool) bool {
	h := normalizeHost(host)
	if h == "" {
		return false
	}
	if h == "localhost" || names[h] {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}

// normalizeHost returns the host name of a Host header value in lower
// case, without the port, the IPv6 brackets, or a trailing dot.
func normalizeHost(host string) string {
	h := host
	if hh, _, err := net.SplitHostPort(host); err == nil {
		h = hh
	}
	h = strings.TrimSuffix(strings.TrimPrefix(h, "["), "]")
	return strings.TrimSuffix(strings.ToLower(h), ".")
}

// safeMethod reports whether a method never changes state.
func safeMethod(m string) bool {
	return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
}

// sameOrigin reports whether a request's Origin, if it has one, is this
// server's own origin: the request's scheme and Host.
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // not a browser: browsers always send Origin here
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return strings.EqualFold(origin, scheme+"://"+r.Host)
}

// truncate shortens a header value for a log line.
func truncate(s string) string {
	const limit = 200
	if len(s) > limit {
		return s[:limit] + "…"
	}
	return s
}
