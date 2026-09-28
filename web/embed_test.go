package web

import (
	"io/fs"
	"testing"
)

// TestBuildKeepsItsPlaceholder is the regression test of bug S02-B01: `vite
// build` empties web/build, which once deleted the committed .gitkeep, and a
// fresh clone without web/build does not compile (//go:embed all:build). The
// build script writes the file again; this test fails when it is missing,
// with or without a built interface.
func TestBuildKeepsItsPlaceholder(t *testing.T) {
	if _, err := fs.Stat(Build(), ".gitkeep"); err != nil {
		t.Fatalf("web/build/.gitkeep is missing: %v", err)
	}
}
