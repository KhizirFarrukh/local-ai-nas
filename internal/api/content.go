package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// contentLengthKey carries the declared size of an upload from
// declaredSize to the upload operation.
type contentLengthKey struct{}

// declaredSize runs before an upload operation. The upload must declare
// its size (Content-Length), and the size must be within limit, so both
// limits and free space are checked before a byte is read: a client that
// waits for "100 Continue" never sends a refused body.
func declaredSize(limit int64, log *slog.Logger, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.ContentLength < 0:
			apperr.Write(w, r, log, apperr.New(apperr.LengthRequired,
				"an upload needs a Content-Length header; for content of unknown size use the resumable upload"))
			return
		case r.ContentLength > limit:
			apperr.Write(w, r, log, apperr.Newf(apperr.TooLarge,
				"the file has %d bytes; the largest upload is %d bytes", r.ContentLength, limit))
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), contentLengthKey{}, r.ContentLength)))
	})
}

// chunkLimit refuses a tus request whose declared body is over limit
// (uploads.max_chunk_size) with 413 before tusd reads any of it, so no
// data is stored (S01.4-T05). A body without Content-Length is cut at the
// limit by the route's body limit instead. Every answer names the limit
// in MaxChunkHeader, so clients can size their requests (S02.4-T01).
func chunkLimit(limit int64, log *slog.Logger, next http.Handler) http.Handler {
	value := strconv.FormatInt(limit, 10)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(MaxChunkHeader, value)
		if r.ContentLength > limit {
			apperr.Write(w, r, log, apperr.Newf(apperr.TooLarge,
				"one upload request may carry at most %d bytes (uploads.max_chunk_size), this one has %d", limit, r.ContentLength))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UploadFile stores the request body as a file (S01.3-T05): 201 when a
// new file was created, 200 when on_conflict=overwrite replaced one.
func (s *server) UploadFile(ctx context.Context, req gen.UploadFileRequestObject) (gen.UploadFileResponseObject, error) {
	size, ok := ctx.Value(contentLengthKey{}).(int64)
	if !ok {
		return nil, errors.New("the upload route runs without declaredSize")
	}
	policy, err := conflictPolicy(req.Params.OnConflict)
	if err != nil {
		return nil, err
	}
	it, created, err := s.files.Upload(ctx, s.owner, req.Params.Path, req.Body, size, files.UploadOptions{OnConflict: policy})
	if err != nil {
		return nil, err
	}
	if created {
		return gen.UploadFile201JSONResponse(fileItem(it)), nil
	}
	return gen.UploadFile200JSONResponse(fileItem(it)), nil
}

// conflictPolicy checks an optional on_conflict value; the generated
// binding does not check enumerations.
func conflictPolicy(p *gen.OnConflict) (files.OnConflict, error) {
	if p == nil {
		return "", nil
	}
	if !p.Valid() {
		return "", apperr.Newf(apperr.InvalidRequest, "on_conflict must be fail, rename, or overwrite, got %q", *p)
	}
	return files.OnConflict(*p), nil
}
