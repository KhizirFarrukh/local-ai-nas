// Command local-ai-nas runs the local-ai-nas server.
//
// This is the repository skeleton (S01.1-T01). The subcommands serve,
// migrate, and version are added in S01.1-T11.
package main

import (
	"fmt"
	"io"
	"os"
)

// version is set at build time with -ldflags "-X main.version=<version>".
var version = "dev"

func main() {
	os.Exit(run(os.Stdout))
}

// run executes the command and returns the process exit code.
func run(stdout io.Writer) int {
	if _, err := fmt.Fprintf(stdout, "local-ai-nas %s\n", version); err != nil {
		return 1
	}
	return 0
}
