package main

import (
	"bytes"
	"testing"
)

func TestRunPrintsVersion(t *testing.T) {
	var out bytes.Buffer
	if code := run(&out); code != 0 {
		t.Fatalf("run() exit code = %d, want 0", code)
	}
	if got, want := out.String(), "local-ai-nas dev\n"; got != want {
		t.Errorf("run() output = %q, want %q", got, want)
	}
}
