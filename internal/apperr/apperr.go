// Package apperr defines the typed domain errors and turns them into RFC
// 9457 problem details (application/problem+json) with stable codes
// (S01.1-T09). Handlers return errors; only this package decides what a
// client sees, so internal details never leak into responses.
package apperr

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/KhizirFarrukh/local-ai-nas/internal/logging"
)

// Kind classifies an error. Each kind has a stable code, which clients may
// rely on, and an HTTP status.
type Kind int

// The kinds. New kinds are added at the end, with a row in the kinds table
// and in the error catalogue (docs/api/errors.md, S01.5-T03).
const (
	// Internal is an unexpected failure. Clients see a generic message.
	Internal Kind = iota
	// InvalidRequest is a malformed request: bad parameters or body.
	InvalidRequest
	// InvalidName is a file or folder name that breaks the name rules.
	InvalidName
	// OutsideRoot is a path that would leave the user's namespace.
	OutsideRoot
	// NotFound is a missing file, folder, or upload.
	NotFound
	// Conflict is a clash with the current state, such as an existing
	// target with on_conflict=fail.
	Conflict
	// TooLarge is an upload or request over a configured size limit.
	TooLarge
	// InsufficientStorage means the write would use the free-space reserve.
	InsufficientStorage
	// NotAvailable is a feature that exists in the API but is not
	// available yet, such as /api/v1/photos in S01.
	NotAvailable
	// MethodNotAllowed is a known endpoint called with a method it does
	// not support.
	MethodNotAllowed
)

type kindInfo struct {
	code   string
	status int
}

var kinds = map[Kind]kindInfo{
	Internal:            {"internal", http.StatusInternalServerError},
	InvalidRequest:      {"invalid_request", http.StatusBadRequest},
	InvalidName:         {"invalid_name", http.StatusBadRequest},
	OutsideRoot:         {"outside_root", http.StatusBadRequest},
	NotFound:            {"not_found", http.StatusNotFound},
	Conflict:            {"conflict", http.StatusConflict},
	TooLarge:            {"too_large", http.StatusRequestEntityTooLarge},
	InsufficientStorage: {"insufficient_storage", http.StatusInsufficientStorage},
	NotAvailable:        {"not_available", http.StatusNotImplemented},
	MethodNotAllowed:    {"method_not_allowed", http.StatusMethodNotAllowed},
}

// Kinds returns every defined kind, in order.
func Kinds() []Kind {
	out := make([]Kind, 0, len(kinds))
	for k := Internal; ; k++ {
		if _, ok := kinds[k]; !ok {
			return out
		}
		out = append(out, k)
	}
}

func (k Kind) info() kindInfo {
	if i, ok := kinds[k]; ok {
		return i
	}
	return kinds[Internal]
}

// Code is the stable, machine-readable code, such as "not_found".
func (k Kind) Code() string { return k.info().code }

// Status is the HTTP status code.
func (k Kind) Status() int { return k.info().status }

func (k Kind) String() string { return k.Code() }

// Error is a domain error: a kind, a detail that is safe to show to the
// client, and an optional cause that is only logged.
type Error struct {
	Kind   Kind
	Detail string
	// Rule optionally names the exact rule that was broken, such as
	// "reserved_name" for an invalid_name error (S01.6-T02). Clients get
	// it in the problem's "rule" field.
	Rule string
	Err  error
}

// New returns an error of the given kind with a client-safe detail.
func New(kind Kind, detail string) *Error {
	return &Error{Kind: kind, Detail: detail}
}

// NewRule returns an error of the given kind for a broken rule, with a
// client-safe detail.
func NewRule(kind Kind, rule, detail string) *Error {
	return &Error{Kind: kind, Rule: rule, Detail: detail}
}

// Newf is New with a formatted detail.
func Newf(kind Kind, format string, args ...any) *Error {
	return &Error{Kind: kind, Detail: fmt.Sprintf(format, args...)}
}

// Wrap returns an error of the given kind with a client-safe detail and a
// cause. The cause is logged but never sent to the client.
func Wrap(kind Kind, detail string, cause error) *Error {
	return &Error{Kind: kind, Detail: detail, Err: cause}
}

func (e *Error) Error() string {
	msg := e.Kind.Code()
	if e.Detail != "" {
		msg += ": " + e.Detail
	}
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}

func (e *Error) Unwrap() error { return e.Err }

// KindOf returns the kind of the first *Error in err's chain, or Internal
// if there is none.
func KindOf(err error) Kind {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return Internal
}

// Problem is an RFC 9457 problem details object. The type is "about:blank",
// so the title is the HTTP status text; the "code" extension carries the
// stable error code, and "correlation_id" the request ID from the logs.
type Problem struct {
	Type          string `json:"type"`
	Title         string `json:"title"`
	Status        int    `json:"status"`
	Detail        string `json:"detail,omitempty"`
	Code          string `json:"code"`
	Rule          string `json:"rule,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

// ContentType is the media type of problem responses.
const ContentType = "application/problem+json"

// genericDetail is shown for internal errors instead of their details.
const genericDetail = "An unexpected error occurred. The server log has the details under this correlation ID."

// ProblemFor builds the problem for err. Internal errors get a generic
// detail, whatever their message says.
func ProblemFor(err error, correlationID string) Problem {
	kind := KindOf(err)
	detail, rule := genericDetail, ""
	var e *Error
	if kind != Internal && errors.As(err, &e) {
		detail, rule = e.Detail, e.Rule
	}
	return Problem{
		Type:          "about:blank",
		Title:         http.StatusText(kind.Status()),
		Status:        kind.Status(),
		Detail:        detail,
		Code:          kind.Code(),
		Rule:          rule,
		CorrelationID: correlationID,
	}
}

// Write sends err as a problem response and logs it: server errors (5xx)
// at error level with the full error, client errors at debug level.
func Write(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	if err == nil {
		err = errors.New("apperr.Write called with a nil error")
	}
	id := logging.RequestIDFrom(r.Context())
	p := ProblemFor(err, id)
	level := slog.LevelDebug
	if p.Status >= 500 {
		level = slog.LevelError
	}
	logger.LogAttrs(r.Context(), level, "request failed",
		slog.String("code", p.Code), slog.Int("status", p.Status),
		slog.String("error", err.Error()), slog.String("request_id", id))
	writeProblem(w, p)
}

func writeProblem(w http.ResponseWriter, p Problem) {
	body, _ := json.Marshal(p) // A Problem of strings and ints always marshals.
	h := w.Header()
	h.Set("Content-Type", ContentType)
	h.Set("Cache-Control", "no-store")
	h.Del("Content-Length")
	w.WriteHeader(p.Status)
	_, _ = w.Write(append(body, '\n'))
}

// Recover is middleware that turns a handler panic into a generic 500
// problem. The panic value and stack trace are logged, never sent.
// http.ErrAbortHandler is re-raised, as net/http expects. Place it inside
// the access log, so the 500 is recorded there too.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				if v == http.ErrAbortHandler {
					panic(v)
				}
				id := logging.RequestIDFrom(r.Context())
				logger.LogAttrs(r.Context(), slog.LevelError, "handler panic",
					slog.Any("panic", v), slog.String("stack", string(debug.Stack())),
					slog.String("request_id", id))
				writeProblem(w, ProblemFor(errors.New("panic"), id))
			}()
			next.ServeHTTP(w, r)
		})
	}
}
