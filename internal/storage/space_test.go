package storage

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

const gib = 1 << 30

func fixedFree(n uint64) FreeFunc {
	return func(string) (uint64, error) { return n, nil }
}

func TestDiskFree(t *testing.T) {
	free, err := DiskFree(t.TempDir())
	if err != nil {
		t.Fatalf("DiskFree: %v", err)
	}
	if free == 0 {
		t.Error("DiskFree reported 0 bytes free on the test machine")
	}
	if _, err := DiskFree(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("DiskFree of a missing directory succeeded")
	}
}

func TestSpaceGuardCheck(t *testing.T) {
	g := NewSpaceGuard("unused", 1*gib, fixedFree(10*gib))
	tests := []struct {
		size int64
		want apperr.Kind // the error kind when ok is false
		ok   bool
	}{
		{0, 0, true},
		{5 * gib, 0, true},
		{9 * gib, 0, true}, // leaves exactly the reserve
		{9*gib + 1, apperr.InsufficientStorage, false},
		{10 * gib, apperr.InsufficientStorage, false},
		{50 * gib, apperr.InsufficientStorage, false}, // more than is free at all
		{-1, apperr.InvalidRequest, false},
	}
	for _, tt := range tests {
		err := g.Check(tt.size)
		if tt.ok {
			if err != nil {
				t.Errorf("Check(%d) = %v, want nil", tt.size, err)
			}
			continue
		}
		if k := apperr.KindOf(err); err == nil || k != tt.want {
			t.Errorf("Check(%d) = %v, want a %s error", tt.size, err, tt.want)
		}
	}
	err := g.Check(9*gib + 1)
	for _, want := range []string{"not enough free space", "9.0 GiB", "10.0 GiB is free", "1.0 GiB must stay free"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("message %q lacks %q", err, want)
		}
	}
	if s := apperr.KindOf(err).Status(); s != http.StatusInsufficientStorage {
		t.Errorf("status = %d, want 507", s)
	}
}

func TestSpaceGuardProviderError(t *testing.T) {
	g := NewSpaceGuard("x", 0, func(string) (uint64, error) { return 0, errors.New("statfs failed") })
	if err := g.Check(1); apperr.KindOf(err) != apperr.Internal || !strings.Contains(err.Error(), "statfs failed") {
		t.Errorf("Check = %v, want an internal error with the cause", err)
	}
}

func TestSpaceGuardDefaults(t *testing.T) {
	dir := t.TempDir()
	g := NewSpaceGuard(dir, -5, nil) // a negative reserve counts as none; nil uses DiskFree
	if g.Reserve() != 0 {
		t.Errorf("Reserve = %d, want 0", g.Reserve())
	}
	free, err := g.Free()
	if err != nil || free == 0 {
		t.Errorf("Free = %d, %v", free, err)
	}
	if err := g.Check(1); err != nil {
		t.Errorf("Check(1) on the real disk = %v", err)
	}
}

func TestFormatBytes(t *testing.T) {
	for n, want := range map[uint64]string{0: "0 B", 1023: "1023 B", 1024: "1.0 KiB", 1536: "1.5 KiB", 5 * gib: "5.0 GiB", 3 << 40: "3.0 TiB"} {
		if got := formatBytes(n); got != want {
			t.Errorf("formatBytes(%d) = %q, want %q", n, got, want)
		}
	}
}

// TestLowSpaceRefusedBeforeWriting is the integration test: a handler that
// stores an upload of a declared size, like the upload handlers of S01.3
// and S01.4 will, answers 507 with a problem body when the injected
// provider reports low space, and no byte reaches the disk.
func TestLowSpaceRefusedBeforeWriting(t *testing.T) {
	r, l := newTestResolver(t)
	guard := NewSpaceGuard(l.Root, 1*gib, fixedFree(1*gib+100))
	logger := slog.New(slog.DiscardHandler)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := guard.Check(req.ContentLength); err != nil {
			apperr.Write(w, req, logger, err)
			return
		}
		root, err := r.OpenRoot(FilesArea, DefaultNamespace)
		if err != nil {
			apperr.Write(w, req, logger, err)
			return
		}
		defer func() { _ = root.Close() }()
		f, err := root.Create("upload.bin")
		if err != nil {
			apperr.Write(w, req, logger, err)
			return
		}
		defer func() { _ = f.Close() }()
		_, _ = io.Copy(f, req.Body)
		w.WriteHeader(http.StatusCreated)
	})
	srv := testutil.NewServer(t, handler)

	for _, tc := range []struct {
		size int
		want int
	}{{100, http.StatusCreated}, {101, http.StatusInsufficientStorage}} {
		_ = os.Remove(filepath.Join(l.Area(FilesArea, DefaultNamespace), "upload.bin"))
		resp, err := http.Post(srv.URL, "application/octet-stream", strings.NewReader(strings.Repeat("x", tc.size)))
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != tc.want {
			t.Errorf("upload of %d bytes: status %d, want %d (%s)", tc.size, resp.StatusCode, tc.want, body)
			continue
		}
		_, statErr := os.Stat(filepath.Join(l.Area(FilesArea, DefaultNamespace), "upload.bin"))
		if tc.want == http.StatusCreated && statErr != nil {
			t.Errorf("the allowed upload was not stored: %v", statErr)
		}
		if tc.want == http.StatusInsufficientStorage {
			if !os.IsNotExist(statErr) {
				t.Error("a refused upload left a file behind")
			}
			var p apperr.Problem
			if err := json.Unmarshal(body, &p); err != nil || p.Code != "insufficient_storage" {
				t.Errorf("body = %s, want an insufficient_storage problem", body)
			}
			if resp.Header.Get("Content-Type") != apperr.ContentType {
				t.Errorf("content type %q", resp.Header.Get("Content-Type"))
			}
		}
	}
}
