// Package uploads is the resumable upload endpoint (S01.4, ADR-0008): an
// embedded tus 1.0 server (tusd) that keeps unfinished uploads in the
// internal data, outside the storage areas, and records every upload
// session in the database. A finished upload becomes a file in the files
// area only through the files service (S01.4-T03), so a partial file is
// never visible there.
package uploads

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/tus/tusd/v2/pkg/filelocker"
	"github.com/tus/tusd/v2/pkg/filestore"
	tus "github.com/tus/tusd/v2/pkg/handler"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/logging"
)

// Metadata keys of an upload (docs/api/conventions.md). Clients send them
// base64-encoded in the Upload-Metadata header, as tus defines.
const (
	MetaTargetPath = "target_path"
	MetaOnConflict = "on_conflict"
	MetaSHA256     = "sha256"
)

// Target is where finished uploads go: the files service. Its checks run
// when an upload is created, so a refused upload stores nothing.
type Target interface {
	// CheckUpload runs every check of an upload to path that needs no
	// content, and writes nothing (files.Local.CheckUpload).
	CheckUpload(ctx context.Context, owner, path string, size int64, opts files.UploadOptions) error
	// CommitUpload makes the finished upload src the file at path, or
	// returns files.ErrNotSameDevice (files.Local.CommitUpload).
	CommitUpload(ctx context.Context, owner, path, src string, opts files.UploadOptions) (files.Item, bool, error)
	// Upload writes the file at path from a stream (the copy fallback).
	Upload(ctx context.Context, owner, path string, body io.Reader, size int64, opts files.UploadOptions) (files.Item, bool, error)
}

// DefaultExpiry is how long an unfinished upload is kept by default.
const DefaultExpiry = 24 * time.Hour

// Options configure New.
type Options struct {
	// Dir holds the unfinished uploads: <internal data>/tmp/uploads.
	Dir string
	// DB holds the index of upload sessions.
	DB *db.DB
	// Files checks and, when an upload is finished, receives the file.
	Files Target
	// Namespace owns every upload; S01 has one owner, S03 takes it from
	// the session.
	Namespace string
	// BasePath is the URL path the server is mounted at, with a trailing
	// slash, such as /api/v1/files/uploads/.
	BasePath string
	// MaxSize is the largest upload in bytes; 0 means no limit.
	MaxSize int64
	// Expiry is how long an unfinished upload is kept; 0 means
	// DefaultExpiry.
	Expiry time.Duration
	// Logger receives tusd's messages.
	Logger *slog.Logger
	// Now returns the current time; nil means time.Now.
	Now func() time.Time
}

// Server serves the tus protocol under Options.BasePath.
type Server struct {
	http.Handler
	index Index
	o     Options
}

// New returns the tus server. Unfinished uploads go to o.Dir through
// tusd's file store and file locker.
func New(o Options) (*Server, error) {
	if o.Expiry <= 0 {
		o.Expiry = DefaultExpiry
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Logger == nil {
		o.Logger = slog.New(slog.DiscardHandler)
	}
	s := &Server{index: NewIndex(o.DB), o: o}
	composer := tus.NewStoreComposer()
	filestore.New(o.Dir).UseIn(composer)
	filelocker.New(o.Dir).UseIn(composer)
	h, err := tus.NewHandler(tus.Config{
		StoreComposer:              composer,
		BasePath:                   o.BasePath,
		MaxSize:                    o.MaxSize,
		DisableDownload:            true, // files are downloaded from the files area, never from here
		DisableConcatenation:       true,
		Cors:                       &tus.CorsConfig{Disable: true}, // the web app is served from the same origin
		Logger:                     tusLogger(o.Logger),
		PreUploadCreateCallback:    s.beforeCreate,
		PreFinishResponseCallback:  s.beforeFinish,
		PreUploadTerminateCallback: s.beforeTerminate,
	})
	if err != nil {
		return nil, err
	}
	// tusd routes on the path below its base path.
	s.Handler = problemBodies(knownIDs(o.BasePath, o.Logger, http.StripPrefix(strings.TrimSuffix(o.BasePath, "/"), h)), o.Logger)
	return s, nil
}

// knownIDs answers 404 for a path whose upload ID this server cannot have
// made (isUploadID) before tusd sees it: tusd uses the ID as a file name,
// and a name such as `"` is an error on Windows, not "not found".
func knownIDs(base string, log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := strings.CutPrefix(r.URL.Path, base); ok && id != "" && !isUploadID(id) {
			w.Header().Set("Tus-Resumable", "1.0.0") // tus puts it on every answer
			apperr.Write(w, r, log, apperr.Newf(apperr.NotFound, "there is no upload %q", id))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// isUploadID reports whether id has the form of the IDs beforeCreate
// makes: 26 characters of lowercase base32.
func isUploadID(id string) bool {
	if len(id) != 26 {
		return false
	}
	for _, c := range id {
		if (c < 'a' || c > 'z') && (c < '2' || c > '7') {
			return false
		}
	}
	return true
}

// beforeCreate runs before tusd creates an upload. It checks the metadata
// the session needs, chooses the upload ID, and records the session, so
// every upload in the store has a row in the index.
func (s *Server) beforeCreate(ev tus.HookEvent) (tus.HTTPResponse, tus.FileInfoChanges, error) {
	ctx := ev.Context
	up := ev.Upload
	target := up.MetaData[MetaTargetPath]
	if !strings.HasPrefix(target, "/") {
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, refuse(ctx, s.o.Logger,
			apperr.Newf(apperr.InvalidRequest, "the upload metadata needs %s, the path of the new file starting with /", MetaTargetPath))
	}
	policy := up.MetaData[MetaOnConflict]
	switch policy {
	case "":
		policy = "fail"
	case "fail", "rename", "overwrite":
	default:
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, refuse(ctx, s.o.Logger,
			apperr.Newf(apperr.InvalidRequest, "%s must be fail, rename, or overwrite, got %q", MetaOnConflict, policy))
	}
	sum := strings.ToLower(up.MetaData[MetaSHA256])
	if sum != "" {
		if b, err := hex.DecodeString(sum); err != nil || len(b) != 32 {
			return tus.HTTPResponse{}, tus.FileInfoChanges{}, refuse(ctx, s.o.Logger,
				apperr.Newf(apperr.InvalidRequest, "%s must be 64 hexadecimal digits", MetaSHA256))
		}
	}
	if up.SizeIsDeferred {
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, refuse(ctx, s.o.Logger,
			apperr.New(apperr.LengthRequired, "an upload needs Upload-Length when it is created, to check the limits and the free space"))
	}
	err := s.o.Files.CheckUpload(ctx, s.o.Namespace, target, up.Size, files.UploadOptions{OnConflict: files.OnConflict(policy)})
	if err != nil {
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, refuse(ctx, s.o.Logger, err)
	}
	now := s.o.Now().UTC()
	id := strings.ToLower(rand.Text())
	err = s.index.Add(ctx, Session{
		ID: id, Namespace: s.o.Namespace, TargetPath: target, OnConflict: policy,
		DeclaredSize: up.Size, SHA256: sum, CreatedAt: now, ExpiresAt: now.Add(s.o.Expiry),
	})
	if err != nil {
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, refuse(ctx, s.o.Logger, err)
	}
	return tus.HTTPResponse{}, tus.FileInfoChanges{ID: id}, nil
}

// beforeTerminate runs before tusd deletes an upload the client cancels:
// its session goes away with it.
func (s *Server) beforeTerminate(ev tus.HookEvent) (tus.HTTPResponse, error) {
	if err := s.index.Remove(ev.Context, ev.Upload.ID); err != nil {
		return tus.HTTPResponse{}, refuse(ev.Context, s.o.Logger, err)
	}
	return tus.HTTPResponse{}, nil
}

// refuse turns err into a tusd error whose response is a problem
// (RFC 9457), so these refusals look like every other API error. Errors
// of the tus protocol itself (such as a wrong Upload-Offset) keep tusd's
// plain-text bodies.
func refuse(ctx context.Context, log *slog.Logger, err error) tus.Error {
	p := apperr.ProblemFor(err, logging.RequestIDFrom(ctx))
	if p.Code == apperr.Internal.Code() {
		log.ErrorContext(ctx, "upload hook failed", "error", err.Error())
	}
	body, jerr := json.Marshal(p)
	if jerr != nil {
		body = []byte(`{}`)
	}
	return tus.Error{
		ErrorCode: p.Code,
		Message:   p.Detail,
		HTTPResponse: tus.HTTPResponse{
			StatusCode: p.Status,
			Body:       string(body),
			Header:     tus.HTTPHeader{"Content-Type": apperr.ContentType},
		},
	}
}
