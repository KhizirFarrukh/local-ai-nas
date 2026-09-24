package uploads

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// sendThenCut sends data[from:to] to the upload and then cuts the
// connection, as a client that loses its network would.
func sendThenCut(t *testing.T, location string, data []byte, from, to int) {
	t.Helper()
	pr, pw := io.Pipe()
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		resp, err := tusRequest(t, ctx, http.MethodPatch, location, map[string]string{
			"Upload-Offset": strconv.Itoa(from), "Content-Type": "application/offset+octet-stream",
		}, pr, int64(len(data)-from))
		if err == nil {
			_ = resp.Body.Close()
		}
		done <- err
	}()
	if _, err := pw.Write(data[from:to]); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	cancel()
	_ = pw.CloseWithError(errors.New("cut off"))
	<-done
}

// headOffset asks the server how many bytes of the upload it has.
func headOffset(t *testing.T, location string) int {
	t.Helper()
	for deadline := time.Now().Add(30 * time.Second); time.Now().Before(deadline); {
		resp, err := tusRequest(t, t.Context(), http.MethodHead, location, nil, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			n, err := strconv.Atoi(resp.Header.Get("Upload-Offset"))
			if err != nil {
				t.Fatal(err)
			}
			return n
		}
		time.Sleep(20 * time.Millisecond) // still locked by the cut-off request
	}
	t.Fatal("the upload stayed locked")
	return 0
}

// TestResumeAtAnyPoint is the S01.4 closure test for criteria 1 and 3: an
// upload cut off at several points (a byte in, mid-chunk, on a boundary,
// a byte before the end) resumes each time from the server's offset, ends
// byte-identical, and is never visible in the files area before it ends.
func TestResumeAtAnyPoint(t *testing.T) {
	e := newTestEnv(t, nil, nil)
	data := make([]byte, 2<<20+777)
	rng := rand.New(rand.NewPCG(8, 9))
	for i := range data {
		data[i] = byte(rng.Uint32())
	}
	location := createUpload(t, e.url, data, MetaTargetPath, "/docs/many.bin")
	notVisible := func(when string) {
		t.Helper()
		entries, err := os.ReadDir(filepath.Join(e.area, "docs"))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.Name() != ".keep" {
				t.Errorf("%s: %s is in the files area", when, entry.Name())
			}
		}
	}

	offset := 0
	for _, cut := range []int{1, 1000, 256<<10 + 3, 1 << 20, len(data) - 1} {
		if cut <= offset {
			continue
		}
		sendThenCut(t, location, data, offset, cut)
		got := headOffset(t, location)
		if got < offset || got > cut {
			t.Fatalf("cut at %d: the server has %d bytes, want %d..%d", cut, got, offset, cut)
		}
		t.Logf("cut at %d: resume from %d", cut, got)
		offset = got
		notVisible("after the cut at " + strconv.Itoa(cut))
	}
	resp, err := tusRequest(t, t.Context(), http.MethodPatch, location, map[string]string{
		"Upload-Offset": strconv.Itoa(offset), "Content-Type": "application/offset+octet-stream",
	}, bytes.NewReader(data[offset:]), int64(len(data)-offset))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent || resp.Header.Get(ItemPathHeader) != "/docs/many.bin" {
		t.Fatalf("the last PATCH: %d %q", resp.StatusCode, resp.Header.Get(ItemPathHeader))
	}
	stored, err := os.ReadFile(filepath.Join(e.area, "docs", "many.bin"))
	if err != nil || sha256.Sum256(stored) != sha256.Sum256(data) {
		t.Fatalf("the finished file differs from the source (%d of %d bytes, %v)", len(stored), len(data), err)
	}
}
