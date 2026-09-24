package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api"
	"github.com/KhizirFarrukh/local-ai-nas/internal/config"
	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/health"
	"github.com/KhizirFarrukh/local-ai-nas/internal/logging"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// maxHeaderBytes limits request headers. Body limits are set by the API
// layer (internal/api).
const maxHeaderBytes = 64 << 10

// app holds what the serve and migrate commands build from the config.
type app struct {
	cfg    *config.Loaded
	layout storage.Layout
	logger *slog.Logger
	db     *db.DB
	close  func()
}

// setup loads the config from the parsed flags, prepares the storage
// layout, starts logging, and opens the database. On error it has already
// reported to stderr.
func setup(ctx context.Context, fs *flag.FlagSet, stderr io.Writer) (*app, bool) {
	cfg, err := config.Load(config.SourcesFromFlags(fs))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "local-ai-nas: invalid configuration:\n%v\n", err)
		return nil, false
	}
	layout := storage.NewLayout(cfg.Storage.Root, storage.Options{DBDir: cfg.Storage.DBDir, LogsDir: cfg.Storage.LogsDir})
	unknown, err := layout.Init()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "local-ai-nas: %v\n", err)
		return nil, false
	}
	logger, logCloser, err := logging.New(logging.Options{
		Level:        cfg.Log.Level,
		Stderr:       stderr,
		Dir:          layout.Logs,
		FileMaxSize:  int64(cfg.Log.FileMaxSize),
		FileMaxFiles: cfg.Log.FileMaxFiles,
	})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "local-ai-nas: %v\n", err)
		return nil, false
	}
	if len(unknown) > 0 {
		logger.Warn("the storage root has entries the server does not use; they are left alone",
			"root", layout.Root, "entries", unknown)
	}
	database, err := db.Open(ctx, filepath.Join(layout.DB, db.FileName))
	if err != nil {
		logger.Error("cannot open the database", "error", err.Error())
		_ = logCloser.Close()
		return nil, false
	}
	return &app{cfg: cfg, layout: layout, logger: logger, db: database, close: func() {
		if err := database.Close(); err != nil {
			logger.Error("closing the database", "error", err.Error())
		}
		_ = logCloser.Close()
	}}, true
}

func cmdServe(ctx context.Context, args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	config.RegisterFlags(fs)
	if code := parseFlags(fs, args); code >= 0 {
		return code
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "local-ai-nas serve: unexpected argument %q\n", fs.Arg(0))
		return exitUsage
	}

	a, ok := setup(ctx, fs, stderr)
	if !ok {
		return exitError
	}
	defer a.close()
	log := a.logger
	log.Info("starting", "version", version, "config_file", a.cfg.File,
		"storage_root", a.cfg.Storage.Root, "bind", a.cfg.Server.Bind)

	applied, err := a.db.Migrate(ctx)
	if err != nil {
		log.Error("database migration failed", "error", err.Error())
		return exitError
	}
	for _, m := range applied {
		log.Info("database migration applied", "version", m.Source.Version, "duration_ms", m.Duration.Milliseconds())
	}

	ln, err := net.Listen("tcp", a.cfg.Server.Bind)
	if err != nil {
		log.Error("cannot listen", "bind", a.cfg.Server.Bind, "error", err.Error())
		return exitError
	}
	guard := storage.NewSpaceGuard(a.layout.Root, int64(a.cfg.Storage.FreeSpaceReserve), nil)
	checks := []health.Check{
		health.Config(a.cfg.File),
		health.StorageWritable(a.layout),
		health.SameFilesystem(a.layout, nil),
		health.FreeSpace(guard),
		health.Database(a.db),
	}
	logStartupChecks(ctx, log, checks)

	srv := &http.Server{
		Handler: api.New(api.Options{
			Logger:  log,
			Version: version,
			Checks:  checks,
			Files:   files.NewLocal(storage.NewResolver(a.layout), files.Options{Space: guard}),
			// The simple upload has the same file size limit as tus.
			MaxUploadBytes: int64(a.cfg.Uploads.MaxFileSize),
		}),
		ReadHeaderTimeout: a.cfg.Server.ReadHeaderTimeout.Duration,
		IdleTimeout:       a.cfg.Server.IdleTimeout.Duration,
		MaxHeaderBytes:    maxHeaderBytes,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
	}
	log.Info("listening", "addr", ln.Addr().String())
	if err := serve(ctx, srv, ln, a.cfg.Server.ShutdownTimeout.Duration, log); err != nil {
		log.Error("server stopped with an error", "error", err.Error())
		return exitError
	}
	log.Info("stopped")
	return exitOK
}

// logStartupChecks runs the health checks once at startup and logs the
// outcome. Problems are logged, not fatal: the conditions that make the
// server unusable (an invalid config, an unwritable layout, no database)
// have already stopped setup.
func logStartupChecks(ctx context.Context, log *slog.Logger, checks []health.Check) {
	rep := health.Run(ctx, version, checks)
	for _, c := range rep.Checks {
		switch c.Status {
		case health.StatusFail:
			log.Error("startup check failed", "check", c.Name, "error", c.Error)
		case health.StatusWarn:
			log.Warn("startup check warning", "check", c.Name, "detail", c.Detail)
		default:
			log.Debug("startup check passed", "check", c.Name, "detail", c.Detail)
		}
	}
	log.Info("startup checks", "status", rep.Status)
}

// serve runs srv on ln until ctx is cancelled, then shuts down gracefully:
// it stops accepting connections and waits up to timeout for in-flight
// requests to finish.
func serve(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration, log *slog.Logger) error {
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()

	select {
	case err := <-errc:
		return err // Serve failed on its own; it never returns nil.
	case <-ctx.Done():
	}
	log.Info("shutting down", "timeout", timeout.String())
	sctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()
	if err := srv.Shutdown(sctx); err != nil {
		_ = srv.Close()
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
