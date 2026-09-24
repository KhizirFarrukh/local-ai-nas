package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/KhizirFarrukh/local-ai-nas/internal/config"
)

func cmdMigrate(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	config.RegisterFlags(fs)
	if code := parseFlags(fs, args); code >= 0 {
		return code
	}
	if fs.NArg() != 1 || (fs.Arg(0) != "up" && fs.Arg(0) != "status") {
		_, _ = fmt.Fprintln(stderr, `local-ai-nas migrate: want exactly one action, "up" or "status"`)
		return exitUsage
	}

	a, ok := setup(ctx, fs, stderr)
	if !ok {
		return exitError
	}
	defer a.close()

	if fs.Arg(0) == "up" {
		applied, err := a.db.Migrate(ctx)
		if err != nil {
			a.logger.Error("database migration failed", "error", err.Error())
			return exitError
		}
		for _, m := range applied {
			_, _ = fmt.Fprintf(stdout, "applied %05d %s\n", m.Source.Version, m.Duration)
		}
		if len(applied) == 0 {
			_, _ = fmt.Fprintln(stdout, "the database is up to date")
		}
		return exitOK
	}

	status, err := a.db.MigrationStatus(ctx)
	if err != nil {
		a.logger.Error("reading the migration status failed", "error", err.Error())
		return exitError
	}
	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "VERSION\tSTATE\tAPPLIED AT")
	for _, s := range status {
		at := "-"
		if !s.AppliedAt.IsZero() {
			at = s.AppliedAt.UTC().Format("2006-01-02 15:04:05Z")
		}
		_, _ = fmt.Fprintf(tw, "%05d\t%s\t%s\n", s.Source.Version, s.State, at)
	}
	_ = tw.Flush()
	return exitOK
}
