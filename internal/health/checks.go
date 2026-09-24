package health

import (
	"context"
	"fmt"

	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// The startup checks of S01.2-T05. Each failing condition is reported
// under its own name.

// Config reports that the configuration loaded and passed validation (the
// server does not start otherwise) and where it came from.
func Config(file string) Check {
	return Check{Name: "config", Run: func(context.Context) (string, error) {
		if file == "" {
			return "valid (defaults, environment, and flags; no config file)", nil
		}
		return "valid (" + file + ")", nil
	}}
}

// StorageWritable checks that every layout directory still exists and
// accepts new files.
func StorageWritable(l storage.Layout) Check {
	return Check{Name: "storage_writable", Run: func(context.Context) (string, error) {
		if err := l.Writable(); err != nil {
			return "", err
		}
		return "the storage root and its areas are writable", nil
	}}
}

// SameFilesystem checks that uploads in progress and the files area are on
// one file system, so finished uploads can be renamed into place. If they
// are not, uploads still work by copying, which is reported as a warning.
// A nil dev uses storage.DeviceID.
func SameFilesystem(l storage.Layout, dev storage.DeviceFunc) Check {
	return Check{Name: "same_filesystem", Run: func(context.Context) (string, error) {
		mode, err := l.FinalizeMode(dev)
		if err != nil {
			return "", err
		}
		if mode == storage.FinalizeCopy {
			return "", Warnf("tmp/uploads and the files area are on different file systems, so finished uploads are copied instead of renamed (upload_finalize_mode=%s)", mode)
		}
		return "upload_finalize_mode=" + mode, nil
	}}
}

// FreeSpace fails when the free space is below the configured reserve.
func FreeSpace(g *storage.SpaceGuard) Check {
	return Check{Name: "free_space", Run: func(context.Context) (string, error) {
		free, err := g.Free()
		if err != nil {
			return "", err
		}
		if free < g.Reserve() {
			return "", fmt.Errorf("%d bytes free, below the reserve of %d bytes; new files are refused", free, g.Reserve())
		}
		return fmt.Sprintf("%d bytes free, reserve %d bytes", free, g.Reserve()), nil
	}}
}

// Database checks that the database answers and every migration is
// applied.
func Database(d *db.DB) Check {
	return Check{Name: "database", Run: func(ctx context.Context) (string, error) {
		if err := d.Read.PingContext(ctx); err != nil {
			return "", err
		}
		status, err := d.MigrationStatus(ctx)
		if err != nil {
			return "", err
		}
		pending := 0
		for _, s := range status {
			if s.State != "applied" {
				pending++
			}
		}
		if pending > 0 {
			return "", fmt.Errorf("%d of %d migrations are not applied", pending, len(status))
		}
		return fmt.Sprintf("open, %d migrations applied", len(status)), nil
	}}
}
