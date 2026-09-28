package uploads

import (
	"context"
	"crypto/sha256"
	"encoding"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/tus/tusd/v2/pkg/filestore"
	tus "github.com/tus/tusd/v2/pkg/handler"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// The content hash of a resumable upload (S01.4-T07, FR-211, ADR-0021) is
// computed while its chunks arrive, so the finished file is not read
// again. Chunks can arrive in several requests, even across restarts, so
// the SHA-256 state is saved next to the upload after every chunk: the
// .hash file holds how many bytes it covers and the state. When the state
// is missing or does not cover exactly the bytes stored (an upload from
// before this change, a failed write), the finished file is hashed once.

// hashState is the content of an upload's .hash file.
type hashState struct {
	Offset int64  `json:"offset"` // the bytes the state covers
	State  []byte `json:"state"`  // the SHA-256 state (MarshalBinary)
}

// hashPath is where an upload's hash state is kept.
func (s *Server) hashPath(id string) string { return filepath.Join(s.o.Dir, id+".hash") }

// hashingStore is tusd's file store with every upload hashed as it is
// written. It registers every part the file store registers, so tusd
// behaves as before: concatenation and downloads stay switched off in the
// handler's configuration and are refused, not ignored.
type hashingStore struct {
	filestore.FileStore
	s *Server
}

func (h hashingStore) useIn(c *tus.StoreComposer) {
	c.UseCore(h)
	c.UseTerminater(h)
	c.UseConcater(h)
	c.UseLengthDeferrer(h)
	c.UseContentServer(h)
}

// NewUpload implements tus.DataStore.
func (h hashingStore) NewUpload(ctx context.Context, info tus.FileInfo) (tus.Upload, error) {
	up, err := h.FileStore.NewUpload(ctx, info)
	if err != nil {
		return nil, err
	}
	return hashingUpload{Upload: up, s: h.s}, nil
}

// GetUpload implements tus.DataStore.
func (h hashingStore) GetUpload(ctx context.Context, id string) (tus.Upload, error) {
	up, err := h.FileStore.GetUpload(ctx, id)
	if err != nil {
		return nil, err
	}
	return hashingUpload{Upload: up, s: h.s}, nil
}

// AsTerminatableUpload implements tus.TerminaterDataStore.
func (h hashingStore) AsTerminatableUpload(up tus.Upload) tus.TerminatableUpload {
	return h.FileStore.AsTerminatableUpload(unwrap(up))
}

// AsLengthDeclarableUpload implements tus.LengthDeferrerDataStore.
func (h hashingStore) AsLengthDeclarableUpload(up tus.Upload) tus.LengthDeclarableUpload {
	return h.FileStore.AsLengthDeclarableUpload(unwrap(up))
}

// AsConcatableUpload implements tus.ConcaterDataStore.
func (h hashingStore) AsConcatableUpload(up tus.Upload) tus.ConcatableUpload {
	return h.FileStore.AsConcatableUpload(unwrap(up))
}

// AsServableUpload implements tus.ContentServerDataStore.
func (h hashingStore) AsServableUpload(up tus.Upload) tus.ServableUpload {
	return h.FileStore.AsServableUpload(unwrap(up))
}

func unwrap(up tus.Upload) tus.Upload {
	if w, ok := up.(hashingUpload); ok {
		return w.Upload
	}
	return up
}

// hashingUpload hashes what WriteChunk stores.
type hashingUpload struct {
	tus.Upload
	s *Server
}

// WriteChunk implements tus.Upload. The state goes on only when it covers
// exactly the bytes before offset and every byte read was stored;
// otherwise the .hash file is removed and the finish hashes the file.
func (u hashingUpload) WriteChunk(ctx context.Context, offset int64, src io.Reader) (int64, error) {
	info, err := u.GetInfo(ctx)
	if err != nil {
		return 0, err
	}
	h := u.s.resumeHash(info.ID, offset)
	if h == nil {
		return u.Upload.WriteChunk(ctx, offset, src)
	}
	counted := &countingWriter{w: h}
	n, err := u.Upload.WriteChunk(ctx, offset, io.TeeReader(src, counted))
	if counted.n != n {
		u.s.dropHash(info.ID) // read more than was stored: the state is wrong
	} else if n > 0 || offset == 0 {
		u.s.saveHash(info.ID, offset+n, h)
	}
	return n, err
}

// resumeHash returns the hash to continue at offset, or nil when there is
// no usable state (the finish then reads the file).
func (s *Server) resumeHash(id string, offset int64) hash.Hash {
	h := sha256.New()
	if offset == 0 {
		return h
	}
	b, err := os.ReadFile(s.hashPath(id)) // #nosec G304 -- a file in the upload directory
	if err != nil {
		return nil
	}
	var st hashState
	if json.Unmarshal(b, &st) != nil || st.Offset != offset {
		return nil
	}
	if h.(encoding.BinaryUnmarshaler).UnmarshalBinary(st.State) != nil {
		return nil
	}
	return h
}

// saveHash writes the state after a chunk, replacing the .hash file in one
// step. A failure only means the finish reads the file.
func (s *Server) saveHash(id string, offset int64, h hash.Hash) {
	state, err := h.(encoding.BinaryMarshaler).MarshalBinary()
	if err == nil {
		var b []byte
		if b, err = json.Marshal(hashState{Offset: offset, State: state}); err == nil {
			err = replaceFile(s.hashPath(id), b)
		}
	}
	if err != nil {
		s.o.Logger.Warn("saving an upload's hash state failed; it is hashed when it finishes", "upload", id, "error", err.Error())
		s.dropHash(id)
	}
}

func (s *Server) dropHash(id string) {
	if err := os.Remove(s.hashPath(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
		s.o.Logger.Warn("removing an upload's hash state failed", "upload", id, "error", err.Error())
	}
}

// contentHash returns the hash of the finished upload of size bytes: from
// the saved state when it covers them all, otherwise by reading the file.
func (s *Server) contentHash(id, data string, size int64) (string, error) {
	if h := s.resumeHash(id, size); h != nil {
		return files.HashPrefix + hex.EncodeToString(h.Sum(nil)), nil
	}
	f, err := os.Open(data) // #nosec G304 -- the path of the upload in tusd's store
	if err != nil {
		return "", apperr.Wrap(apperr.Internal, "opening the upload failed", err)
	}
	defer func() { _ = f.Close() }() // read-only
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", apperr.Wrap(apperr.Internal, "reading the upload failed", err)
	}
	return files.HashPrefix + hex.EncodeToString(h.Sum(nil)), nil
}

// countingWriter counts what passes to w.
type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}
