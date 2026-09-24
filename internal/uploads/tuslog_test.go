package uploads

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"slices"
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

// TestTusLoggerBodyReadError: tusd logs a request body that ends early as
// an error, but that is the client's side (a dropped connection, the case
// tus exists for), so it reaches the log as a warning with its reason.
// tusd's other errors stay errors, and a log at error level leaves the
// warning out.
func TestTusLoggerBodyReadError(t *testing.T) {
	for _, tc := range []struct {
		level slog.Level
		want  []string
	}{
		{slog.LevelInfo, []string{"WARN BodyReadError ERR_UNEXPECTED_EOF", "ERROR InternalServerError "}},
		{slog.LevelError, []string{"ERROR InternalServerError "}},
	} {
		var buf bytes.Buffer
		l := tusLogger(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: tc.level})))
		l.Error("BodyReadError", "error", "ERR_UNEXPECTED_EOF")
		l.Error("InternalServerError", "message", "disk broken")
		var got []string
		for line := range strings.Lines(buf.String()) {
			var rec struct {
				Level, Msg string
				Tusd       struct{ Error string }
			}
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				t.Fatal(err)
			}
			got = append(got, rec.Level+" "+rec.Msg+" "+rec.Tusd.Error)
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("at level %v logged %q, want %q", tc.level, got, tc.want)
		}
	}
}
