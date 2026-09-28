package uploads

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// hashRecorder remembers the content hash each finished upload was
// committed with.
type hashRecorder struct {
	Target
	mu     sync.Mutex
	hashes map[string]string
}

func (r *hashRecorder) CommitUpload(ctx context.Context, owner, path, src string, opts files.UploadOptions) (files.Item, bool, error) {
	r.mu.Lock()
	r.hashes[path] = opts.ContentHash
	r.mu.Unlock()
	return r.Target.CommitUpload(ctx, owner, path, src, opts)
}

func (r *hashRecorder) get(path string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.hashes[path]
}

func sha(b []byte) string {
	d := sha256.Sum256(b)
	return files.HashPrefix + hex.EncodeToString(d[:])
}

func randomData(n int, seed uint64) []byte {
	rng := rand.New(rand.NewPCG(seed, seed+1))
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(rng.Uint32())
	}
	return b
}

// savedOffset reads how many bytes the upload's .hash state covers (-1:
// no state).
func savedOffset(t *testing.T, dir, id string) int64 {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, id+".hash"))
	if err != nil {
		return -1
	}
	var st hashState
	if err := json.Unmarshal(b, &st); err != nil {
		t.Fatal(err)
	}
	return st.Offset
}

func patch(t *testing.T, location string, offset int, data []byte) *http.Response {
	t.Helper()
	resp, err := tusRequest(t, t.Context(), http.MethodPatch, location, map[string]string{
		"Upload-Offset": strconv.Itoa(offset), "Content-Type": "application/offset+octet-stream",
	}, bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	return resp
}

// TestHashAcrossChunks is S01.4-T07's main case: an upload sent in
// several requests, one cut off, keeps its hash state after each chunk,
// and is committed with the hash of all its bytes.
func TestHashAcrossChunks(t *testing.T) {
	var rec *hashRecorder
	e := newTestEnv(t, func(t Target) Target {
		rec = &hashRecorder{Target: t, hashes: map[string]string{}}
		return rec
	}, nil)
	data := randomData(3<<20+333, 11)
	loc := createUpload(t, e.url, data, MetaTargetPath, "/docs/chunks.bin")
	id := loc[strings.LastIndex(loc, "/")+1:]

	// Two whole requests.
	for _, end := range []int{1000, 1<<20 + 5} {
		from := int(savedOffsetOrZero(t, e.dir, id))
		if resp := patch(t, loc, from, data[from:end]); resp.StatusCode != http.StatusNoContent {
			t.Fatalf("chunk to %d: status %d", end, resp.StatusCode)
		}
		if got := savedOffset(t, e.dir, id); got != int64(end) {
			t.Errorf("after the chunk to %d the state covers %d bytes", end, got)
		}
	}
	// One cut off mid-way: the state covers exactly what the server kept.
	sendThenCut(t, loc, data, 1<<20+5, 2<<20)
	offset := headOffset(t, loc)
	if got := savedOffset(t, e.dir, id); got != int64(offset) {
		t.Errorf("after the cut the server has %d bytes, the state covers %d", offset, got)
	}
	// The rest.
	resp := patch(t, loc, offset, data[offset:])
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("last chunk: status %d", resp.StatusCode)
	}
	if got, want := rec.get("/docs/chunks.bin"), sha(data); got != want {
		t.Errorf("committed with hash %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(e.dir, id+".hash")); !os.IsNotExist(err) {
		t.Errorf("the .hash file is left after the finish: %v", err)
	}
}

func savedOffsetOrZero(t *testing.T, dir, id string) int64 {
	t.Helper()
	if n := savedOffset(t, dir, id); n > 0 {
		return n
	}
	return 0
}

// TestHashFallback: without a usable state (missing, broken, or for other
// bytes), the finished upload is hashed from the file, and the hash is
// still right.
func TestHashFallback(t *testing.T) {
	for _, tc := range []struct {
		name  string
		spoil func(path string) error
	}{
		{"missing", os.Remove},
		{"broken", func(p string) error { return os.WriteFile(p, []byte("{"), 0o600) }},
		{"other offset", func(p string) error {
			return os.WriteFile(p, []byte(`{"offset":7,"state":""}`), 0o600)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var rec *hashRecorder
			e := newTestEnv(t, func(t Target) Target {
				rec = &hashRecorder{Target: t, hashes: map[string]string{}}
				return rec
			}, nil)
			data := randomData(200_000, 21)
			loc := createUpload(t, e.url, data, MetaTargetPath, "/docs/f.bin")
			id := loc[strings.LastIndex(loc, "/")+1:]
			if resp := patch(t, loc, 0, data[:100_000]); resp.StatusCode != http.StatusNoContent {
				t.Fatalf("first half: %d", resp.StatusCode)
			}
			if err := tc.spoil(filepath.Join(e.dir, id+".hash")); err != nil {
				t.Fatal(err)
			}
			if resp := patch(t, loc, 100_000, data[100_000:]); resp.StatusCode != http.StatusNoContent {
				t.Fatalf("second half: %d", resp.StatusCode)
			}
			if got, want := rec.get("/docs/f.bin"), sha(data); got != want {
				t.Errorf("committed with hash %q, want %q", got, want)
			}
		})
	}
}

// TestContentHashUsesState proves the finish does not read the file when
// the state covers it: a state made from other bytes than the file's
// gives the state's hash.
func TestContentHashUsesState(t *testing.T) {
	dir := t.TempDir()
	s := &Server{o: Options{Dir: dir, Logger: slog.New(slog.DiscardHandler)}}
	const id = "aaaaaaaaaaaaaaaaaaaaaaaaaa"
	h := sha256.New()
	h.Write([]byte("from the chunks"))
	s.saveHash(id, 15, h)
	data := filepath.Join(dir, id)
	if err := os.WriteFile(data, []byte("different bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := s.contentHash(id, data, 15)
	if err != nil || got != sha([]byte("from the chunks")) {
		t.Errorf("with a state: %q, %v; want the state's hash", got, err)
	}
	// A state for another size is not used: the file is read.
	got, err = s.contentHash(id, data, 16)
	if err != nil || got != sha([]byte("different bytes")) {
		t.Errorf("without a matching state: %q, %v; want the file's hash", got, err)
	}
}
