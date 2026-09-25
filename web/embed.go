// Package web holds the built web interface (S02, ADR-0009), embedded into
// the core binary (ADR-0004). The files come from `pnpm build` in this
// folder. Without a build, build/ holds only its .gitkeep, and the core
// serves a notice instead of the interface (internal/webapp).
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:build
var files embed.FS

// Build returns the built interface: the contents of build/.
func Build() fs.FS {
	sub, err := fs.Sub(files, "build")
	if err != nil {
		panic(err) // build/ is always embedded, so this cannot happen
	}
	return sub
}
