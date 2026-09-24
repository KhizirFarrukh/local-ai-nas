package api

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/config"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// perfSize returns the transfer size of the performance baseline from
// LOCALAINAS_PERF_SIZE, or skips the test.
func perfSize(t *testing.T) int64 {
	t.Helper()
	v := os.Getenv("LOCALAINAS_PERF_SIZE")
	if v == "" {
		t.Skip("set LOCALAINAS_PERF_SIZE (such as 1GiB) to run the performance baseline")
	}
	var size config.ByteSize
	if err := size.UnmarshalText([]byte(v)); err != nil {
		t.Fatalf("LOCALAINAS_PERF_SIZE: %v", err)
	}
	return int64(size)
}

// p95 returns the 95th percentile of the durations.
func p95(d []time.Duration) time.Duration {
	s := slices.Clone(d)
	slices.Sort(s)
	return s[(len(s)*95+99)/100-1]
}

// mbps returns megabytes (10^6) per second.
func mbps(n int64, d time.Duration) float64 { return float64(n) / 1e6 / d.Seconds() }

// rawWrite writes n bytes to path with 1 MiB writes and an fsync, as dd
// with conv=fsync would, and returns the time.
func rawWrite(t *testing.T, path string, n int64) time.Duration {
	t.Helper()
	buf := make([]byte, 1<<20)
	start := time.Now()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	for left := n; left > 0; left -= int64(len(buf)) {
		if _, err := f.Write(buf[:min(left, int64(len(buf)))]); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return time.Since(start)
}

// rawRead reads path with 1 MiB reads and returns the time.
func rawRead(t *testing.T, path string) time.Duration {
	t.Helper()
	start := time.Now()
	f, err := os.Open(path) // #nosec G304 -- a file this test wrote
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.CopyBuffer(io.Discard, onlyReader{f}, make([]byte, 1<<20)); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return time.Since(start)
}

// rawLoopback sends n bytes over a loopback TCP connection with 1 MiB
// writes and reads, and returns the time: the network baseline.
func rawLoopback(t *testing.T, n int64) time.Duration {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	done := make(chan error, 1)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			done <- err
			return
		}
		_, err = io.CopyBuffer(struct{ io.Writer }{io.Discard}, onlyReader{conn}, make([]byte, 1<<20))
		_ = conn.Close()
		done <- err
	}()
	start := time.Now()
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.CopyBuffer(struct{ io.Writer }{conn}, onlyReader{io.LimitReader(zeros{}, n)}, make([]byte, 1<<20)); err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	return time.Since(start)
}

// onlyReader hides WriterTo, so the copy uses the given buffer.
type onlyReader struct{ r io.Reader }

func (o onlyReader) Read(p []byte) (int, error) { return o.r.Read(p) }

// TestPerfBaseline is the S01.7-T04 measurement against NFR-003: listing
// a 10,000-entry folder (p95 ≤ 500 ms) and upload and download throughput
// (≥ 80% of the raw disk), over real HTTP on loopback. It writes the
// results to LOCALAINAS_PERF_OUT when set.
func TestPerfBaseline(t *testing.T) {
	size := perfSize(t)
	svc, area := testFilesDir(t)
	srv := testutil.NewServer(t, New(Options{Files: svc, Uploads: testUploads(t, svc), MaxUploadBytes: size}))
	c := client{t, srv.URL}
	// A client with large buffers, so the client is not what limits the
	// transfers (Go's default transport buffers are 4 KiB).
	fast := &http.Client{Transport: &http.Transport{WriteBufferSize: 1 << 20, ReadBufferSize: 1 << 20}}
	var report strings.Builder
	line := func(format string, args ...any) {
		fmt.Fprintf(&report, format+"\n", args...)
		t.Logf(format, args...)
	}
	line("| Measurement | Result | Target |")
	line("|---|---|---|")

	// Listing: 10,000 files, the first page (100) and a whole-folder page
	// (1000 × 10), each sorted by name and by size.
	files := make(map[string]string, 10000)
	for i := range 10000 {
		files[fmt.Sprintf("big/file-%05d.txt", i)] = strings.Repeat("x", i%100)
	}
	if err := testutil.WriteFiles(area, files); err != nil {
		t.Fatal(err)
	}
	for _, sort := range []string{"name", "size", "mod_time"} {
		var times []time.Duration
		for range 40 {
			start := time.Now()
			st, _, _ := c.do("GET", "/api/v1/files/items?path=/big&limit=100&sort="+sort, "", nil, nil)
			times = append(times, time.Since(start))
			if st != http.StatusOK {
				t.Fatalf("list: %d", st)
			}
		}
		p := p95(times)
		line("| List 10,000 entries, first page of 100, sort=%s: p95 | %d ms | ≤ 500 ms (%s) |", sort, p.Milliseconds(), verdict(p <= 500*time.Millisecond))
	}
	start := time.Now()
	if names := listAll(t, c, "/big", 1000); len(names) != 10000 {
		t.Fatalf("listed %d", len(names))
	}
	line("| List all 10,000 entries in pages of 1,000 (10 requests) | %d ms | (information) |", time.Since(start).Milliseconds())

	// Transfers. Every figure is the median of three runs, because the
	// disk's caches make single runs vary a lot. NFR-003 asks for 80% of
	// the raw disk or network, whichever is slower: here the network is
	// loopback TCP, measured raw as well.
	med := func(run func() time.Duration) time.Duration {
		d := []time.Duration{run(), run(), run()}
		slices.Sort(d)
		return d[1]
	}
	raw := filepath.Join(filepath.Dir(area), "raw.bin")
	rw := med(func() time.Duration { return rawWrite(t, raw, size) })
	rr := med(func() time.Duration { return rawRead(t, raw) })
	_ = os.Remove(raw)
	rn := med(func() time.Duration { return rawLoopback(t, size) })
	line("| Raw disk write, 1 MiB writes + fsync (%d MiB) | %.0f MB/s | baseline |", size>>20, mbps(size, rw))
	line("| Raw disk read, 1 MiB reads (%d MiB, from the cache) | %.0f MB/s | baseline |", size>>20, mbps(size, rr))
	line("| Raw loopback TCP, 1 MiB writes (%d MiB) | %.0f MB/s | baseline |", size>>20, mbps(size, rn))
	against := func(name string, d, disk time.Duration) {
		base, of := disk, "raw disk"
		if rn > base {
			base, of = rn, "raw loopback TCP"
		}
		ratio := base.Seconds() / d.Seconds()
		line("| %s (%d MiB) | %.0f MB/s = %.0f%% of the %s | ≥ 80%% (%s) |", name, size>>20, mbps(size, d), 100*ratio, of, verdict(ratio >= 0.8))
	}

	upload := func() time.Duration {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, srv.URL+"/api/v1/files/content?path=/up.bin&on_conflict=overwrite", io.LimitReader(zeros{}, size))
		if err != nil {
			t.Fatal(err)
		}
		req.ContentLength = size
		start := time.Now()
		resp, err := fast.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			t.Fatalf("upload: %d", resp.StatusCode)
		}
		return time.Since(start)
	}
	against("Simple upload over HTTP", med(upload), rw)

	download := func() time.Duration {
		start := time.Now()
		resp, err := fast.Get(srv.URL + "/api/v1/files/content?path=/up.bin")
		if err != nil {
			t.Fatal(err)
		}
		n, err := io.CopyBuffer(struct{ io.Writer }{io.Discard}, onlyReader{resp.Body}, make([]byte, 1<<20))
		_ = resp.Body.Close()
		if err != nil || n != size {
			t.Fatalf("download: %d bytes, %v", n, err)
		}
		return time.Since(start)
	}
	against("Download over HTTP (from the cache)", med(download), rr)

	tus := func() time.Duration {
		start := time.Now()
		st, hdr, b := c.do("POST", UploadsPath, "", nil, map[string]string{
			"Tus-Resumable": "1.0.0", "Upload-Length": strconv.FormatInt(size, 10),
			"Upload-Metadata": "target_path " + b64("/tus.bin") + ",on_conflict " + b64("overwrite"),
		})
		if st != http.StatusCreated {
			t.Fatalf("tus create: %d %s", st, b)
		}
		loc := hdr.Get("Location")
		for off := int64(0); off < size; {
			chunk := min(int64(DefaultMaxChunkBytes), size-off)
			req, err := http.NewRequestWithContext(t.Context(), http.MethodPatch, loc, io.LimitReader(zeros{}, chunk))
			if err != nil {
				t.Fatal(err)
			}
			req.ContentLength = chunk
			req.Header.Set("Tus-Resumable", "1.0.0")
			req.Header.Set("Upload-Offset", strconv.FormatInt(off, 10))
			req.Header.Set("Content-Type", "application/offset+octet-stream")
			resp, err := fast.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusNoContent {
				t.Fatalf("tus PATCH: %d", resp.StatusCode)
			}
			off += chunk
		}
		return time.Since(start)
	}
	against("tus upload over HTTP, 64 MiB chunks", med(tus), rw)

	line("")
	line("Go %s, %s/%s, %d CPUs, transfer size %d MiB.", runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), size>>20)
	if out := os.Getenv("LOCALAINAS_PERF_OUT"); out != "" {
		if err := os.WriteFile(out, []byte(report.String()), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func verdict(ok bool) string {
	if ok {
		return "met"
	}
	return "**not met**"
}
