package uploads

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"

	tus "github.com/tus/tusd/v2/pkg/handler"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// ItemPathHeader carries the path of the finished file in the answer to
// the request that completes an upload. With on_conflict=rename it can
// differ from target_path.
const ItemPathHeader = "Item-Path"

// finalizeFault is called at the finalize stages "verified" (the data is
// complete and checked, nothing is committed yet) and "committed" (the
// file is in the files area). Tests make it fail to stand in for a crash.
var finalizeFault = func(stage string) error { return nil }

// beforeFinish runs when the last byte of an upload has arrived, before
// the answer: it makes the upload a file in the files area (S01.4-T03).
// If that fails, the upload is removed: it is complete, and there is no
// way to finalize it again with other settings.
func (s *Server) beforeFinish(ev tus.HookEvent) (tus.HTTPResponse, error) {
	it, err := s.finalize(ev.Context, ev.Upload)
	if err != nil {
		s.remove(ev.Context, ev.Upload)
		return tus.HTTPResponse{}, refuse(ev.Context, s.o.Logger, err)
	}
	return tus.HTTPResponse{Header: tus.HTTPHeader{ItemPathHeader: "/" + it.RelPath}}, nil
}

func (s *Server) finalize(ctx context.Context, up tus.FileInfo) (files.Item, error) {
	sess, err := s.index.Get(ctx, up.ID)
	if err != nil {
		return files.Item{}, apperr.Wrap(apperr.Internal, "the upload has no session", err)
	}
	data := up.Storage["Path"]
	if err := syncAndVerify(data, sess.SHA256); err != nil {
		return files.Item{}, err
	}
	if err := finalizeFault("verified"); err != nil {
		return files.Item{}, err
	}
	opts := files.UploadOptions{OnConflict: files.OnConflict(sess.OnConflict)}
	it, _, err := s.o.Files.CommitUpload(ctx, sess.Namespace, sess.TargetPath, data, opts)
	if errors.Is(err, files.ErrNotSameDevice) {
		// The upload directory is on another file system than the area
		// (the health check warns about this): copy instead of renaming.
		it, err = s.copyIn(ctx, sess, data, opts)
	}
	if err != nil {
		return files.Item{}, err
	}
	if err := finalizeFault("committed"); err != nil {
		return files.Item{}, err
	}
	s.remove(ctx, up) // the data has moved; the .info file and the session go
	return it, nil
}

// syncAndVerify makes the upload's data durable and, when the metadata
// gave one, checks its SHA-256.
func syncAndVerify(path, wantSHA256 string) (err error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0) // #nosec G304 -- the path of the upload in tusd's store
	if err != nil {
		return apperr.Wrap(apperr.Internal, "opening the upload failed", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = apperr.Wrap(apperr.Internal, "closing the upload failed", cerr)
		}
	}()
	if err := f.Sync(); err != nil {
		return apperr.Wrap(apperr.Internal, "syncing the upload failed", err)
	}
	if wantSHA256 == "" {
		return nil
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return apperr.Wrap(apperr.Internal, "reading the upload failed", err)
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != wantSHA256 {
		return apperr.Newf(apperr.InvalidRequest, "the upload's SHA-256 is %s, but its metadata says %s; the upload was removed", got, wantSHA256)
	}
	return nil
}

// copyIn copies the upload into the area with the simple upload (copy,
// fsync, commit), for an upload directory on another file system.
func (s *Server) copyIn(ctx context.Context, sess Session, data string, opts files.UploadOptions) (files.Item, error) {
	f, err := os.Open(data) // #nosec G304 -- the path of the upload in tusd's store
	if err != nil {
		return files.Item{}, apperr.Wrap(apperr.Internal, "opening the upload failed", err)
	}
	defer func() { _ = f.Close() }() // read-only
	it, _, err := s.o.Files.Upload(ctx, sess.Namespace, sess.TargetPath, f, sess.DeclaredSize, opts)
	return it, err
}

// remove deletes what is left of an upload in the store and its session.
// Leftovers after a failure here are removed by the expiry cleanup
// (S01.4-T06).
func (s *Server) remove(ctx context.Context, up tus.FileInfo) {
	for _, p := range []string{up.Storage["Path"], up.Storage["InfoPath"]} {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			s.o.Logger.WarnContext(ctx, "removing an upload file failed", "upload", up.ID, "error", err.Error())
		}
	}
	if err := s.index.Remove(ctx, up.ID); err != nil {
		s.o.Logger.WarnContext(ctx, "removing an upload session failed", "upload", up.ID, "error", err.Error())
	}
}
