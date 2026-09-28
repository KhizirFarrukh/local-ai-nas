package uploads

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	tus "github.com/tus/tusd/v2/pkg/handler"
)

// TestRepairInfoPaths covers the startup repair after a move of the
// storage root (S01.2-T07): .info files naming another place are pointed
// at the upload directory; current ones, files that are not uploads of
// this server, and unreadable ones are left alone.
func TestRepairInfoPaths(t *testing.T) {
	dir := t.TempDir()
	const (
		moved   = "aaaaaaaaaaaaaaaaaaaaaaaaaa"
		current = "bbbbbbbbbbbbbbbbbbbbbbbbbb"
		noStore = "cccccccccccccccccccccccccc"
		broken  = "dddddddddddddddddddddddddd"
		wrongID = "eeeeeeeeeeeeeeeeeeeeeeeeee"
	)
	write := func(name string, info any) {
		t.Helper()
		b, err := json.Marshal(info)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	old := filepath.Join(t.TempDir(), "old-root", "uploads")
	write(moved+".info", tus.FileInfo{ID: moved, Size: 10, Storage: map[string]string{
		"Type": "filestore", "Path": filepath.Join(old, moved), "InfoPath": filepath.Join(old, moved+".info"),
	}})
	write(current+".info", tus.FileInfo{ID: current, Storage: map[string]string{
		"Type": "filestore", "Path": filepath.Join(dir, current), "InfoPath": filepath.Join(dir, current+".info"),
	}})
	write(noStore+".info", tus.FileInfo{ID: noStore})
	write(wrongID+".info", tus.FileInfo{ID: "someone-else"})
	write("not-an-upload.info", tus.FileInfo{ID: "not-an-upload"})
	if err := os.WriteFile(filepath.Join(dir, broken+".info"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	before := func(name string) []byte {
		b, _ := os.ReadFile(filepath.Join(dir, name))
		return b
	}
	unchanged := map[string][]byte{}
	for _, n := range []string{current + ".info", broken + ".info", wrongID + ".info", "not-an-upload.info"} {
		unchanged[n] = before(n)
	}

	n, err := repairInfoPaths(t.Context(), dir, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("repaired %d files, want 2 (the moved one and the one without paths)", n)
	}
	for _, id := range []string{moved, noStore} {
		var info tus.FileInfo
		if err := json.Unmarshal(before(id+".info"), &info); err != nil {
			t.Fatal(err)
		}
		if info.Storage["Path"] != filepath.Join(dir, id) || info.Storage["InfoPath"] != filepath.Join(dir, id+".info") {
			t.Errorf("%s: storage %v, want paths in %s", id, info.Storage, dir)
		}
	}
	for name, b := range unchanged {
		if got := before(name); string(got) != string(b) {
			t.Errorf("%s changed:\n%s\nwas\n%s", name, got, b)
		}
	}
	if n, err := repairInfoPaths(t.Context(), dir, slog.New(slog.DiscardHandler)); err != nil || n != 0 {
		t.Errorf("a second run changed %d files (%v); want none", n, err)
	}
}
