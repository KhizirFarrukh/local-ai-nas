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

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/config"
	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
	"github.com/KhizirFarrukh/local-ai-nas/internal/health"
	"github.com/KhizirFarrukh/local-ai-nas/internal/logging"
)

// internalDirName is the internal data directory under the storage root
// (ADR-0003). S01.2 moves the layout into internal/storage.
const internalDirName = ".local-ai-nas"

// Request limits. Routes with large bodies (uploads, S01.3/S01.4) set
// their own limits.
const (
	maxHeaderBytes = 64 << 10
	maxBodyBytes   = 1 << 20
)

// app holds what the serve and migrate commands build from the config.
type app struct {
	cfg    *config.Loaded
	logger *slog.Logger
	db     *db.DB
	close  func()
}

// setup loads the config from the parsed flags, starts logging, and opens
// the database. On error it has already reported to stderr.
func setup(ctx context.Context, fs *flag.FlagSet, stderr io.Writer) (*app, bool) {
	cfg, err := config.Load(config.SourcesFromFlags(fs))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "local-ai-nas: invalid configuration:\n%v\n", err)
		return nil, false
	}
	internal := filepath.Join(cfg.Storage.Root, internalDirName)
	logger, logCloser, err := logging.New(logging.Options{
		Level:        cfg.Log.Level,
		Stderr:       stderr,
		Dir:          filepath.Join(internal, "logs"),
		FileMaxSize:  int64(cfg.Log.FileMaxSize),
		FileMaxFiles: cfg.Log.FileMaxFiles,
	})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "local-ai-nas: %v\n", err)
		return nil, false
	}
	database, err := db.Open(ctx, filepath.Join(internal, "db", db.FileName))
	if err != nil {
		logger.Error("cannot open the database", "error", err.Error())
		_ = logCloser.Close()
		return nil, false
	}
	return &app{cfg: cfg, logger: logger, db: database, close: func() {
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
	srv := &http.Server{
		Handler:           newHandler(log, a.db),
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

// newHandler builds the HTTP handler. From the outside in: request ID,
// body limit, access log, panic recovery, routes. The body limit sits
// outside the access log because http.MaxBytesHandler passes a copy of the
// request on, and the access log must see the request the mux fills in
// (its route pattern).
func newHandler(log *slog.Logger, database *db.DB) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/system/health", health.Handler(version,
		health.Check{Name: "database", Run: database.Read.PingContext},
	))
	var h http.Handler = mux
	h = apperr.Recover(log)(h)
	h = logging.AccessLog(log)(h)
	h = http.MaxBytesHandler(h, maxBodyBytes)
	return logging.RequestID(h)
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
