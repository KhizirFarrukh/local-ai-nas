package uploads

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// createUpload creates an upload of data with the metadata and returns its
// URL.
func createUpload(t *testing.T, url string, data []byte, meta ...string) string {
	t.Helper()
	resp, err := tusRequest(t, t.Context(), http.MethodPost, url+basePath, map[string]string{
		"Upload-Length": strconv.Itoa(len(data)), "Upload-Metadata": metadata(meta...),
	}, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d", resp.StatusCode)
	}
	return resp.Header.Get("Location")
}

// sendAll sends all of data to the upload and returns the answer, with a
// decoded problem when it is an error.
func sendAll(t *testing.T, location string, data []byte) (*http.Response, apperr.Problem) {
	t.Helper()
	resp, err := tusRequest(t, t.Context(), http.MethodPatch, location, map[string]string{
		"Upload-Offset": "0", "Content-Type": "application/offset+octet-stream",
	}, bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var p apperr.Problem
	if resp.StatusCode >= 400 {
		if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
			t.Fatalf("error %d without a problem: %v", resp.StatusCode, err)
		}
	}
	return resp, p
}

// areaFiles returns the files area's content, temporary files included.
func areaFiles(t *testing.T, area string) map[string]string {
	t.Helper()
	got, err := testutil.ReadFiles(area)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// uploadGone checks that nothing of an upload is left in the store and
// the index.
func uploadGone(t *testing.T, dir string, index Index) {
	t.Helper()
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("%d entries left in the upload directory", len(entries))
	}
	var n int
	if err := index.db.Read.QueryRowContext(t.Context(), `SELECT count(*) FROM uploads`).Scan(&n); err != nil || n != 0 {
		t.Errorf("%d session rows left (%v)", n, err)
	}
}

// TestFinalizePolicies: the conflict policy applies when the upload is
// finished, against what is there then.
func TestFinalizePolicies(t *testing.T) {
	url, dir, index, area := testServer(t)
	tests := []struct {
		target, policy, wantPath string
	}{
		{"/taken.txt", "rename", "/taken (1).txt"},
		{"/taken.txt", "overwrite", "/taken.txt"},
		{"/docs/new.bin", "fail", "/docs/new.bin"},
	}
	for _, tt := range tests {
		data := []byte("content for " + tt.target + " with " + tt.policy)
		loc := createUpload(t, url, data, MetaTargetPath, tt.target, MetaOnConflict, tt.policy)
		resp, p := sendAll(t, loc, data)
		if resp.StatusCode != http.StatusNoContent || resp.Header.Get(ItemPathHeader) != tt.wantPath {
			t.Errorf("%s %s: %d %s %+v, want 204 %s", tt.target, tt.policy, resp.StatusCode, resp.Header.Get(ItemPathHeader), p, tt.wantPath)
			continue
		}
		if got := areaFiles(t, area)[tt.wantPath[1:]]; got != string(data) {
			t.Errorf("%s holds %q", tt.wantPath, got)
		}
	}

	// fail: a file that appears after the upload was created wins; the
	// upload is removed.
	data := []byte("late")
	loc := createUpload(t, url, data, MetaTargetPath, "/late.txt")
	if err := os.WriteFile(filepath.Join(area, "late.txt"), []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if resp, p := sendAll(t, loc, data); resp.StatusCode != http.StatusConflict || p.Code != "conflict" {
		t.Errorf("a target that appeared meanwhile: %d %+v, want 409", resp.StatusCode, p)
	}
	if got := areaFiles(t, area)["late.txt"]; got != "first" {
		t.Errorf("late.txt = %q, want the first file", got)
	}
	for name := range areaFiles(t, area) {
		if strings.Contains(name, ".local-ai-nas-tmp-") {
			t.Errorf("a temporary file was left: %s", name)
		}
	}
	uploadGone(t, dir, index)
}

// TestFinalizeChecksum is part of the S01.4-T03 acceptance test: a SHA-256
// mismatch rejects the upload and removes its data.
func TestFinalizeChecksum(t *testing.T) {
	url, dir, index, area := testServer(t)
	data := []byte("the real content")
	sum := sha256.Sum256(data)

	loc := createUpload(t, url, data, MetaTargetPath, "/ok.txt", MetaSHA256, hex.EncodeToString(sum[:]))
	if resp, p := sendAll(t, loc, data); resp.StatusCode != http.StatusNoContent {
		t.Errorf("a matching checksum: %d %+v", resp.StatusCode, p)
	}

	other := sha256.Sum256([]byte("something else"))
	loc = createUpload(t, url, data, MetaTargetPath, "/bad.txt", MetaSHA256, strings.ToUpper(hex.EncodeToString(other[:])))
	resp, p := sendAll(t, loc, data)
	if resp.StatusCode != http.StatusBadRequest || p.Code != "invalid_request" || !strings.Contains(p.Detail, "SHA-256") {
		t.Errorf("a wrong checksum: %d %+v, want 400 about the SHA-256", resp.StatusCode, p)
	}
	got := areaFiles(t, area)
	if _, ok := got["bad.txt"]; ok || got["ok.txt"] != string(data) || len(got) != len(fixture)+1 {
		t.Errorf("files area after the checksums: %v", got)
	}
	uploadGone(t, dir, index)
}

// TestFinalizeInOneRequest: an upload whose data comes with its creation
// (creation-with-upload), and an empty upload, finish at once.
func TestFinalizeInOneRequest(t *testing.T) {
	url, dir, index, area := testServer(t)
	data := []byte("all in the POST")
	resp, err := tusRequest(t, t.Context(), http.MethodPost, url+basePath, map[string]string{
		"Upload-Length": strconv.Itoa(len(data)), "Upload-Metadata": metadata(MetaTargetPath, "/one.txt"),
		"Content-Type": "application/offset+octet-stream",
	}, bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated || resp.Header.Get(ItemPathHeader) != "/one.txt" {
		t.Errorf("creation with upload: %d %q", resp.StatusCode, resp.Header.Get(ItemPathHeader))
	}
	resp, err = tusRequest(t, t.Context(), http.MethodPost, url+basePath, map[string]string{
		"Upload-Length": "0", "Upload-Metadata": metadata(MetaTargetPath, "/empty.txt"),
	}, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated || resp.Header.Get(ItemPathHeader) != "/empty.txt" {
		t.Errorf("empty upload: %d %q", resp.StatusCode, resp.Header.Get(ItemPathHeader))
	}
	got := areaFiles(t, area)
	if got["one.txt"] != string(data) || got["empty.txt"] != "" {
		t.Errorf("files area: %v", got)
	}
	if _, ok := got["empty.txt"]; !ok {
		t.Error("the empty upload is missing")
	}
	uploadGone(t, dir, index)
}

// TestFinalizeFaults is the S01.4-T03 fault-injection test: a crash before
// the commit leaves nothing in the files area; a crash after it leaves
// the complete file.
func TestFinalizeFaults(t *testing.T) {
	for _, tt := range []struct {
		stage    string
		wantFile bool
	}{
		{"verified", false},
		{"committed", true},
	} {
		t.Run(tt.stage, func(t *testing.T) {
			orig := finalizeFault
			finalizeFault = func(stage string) error {
				if stage == tt.stage {
					return errors.New("simulated crash at " + stage)
				}
				return nil
			}
			t.Cleanup(func() { finalizeFault = orig })

			url, _, _, area := testServer(t)
			data := bytes.Repeat([]byte("x"), 1<<20)
			loc := createUpload(t, url, data, MetaTargetPath, "/crash.bin")
			if resp, _ := sendAll(t, loc, data); resp.StatusCode != http.StatusInternalServerError {
				t.Errorf("the crashed finalize answered %d, want 500", resp.StatusCode)
			}
			got := areaFiles(t, area)
			_, there := got["crash.bin"]
			if there != tt.wantFile || there && got["crash.bin"] != string(data) {
				t.Errorf("crash.bin present: %v (want %v), %d bytes", there, tt.wantFile, len(got["crash.bin"]))
			}
			for name := range got {
				if strings.Contains(name, ".local-ai-nas-tmp-") {
					t.Errorf("a temporary file was left: %s", name)
				}
			}
		})
	}
}

// notSameDevice pretends that the upload directory is on another file
// system than the area.
type notSameDevice struct {
	Target
	tries int
}

func (n *notSameDevice) CommitUpload(context.Context, string, string, string, files.UploadOptions) (files.Item, bool, error) {
	n.tries++
	return files.Item{}, false, files.ErrNotSameDevice
}

// TestFinalizeCopyFallback: when the upload cannot be renamed into the
// area, it is copied (copy, fsync, commit) instead.
func TestFinalizeCopyFallback(t *testing.T) {
	var fake *notSameDevice
	url, dir, index, area := testServerWrap(t, func(t Target) Target {
		fake = &notSameDevice{Target: t}
		return fake
	})
	data := bytes.Repeat([]byte("copy me "), 100000)
	loc := createUpload(t, url, data, MetaTargetPath, "/copied.bin")
	if resp, p := sendAll(t, loc, data); resp.StatusCode != http.StatusNoContent || resp.Header.Get(ItemPathHeader) != "/copied.bin" {
		t.Fatalf("finish: %d %+v", resp.StatusCode, p)
	}
	if fake.tries != 1 || areaFiles(t, area)["copied.bin"] != string(data) {
		t.Errorf("rename tries %d; the copied file is complete: %v", fake.tries, areaFiles(t, area)["copied.bin"] == string(data))
	}
	uploadGone(t, dir, index)
}
