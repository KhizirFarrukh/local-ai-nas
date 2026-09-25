package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/config"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
	"github.com/KhizirFarrukh/local-ai-nas/internal/uploads"
)

// heapBound is the most the heap may grow during one transfer (S01.4-T04).
const heapBound = 256 << 20

// memTestSize returns the transfer size from LOCALAINAS_MEMTEST_SIZE, or
// skips the test: it writes that many bytes several times.
func memTestSize(t *testing.T) int64 {
	t.Helper()
	v := os.Getenv("LOCALAINAS_MEMTEST_SIZE")
	if v == "" {
		t.Skip("set LOCALAINAS_MEMTEST_SIZE (such as 10GiB) to run the memory test")
	}
	var size config.ByteSize
	if err := size.UnmarshalText([]byte(v)); err != nil {
		t.Fatalf("LOCALAINAS_MEMTEST_SIZE: %v", err)
	}
	return int64(size)
}

// heapSampler records the largest heap size while a transfer runs.
type heapSampler struct {
	stop, done chan struct{}
	base, max  uint64
}

func sampleHeap() *heapSampler {
	runtime.GC()
	h := &heapSampler{stop: make(chan struct{}), done: make(chan struct{}), base: heapObjects()}
	h.max = h.base
	go func() {
		defer close(h.done)
		tick := time.NewTicker(20 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-h.stop:
				return
			case <-tick.C:
				h.max = max(h.max, heapObjects())
			}
		}
	}()
	return h
}

// growth stops the sampler and returns the largest growth over the base.
func (h *heapSampler) growth() uint64 {
	close(h.stop)
	<-h.done
	return h.max - h.base
}

// zeros is an endless source of zero bytes.
type zeros struct{}

func (zeros) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}

// TestMemoryBound is the S01.4-T04 acceptance test: a transfer of
// LOCALAINAS_MEMTEST_SIZE bytes (10 GiB in the Linux CI job, 1 GiB on
// Windows) never grows the heap by more than 256 MiB, for the simple
// upload, the tus upload, a copy, and a download.
func TestMemoryBound(t *testing.T) {
	size := memTestSize(t)
	svc, area := testFilesDir(t)
	srv := testutil.NewServer(t, New(Options{Files: svc, Uploads: testUploads(t, svc), MaxUploadBytes: size}))
	c := client{t, srv.URL}
	if err := os.MkdirAll(filepath.Join(area, "mem"), 0o750); err != nil {
		t.Fatal(err)
	}
	check := func(what string, h *heapSampler, path string) {
		t.Helper()
		g := h.growth()
		t.Logf("%s of %d bytes: the heap grew by at most %.1f MiB", what, size, float64(g)/(1<<20))
		if g > heapBound {
			t.Errorf("%s: the heap grew by %d bytes, more than %d", what, g, heapBound)
		}
		if path != "" {
			st, _, b := c.do("GET", "/api/v1/files/items?path="+q(path), "", nil, nil)
			var r gen.ItemsResponse
			if err := json.Unmarshal(b, &r); st != http.StatusOK || err != nil || r.Item.Size != size {
				t.Errorf("%s: %s is %d bytes (%d, %v), want %d", what, path, r.Item.Size, st, err, size)
			}
			if st, _, b := c.do("DELETE", "/api/v1/files/items?path="+q(path), "", nil, nil); st != http.StatusNoContent {
				t.Errorf("%s: delete %s: %d %s", what, path, st, b)
			}
		}
	}

	// The simple upload.
	h := sampleHeap()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, srv.URL+"/api/v1/files/content?path="+q("/mem/simple.bin"), io.LimitReader(zeros{}, size))
	if err != nil {
		t.Fatal(err)
	}
	req.ContentLength = size
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("simple upload: %d", resp.StatusCode)
	}
	check("simple upload", h, "/mem/simple.bin")

	// The tus upload, in chunks of the default chunk limit.
	h = sampleHeap()
	st, hdr, b := c.do("POST", UploadsPath, "", nil, map[string]string{
		"Tus-Resumable": "1.0.0", "Upload-Length": strconv.FormatInt(size, 10),
		"Upload-Metadata": "target_path " + b64("/mem/tus.bin"),
	})
	if st != http.StatusCreated {
		t.Fatalf("tus create: %d %s", st, b)
	}
	location := hdr.Get("Location")
	for offset := int64(0); offset < size; {
		n := min(int64(DefaultMaxChunkBytes), size-offset)
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPatch, location, io.LimitReader(zeros{}, n))
		if err != nil {
			t.Fatal(err)
		}
		req.ContentLength = n
		req.Header.Set("Tus-Resumable", "1.0.0")
		req.Header.Set("Upload-Offset", strconv.FormatInt(offset, 10))
		req.Header.Set("Content-Type", "application/offset+octet-stream")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("tus PATCH at %d: %d", offset, resp.StatusCode)
		}
		offset += n
		if offset == size && resp.Header.Get(uploads.ItemPathHeader) != "/mem/tus.bin" {
			t.Errorf("the last PATCH answered %s %q", uploads.ItemPathHeader, resp.Header.Get(uploads.ItemPathHeader))
		}
	}
	check("tus upload", h, "/mem/tus.bin")

	// A sparse file (no disk used) to copy and to download.
	sparse := filepath.Join(area, "mem", "sparse.bin")
	f, err := os.Create(sparse)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(size); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	h = sampleHeap()
	if st, _, b := c.do("POST", "/api/v1/files/operations/copy", "application/json", []byte(`{"from":"/mem/sparse.bin","to":"/mem/copy.bin"}`), nil); st != http.StatusCreated {
		t.Fatalf("copy: %d %s", st, b)
	}
	check("copy", h, "/mem/copy.bin")

	h = sampleHeap()
	resp, err = http.Get(srv.URL + "/api/v1/files/content?path=" + q("/mem/sparse.bin"))
	if err != nil {
		t.Fatal(err)
	}
	n, err := io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if err != nil || n != size || resp.StatusCode != http.StatusOK {
		t.Errorf("download: %d bytes, %d, %v", n, resp.StatusCode, err)
	}
	check("download", h, "")
}

// b64 encodes a tus metadata value.
func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
