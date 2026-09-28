package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestContentHashes is the S01.3-T10 acceptance test (FR-211): the
// program stores the hash of every upload and copy, the hash follows
// renames and moves (files and folders), an overwrite gets the new hash,
// a file changed outside the NAS has none, and a deleted file's hash is
// gone. It runs the real program with its real database.
func TestContentHashes(t *testing.T) {
	s := startLive(t)
	sum := func(b []byte) string {
		d := sha256.Sum256(b)
		return "sha256:" + hex.EncodeToString(d[:])
	}
	alpha, beta := []byte("alpha\n"), bytes.Repeat([]byte("beta"), 300_000)

	mkdir(t, s.url, "/docs")
	put(t, s.url, "/docs/a.txt", alpha)
	put(t, s.url, "/docs/b.bin", beta)
	check := func(path, want string) {
		t.Helper()
		if got := contentHash(t, s.url, path); got != want {
			t.Errorf("%s: content_hash %q, want %q", path, got, want)
		}
	}
	check("/docs/a.txt", sum(alpha))
	check("/docs/b.bin", sum(beta))

	// The upload's own answer carries it too.
	if got := putHash(t, s.url, "/c.txt", []byte("c")); got != sum([]byte("c")) {
		t.Errorf("upload answer: content_hash %q", got)
	}

	// Renames and moves keep it; the old path has none.
	post(t, s.url, "/api/v1/files/operations/rename", `{"path":"/docs/a.txt","new_name":"a2.txt"}`)
	check("/docs/a2.txt", sum(alpha))
	mkdir(t, s.url, "/archive")
	post(t, s.url, "/api/v1/files/operations/move", `{"from":"/docs","to":"/archive/docs"}`)
	check("/archive/docs/a2.txt", sum(alpha))
	check("/archive/docs/b.bin", sum(beta))

	// Copies get their own row, for a file and for every file of a folder.
	post(t, s.url, "/api/v1/files/operations/copy", `{"from":"/archive/docs/b.bin","to":"/b-copy.bin"}`)
	check("/b-copy.bin", sum(beta))
	post(t, s.url, "/api/v1/files/operations/copy", `{"from":"/archive/docs","to":"/docs-copy"}`)
	check("/docs-copy/a2.txt", sum(alpha))
	check("/docs-copy/b.bin", sum(beta))

	// An overwrite gets the new content's hash.
	putOverwrite(t, s.url, "/b-copy.bin", alpha)
	check("/b-copy.bin", sum(alpha))

	// A file changed outside the NAS has no hash until it is hashed again.
	outside := filepath.Join(s.root, "files", "u0001", "docs-copy", "a2.txt")
	if err := os.WriteFile(outside, []byte("changed outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	later := time.Now().Add(2 * time.Second) // a new version even on coarse clocks
	if err := os.Chtimes(outside, later, later); err != nil {
		t.Fatal(err)
	}
	check("/docs-copy/a2.txt", "")

	// Deleting forgets: the same name written outside afterwards has none.
	del(t, s.url, "/archive/docs")
	if err := os.MkdirAll(filepath.Join(s.root, "files", "u0001", "archive", "docs"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.root, "files", "u0001", "archive", "docs", "a2.txt"), alpha, 0o600); err != nil {
		t.Fatal(err)
	}
	check("/archive/docs/a2.txt", "")
}

// contentHash returns the content_hash of the item at path ("" if none).
func contentHash(t *testing.T, base, path string) string {
	t.Helper()
	resp, err := http.Get(base + "/api/v1/files/items?path=" + url.QueryEscape(path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("details of %s: status %d", path, resp.StatusCode)
	}
	var body struct {
		Item struct {
			ContentHash string `json:"content_hash"`
		} `json:"item"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body.Item.ContentHash
}

func post(t *testing.T, base, path, body string) {
	t.Helper()
	resp, err := http.Post(base+path, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode >= 300 {
		t.Fatalf("POST %s %s: status %d", path, body, resp.StatusCode)
	}
}

func del(t *testing.T, base, path string) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodDelete,
		base+"/api/v1/files/items?recursive=true&path="+url.QueryEscape(path), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE %s: status %d", path, resp.StatusCode)
	}
}

// putHash uploads body to path and returns the content_hash of the answer.
func putHash(t *testing.T, base, path string, body []byte) string {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut,
		base+"/api/v1/files/content?path="+url.QueryEscape(path), bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var item struct {
		ContentHash string `json:"content_hash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		t.Fatal(err)
	}
	return item.ContentHash
}

func putOverwrite(t *testing.T, base, path string, body []byte) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut,
		base+"/api/v1/files/content?on_conflict=overwrite&path="+url.QueryEscape(path), bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("overwrite %s: status %d", path, resp.StatusCode)
	}
}
