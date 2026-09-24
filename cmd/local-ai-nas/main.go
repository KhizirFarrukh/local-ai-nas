// Command local-ai-nas runs the local-ai-nas server.
//
// Usage:
//
//	local-ai-nas serve   [flags]        run the server
//	local-ai-nas migrate [flags] up     apply pending database migrations
//	local-ai-nas migrate [flags] status list migrations and their state
//	local-ai-nas version                print the version
//
// Flags for serve and migrate: --config and one flag per setting (see
// "local-ai-nas serve -h").
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// version is set at build time with -ldflags "-X main.version=<version>".
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

// Exit codes.
const (
	exitOK    = 0
	exitError = 1 // runtime or configuration error
	exitUsage = 2 // bad command line
)

const usageText = `Usage:
  local-ai-nas serve   [flags]          run the server
  local-ai-nas migrate [flags] up       apply pending database migrations
  local-ai-nas migrate [flags] status   list migrations and their state
  local-ai-nas version                  print the version

Run "local-ai-nas serve -h" for the flags.
`

// run executes the command in args and returns the process exit code.
// Cancelling ctx (Ctrl+C, SIGTERM) stops a running server gracefully.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, usageText)
		return exitUsage
	}
	switch args[0] {
	case "serve":
		return cmdServe(ctx, args[1:], stderr)
	case "migrate":
		return cmdMigrate(ctx, args[1:], stdout, stderr)
	case "version", "--version", "-version":
		_, _ = fmt.Fprintf(stdout, "local-ai-nas %s\n", version)
		return exitOK
	case "help", "-h", "--help", "-help":
		_, _ = io.WriteString(stdout, usageText)
		return exitOK
	}
	_, _ = fmt.Fprintf(stderr, "local-ai-nas: unknown command %q\n\n%s", args[0], usageText)
	return exitUsage
}

// parseFlags parses args with fs and maps the outcome to an exit code:
// -1 to continue, exitOK after -h, or exitUsage.
func parseFlags(fs *flag.FlagSet, args []string) int {
	switch err := fs.Parse(args); {
	case errors.Is(err, flag.ErrHelp):
		return exitOK
	case err != nil:
		return exitUsage
	}
	return -1
}
