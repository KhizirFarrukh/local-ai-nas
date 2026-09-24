package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// strictJSON checks a JSON request body before the generated handler
// decodes it (docs/api/conventions.md): the content type must be JSON, the
// body must decode into T without unknown fields, and nothing may follow
// the object. The generated decoder is lenient on all three. The body is
// then handed on unchanged. Its size is already limited by the API's body
// limit.
func strictJSON[T any](log *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
			apperr.Write(w, r, log, apperr.New(apperr.InvalidRequest, "the request body must be JSON (Content-Type: application/json)"))
			return
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				apperr.Write(w, r, log, apperr.Wrap(apperr.TooLarge, "the request body is too large", err))
				return
			}
			apperr.Write(w, r, log, apperr.Wrap(apperr.InvalidRequest, "the request body could not be read", err))
			return
		}
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		var v T
		if err := dec.Decode(&v); err != nil {
			apperr.Write(w, r, log, apperr.Wrap(apperr.InvalidRequest, "the request body is not valid: "+err.Error(), err))
			return
		}
		if dec.More() {
			apperr.Write(w, r, log, apperr.New(apperr.InvalidRequest, "the request body must be a single JSON object"))
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(data))
		next(w, r)
	}
}
