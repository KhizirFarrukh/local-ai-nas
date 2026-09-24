package uploads

import (
	"bytes"
	"log/slog"
	"net/http"
	"strings"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// statusKinds maps the error statuses tusd uses to the problem kinds with
// the same status, so a converted error keeps the status tus clients act
// on (for example 409 to ask for the offset again, 423 to retry).
var statusKinds = map[int]apperr.Kind{
	http.StatusBadRequest:            apperr.InvalidRequest,
	http.StatusNotFound:              apperr.NotFound,
	http.StatusMethodNotAllowed:      apperr.MethodNotAllowed,
	http.StatusConflict:              apperr.Conflict,
	http.StatusPreconditionFailed:    apperr.PreconditionFailed,
	http.StatusRequestEntityTooLarge: apperr.TooLarge,
	http.StatusLocked:                apperr.Locked,
	http.StatusInternalServerError:   apperr.Internal,
	http.StatusNotImplemented:        apperr.NotAvailable,
	http.StatusServiceUnavailable:    apperr.Unavailable,
}

// problemBodies makes every error of the tus server a problem, like every
// other API error: tusd answers protocol errors (a missing Tus-Resumable
// header, a wrong Upload-Offset) with plain text. The status and the
// headers stay as tusd set them; tusd's message becomes the detail.
func problemBodies(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pw := &problemWriter{ResponseWriter: w}
		next.ServeHTTP(pw, r)
		if pw.status == 0 {
			return
		}
		kind := statusKinds[pw.status]
		detail := strings.TrimSpace(pw.body.String())
		// tusd's message starts with its own code, such as
		// "ERR_UPLOAD_NOT_FOUND: upload not found".
		if _, msg, ok := strings.Cut(detail, ": "); ok && strings.HasPrefix(detail, "ERR_") {
			detail = msg
		}
		if detail == "" {
			detail = http.StatusText(pw.status)
		}
		if kind == apperr.Internal {
			log.ErrorContext(r.Context(), "the tus server failed", "status", pw.status, "message", detail)
		}
		w.Header().Del("Content-Length") // it was set for tusd's text
		apperr.Write(w, r, log, apperr.New(kind, detail))
	})
}

// problemWriter holds back an error response that is not a problem yet;
// everything else passes through.
type problemWriter struct {
	http.ResponseWriter
	status int // an error status being converted, or 0
	body   bytes.Buffer
}

func (p *problemWriter) WriteHeader(code int) {
	_, known := statusKinds[code]
	if code < http.StatusBadRequest || !known || p.Header().Get("Content-Type") == apperr.ContentType {
		p.ResponseWriter.WriteHeader(code)
		return
	}
	p.status = code
}

func (p *problemWriter) Write(b []byte) (int, error) {
	if p.status != 0 {
		if p.body.Len() < 4096 {
			p.body.Write(b)
		}
		return len(b), nil
	}
	return p.ResponseWriter.Write(b)
}

// Unwrap lets tusd reach the connection (http.ResponseController), which
// it uses for read deadlines while it receives a body.
func (p *problemWriter) Unwrap() http.ResponseWriter { return p.ResponseWriter }
