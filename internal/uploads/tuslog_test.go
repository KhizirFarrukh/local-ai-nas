package uploads

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	expslog "golang.org/x/exp/slog"
)

// TestTusLogger: tusd's warnings and errors reach the server's log with
// their attributes; its per-request info messages do not.
func TestTusLogger(t *testing.T) {
	var buf bytes.Buffer
	l := tusLogger(slog.New(slog.NewJSONHandler(&buf, nil))).With("method", "PATCH")
	l.Info("RequestIncoming", "path", "/x")
	l.WithGroup("upload").Warn("slow store", "id", "abc", expslog.Group("offset", "from", 1, "to", 2))
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("logged %d lines, want only the warning: %q", len(lines), buf.String())
	}
	var rec map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &rec); err != nil {
		t.Fatal(err)
	}
	tusd, _ := rec["tusd"].(map[string]any)
	upload, _ := tusd["upload"].(map[string]any)
	offset, _ := upload["offset"].(map[string]any)
	if rec["level"] != "WARN" || rec["msg"] != "slow store" || tusd["method"] != "PATCH" || upload["id"] != "abc" || offset["to"] != float64(2) {
		t.Errorf("log record %v", rec)
	}
}
