package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestIDHeader carries the request ID in requests and responses.
const RequestIDHeader = "X-Request-ID"

type ctxKey struct{}

// RequestID is middleware that gives every request an ID. A valid incoming
// X-Request-ID is kept (so a client or proxy can correlate its own logs);
// otherwise a random 128-bit hex ID is generated. The ID is set on the
// response header and stored in the request context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if !ValidRequestID(id) {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

// ValidRequestID reports whether an incoming request ID is acceptable:
// 1 to 128 characters from A-Z, a-z, 0-9, and "-", "_", ".", ":". This
// keeps IDs such as UUIDs and rejects anything that could break log lines.
func ValidRequestID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, c := range []byte(id) {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case c == '-' || c == '_' || c == '.' || c == ':':
		default:
			return false
		}
	}
	return true
}

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:]) // crypto/rand.Read never returns an error (Go 1.24+).
	return hex.EncodeToString(b[:])
}

// RequestIDFrom returns the request ID stored by RequestID, or "".
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// AccessLog is middleware that writes one "request" line per request with
// the method, route pattern, status, response bytes, duration, and request
// ID. Query values under sensitive names are redacted. At debug level the
// request headers are included, with credentials redacted. Place it inside
// RequestID and outside the ServeMux, so the matched pattern is known.
func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &recorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)
			if rec.status == 0 {
				rec.status = http.StatusOK
			}
			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("route", r.Pattern),
				slog.Int("status", rec.status),
				slog.Int64("bytes", rec.bytes),
				slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
				slog.String("request_id", RequestIDFrom(r.Context())),
				slog.String("remote", r.RemoteAddr),
			}
			if r.URL.RawQuery != "" {
				attrs = append(attrs, slog.String("query", RedactQuery(r.URL.Query()).Encode()))
			}
			ctx := r.Context()
			if logger.Enabled(ctx, slog.LevelDebug) {
				attrs = append(attrs, slog.Any("headers", RedactHeaders(r.Header)))
			}
			logger.LogAttrs(ctx, slog.LevelInfo, "request", attrs...)
		})
	}
}

// Redacted replaces secret values in logs.
const Redacted = "[REDACTED]"

var sensitiveHeaders = map[string]bool{
	"Authorization":       true,
	"Proxy-Authorization": true,
	"Cookie":              true,
	"Set-Cookie":          true,
	"X-Api-Key":           true,
	"X-Auth-Token":        true,
	"X-Csrf-Token":        true,
}

// RedactHeaders returns a copy of h with credential headers replaced by
// Redacted.
func RedactHeaders(h http.Header) http.Header {
	out := h.Clone()
	for name := range out {
		if sensitiveHeaders[http.CanonicalHeaderKey(name)] {
			out[name] = []string{Redacted}
		}
	}
	return out
}

// sensitiveQueryWords marks query parameter names whose values are
// redacted when the name contains one of them (case-insensitive).
var sensitiveQueryWords = []string{"token", "password", "passwd", "secret", "key", "signature", "sig", "auth", "session", "code"}

// RedactQuery returns a copy of q with the values of sensitive parameters
// replaced by Redacted.
func RedactQuery(q url.Values) url.Values {
	out := make(url.Values, len(q))
	for name, values := range q {
		if isSensitiveParam(name) {
			out[name] = []string{Redacted}
			continue
		}
		out[name] = append([]string(nil), values...)
	}
	return out
}

func isSensitiveParam(name string) bool {
	n := strings.ToLower(name)
	for _, w := range sensitiveQueryWords {
		if strings.Contains(n, w) {
			return true
		}
	}
	return false
}

// recorder captures the status and body size of a response.
type recorder struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (r *recorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(p)
	r.bytes += int64(n)
	return n, err
}

// Unwrap lets http.ResponseController reach the underlying writer (for
// Flush, deadlines, and so on).
func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// Flush supports handlers that assert http.Flusher directly.
func (r *recorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		if r.status == 0 {
			r.status = http.StatusOK
		}
		f.Flush()
	}
}
