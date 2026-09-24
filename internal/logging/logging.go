// Package logging sets up structured JSON logging with log/slog, to stderr
// and to a size-rotated log file, plus the request-ID and access-log
// middleware (S01.1-T08, NFR-016). Logs never leave the machine.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Options configures New.
type Options struct {
	// Level is debug, info, warn, or error.
	Level string
	// Stderr receives every log line (os.Stderr in production).
	Stderr io.Writer
	// Dir is the logs directory, such as <root>/.local-ai-nas/logs. It is
	// created if missing. An empty Dir disables the log file.
	Dir string
	// FileMaxSize is the size in bytes at which the log file is rotated.
	FileMaxSize int64
	// FileMaxFiles is the number of rotated files kept.
	FileMaxFiles int
}

// ParseLevel parses debug, info, warn, or error (any case).
func ParseLevel(s string) (slog.Level, error) {
	var l slog.Level
	switch strings.ToLower(s) {
	case "debug", "info", "warn", "error":
		return l, l.UnmarshalText([]byte(s))
	}
	return l, fmt.Errorf("logging: unknown level %q (want debug, info, warn, or error)", s)
}

// New returns a logger that writes JSON lines to opts.Stderr and to the
// rotated file in opts.Dir. The returned Closer closes the file; call it
// after the last log line.
func New(opts Options) (*slog.Logger, io.Closer, error) {
	level, err := ParseLevel(opts.Level)
	if err != nil {
		return nil, nil, err
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	out := stderr
	var closer io.Closer = nopCloser{}
	if opts.Dir != "" {
		if err := os.MkdirAll(opts.Dir, 0o750); err != nil {
			return nil, nil, fmt.Errorf("logging: create logs directory: %w", err)
		}
		f, err := OpenRotatingFile(DefaultPath(opts.Dir), opts.FileMaxSize, opts.FileMaxFiles)
		if err != nil {
			return nil, nil, err
		}
		out = &teeWriter{primary: f, secondary: stderr}
		closer = f
	}
	h := slog.NewJSONHandler(out, &slog.HandlerOptions{Level: level})
	return slog.New(h), closer, nil
}

// teeWriter writes to both writers. A failing stderr never stops the file
// from receiving the line, and the reverse.
type teeWriter struct {
	primary, secondary io.Writer
}

func (t *teeWriter) Write(p []byte) (int, error) {
	_, err1 := t.primary.Write(p)
	_, err2 := t.secondary.Write(p)
	if err1 != nil {
		return 0, err1
	}
	if err2 != nil {
		return 0, err2
	}
	return len(p), nil
}

type nopCloser struct{}

func (nopCloser) Close() error { return nil }
