package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestRelocatedRoot is the S01.2-T07 acceptance test (NFR-036): a storage
// root copied to another path, with an unfinished resumable upload,
// works there. The upload finishes in the new root, every file reads
// back, and the old copy is left exactly as it was: nothing is written,
// read into the new root, or deleted at the old place.
func TestRelocatedRoot(t *testing.T) {
	first := startLive(t)
	mkdir(t, first.url, "/docs")
	put(t, first.url, "/docs/a.txt", []byte("alpha\n"))
	data := bytes.Repeat([]byte("0123456789abcdef"), 64<<10) // 1 MiB
	half := len(data) / 2

	location := tusCreate(t, first.url, "/docs/big.bin", len(data))
	id := location[strings.LastIndex(location, "/")+1:]
	if got := tusPatch(t, first.url+location, 0, data[:half]); got != http.StatusNoContent {
		t.Fatalf("first half: status %d", got)
	}
	first.stop()

	moved := filepath.Join(t.TempDir(), "moved-root")
	copyTree(t, first.root, moved)
	oldData := filepath.Join(first.root, ".local-ai-nas", "tmp", "uploads", id)
	oldInfo, err := os.ReadFile(oldData + ".info")
	if err != nil {
		t.Fatal(err)
	}

	second := startLiveAt(t, moved)
	if got := tusPatch(t, second.url+location, half, data[half:]); got != http.StatusNoContent {
		t.Fatalf("second half: status %d", got)
	}
	for path, want := range map[string][]byte{"/docs/a.txt": []byte("alpha\n"), "/docs/big.bin": data} {
		if got := get(t, second.url, path); !bytes.Equal(got, want) {
			t.Errorf("%s in the moved root: %d bytes, want %d", path, len(got), len(want))
		}
	}

	// The old place is untouched: the half upload and its info are still
	// there, and its files area did not get the file.
	if st, err := os.Stat(oldData); err != nil || st.Size() != int64(half) {
		t.Errorf("the old upload data changed: %v, %v", st, err)
	}
	if info, err := os.ReadFile(oldData + ".info"); err != nil || !bytes.Equal(info, oldInfo) {
		t.Errorf("the old upload info changed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(first.root, "files", "u0001", "docs", "big.bin")); !os.IsNotExist(err) {
		t.Errorf("the old files area got the upload: %v", err)
	}
}

// copyTree copies the directory from to to, as a user moving a storage
// root would (a copy, so the old place still exists).
func copyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(from, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o750)
		}
		b, err := os.ReadFile(path) // #nosec G304 -- the test's own files
		if err != nil {
			return err
		}
		return os.WriteFile(dst, b, 0o600)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func mkdir(t *testing.T, base, path string) {
	t.Helper()
	resp, err := http.Post(base+"/api/v1/files/folders", "application/json",
		strings.NewReader(`{"path":"`+path+`"}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("creating %s: status %d", path, resp.StatusCode)
	}
}

func put(t *testing.T, base, path string, body []byte) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut,
		base+"/api/v1/files/content?path="+path, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT %s: status %d", path, resp.StatusCode)
	}
}

func get(t *testing.T, base, path string) []byte {
	t.Helper()
	resp, err := http.Get(base + "/api/v1/files/content?path=" + path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// tusCreate starts a resumable upload and returns its path.
func tusCreate(t *testing.T, base, target string, size int) string {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, base+"/api/v1/files/uploads/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", strconv.Itoa(size))
	req.Header.Set("Upload-Metadata", "target_path "+base64.StdEncoding.EncodeToString([]byte(target)))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("creating the upload: status %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if i := strings.Index(loc, "/api/"); i >= 0 {
		loc = loc[i:] // tusd may answer with a full URL
	}
	return loc
}

// tusPatch sends data at offset and returns the status.
func tusPatch(t *testing.T, url string, offset int, data []byte) int {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPatch, url, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", strconv.Itoa(offset))
	req.Header.Set("Content-Type", "application/offset+octet-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	return resp.StatusCode
}
