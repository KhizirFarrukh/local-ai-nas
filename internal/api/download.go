package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// requestKey carries the request to an operation whose response needs it:
// a download answers the request's Range and conditional headers.
type requestKey struct{}

// withRequest makes the request available to the operation.
func withRequest(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next(w, r.WithContext(context.WithValue(r.Context(), requestKey{}, r)))
	})
}

// DownloadFile sends a file (S01.3-T06).
func (s *server) DownloadFile(ctx context.Context, req gen.DownloadFileRequestObject) (gen.DownloadFileResponseObject, error) {
	r, ok := ctx.Value(requestKey{}).(*http.Request)
	if !ok {
		return nil, errors.New("the download route runs without withRequest")
	}
	it, body, err := s.files.Download(ctx, s.owner, req.Params.Path)
	if err != nil {
		return nil, err
	}
	return fileContent{r: r, log: s.log, item: it, body: body}, nil
}

// fileContent is a download's response. http.ServeContent answers with
// 200, 206, 304, 412, or 416, depending on the request's headers.
type fileContent struct {
	r    *http.Request
	log  *slog.Logger
	item files.Item
	body io.ReadSeekCloser
}

// VisitDownloadFileResponse implements gen.DownloadFileResponseObject.
func (c fileContent) VisitDownloadFileResponse(w http.ResponseWriter) error {
	defer func() { _ = c.body.Close() }() // read-only
	h := w.Header()
	mediaType := c.item.MIME
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	h.Set("Content-Type", mediaType) // set, so ServeContent does not sniff
	h.Set("ETag", c.item.ETag)
	h.Set("Content-Disposition", attachment(c.item.Name))
	// Cached copies must be revalidated with the ETag (304).
	h.Set("Cache-Control", "private, no-cache")
	// Never render a file as something else, or as part of this site: an
	// uploaded HTML or SVG file must not run scripts in the app's origin.
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Content-Security-Policy", "default-src 'none'; sandbox")
	r := c.r
	if c.item.Size == 0 && r.Header.Get("Range") != "" {
		// An empty file has no byte to range over. Go's ServeContent would
		// answer a suffix range ("bytes=-1") with 206 and the invalid
		// Content-Range "bytes 0--1/0"; RFC 9110 lets a server ignore
		// Range, so the whole (empty) file is sent (S01.7-T03).
		r = r.Clone(r.Context())
		r.Header.Del("Range")
	}
	http.ServeContent(&problemStatus{ResponseWriter: w, r: r, log: c.log}, r, "", c.item.ModTime, c.body)
	return nil
}

// problemStatus passes a response through, but turns the plain-text errors
// of http.ServeContent into problems, so every error of the API is one:
// 412 precondition_failed, 416 range_not_satisfiable (keeping its
// Content-Range header), and any other error status internal.
type problemStatus struct {
	http.ResponseWriter
	r      *http.Request
	log    *slog.Logger
	failed bool
}

func (p *problemStatus) WriteHeader(code int) {
	if code < http.StatusBadRequest || p.failed {
		p.ResponseWriter.WriteHeader(code)
		return
	}
	p.failed = true
	h := p.Header()
	h.Del("Content-Disposition")
	h.Set("Cache-Control", "no-store")
	var err error
	switch code {
	case http.StatusPreconditionFailed:
		err = apperr.New(apperr.PreconditionFailed, "the file does not match the request's If-Match or If-Unmodified-Since condition")
	case http.StatusRequestedRangeNotSatisfiable:
		err = apperr.Newf(apperr.RangeNotSatisfiable, "the requested range is outside the file (%s)", h.Get("Content-Range"))
	default:
		err = fmt.Errorf("serving the file failed with status %d", code)
	}
	apperr.Write(p.ResponseWriter, p.r, p.log, err)
}

// Write drops ServeContent's plain-text error body; the problem replaces
// it.
func (p *problemStatus) Write(b []byte) (int, error) {
	if p.failed {
		return len(b), nil
	}
	return p.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (p *problemStatus) Unwrap() http.ResponseWriter { return p.ResponseWriter }

// attachment returns a Content-Disposition value that makes browsers save
// the file as name (RFC 6266): an ASCII fallback in filename, and, when
// the name needs more than that, the exact UTF-8 name in filename*
// (RFC 8187), which current browsers prefer.
func attachment(name string) string {
	var fallback strings.Builder
	for _, r := range name {
		// Quotes and backslashes would need escaping, and some browsers
		// decode percent signs; control and non-ASCII characters are not
		// allowed in a quoted string.
		if r < 0x20 || r >= 0x7f || r == '"' || r == '\\' || r == '%' {
			r = '_'
		}
		fallback.WriteRune(r)
	}
	v := `attachment; filename="` + fallback.String() + `"`
	if fallback.String() != name {
		v += "; filename*=UTF-8''" + encodeExtValue(name)
	}
	return v
}

// encodeExtValue percent-encodes s as the value-chars of an RFC 8187
// ext-value: attr-char stays, every other byte becomes %XX.
func encodeExtValue(s string) string {
	const attrChar = "!#$&+-.^_`|~"
	var b strings.Builder
	for i := range len(s) {
		c := s[i]
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || strings.IndexByte(attrChar, c) >= 0 {
			b.WriteByte(c)
			continue
		}
		fmt.Fprintf(&b, "%%%02X", c)
	}
	return b.String()
}
