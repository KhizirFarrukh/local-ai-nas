package uploads

import (
	"context"
	"log/slog"

	expslog "golang.org/x/exp/slog"
)

// tusLogger adapts the server's logger to tusd, which logs through
// golang.org/x/exp/slog, a separate type from log/slog. Only warnings and
// errors pass: tusd also logs every request and response at info level,
// which the access log already records.
func tusLogger(l *slog.Logger) *expslog.Logger {
	return expslog.New(tusHandler{h: l.Handler().WithGroup("tusd")})
}

// tusHandler forwards x/exp/slog records to a log/slog handler.
type tusHandler struct{ h slog.Handler }

func (t tusHandler) Enabled(ctx context.Context, level expslog.Level) bool {
	return level >= expslog.LevelWarn && t.h.Enabled(ctx, slog.Level(level))
}

func (t tusHandler) Handle(ctx context.Context, r expslog.Record) error {
	level := slog.Level(r.Level)
	// tusd logs a request body that ends early as an error. That happens on
	// the client's side (a dropped connection, a timeout, a body over the
	// limit), and resuming after it is what tus is for. So it is a warning
	// here, and errors stay for faults of the server.
	if r.Message == "BodyReadError" && level > slog.LevelWarn {
		level = slog.LevelWarn
		if !t.h.Enabled(ctx, level) {
			return nil
		}
	}
	rec := slog.NewRecord(r.Time, level, r.Message, r.PC)
	r.Attrs(func(a expslog.Attr) bool {
		rec.AddAttrs(convertAttr(a))
		return true
	})
	return t.h.Handle(ctx, rec)
}

func (t tusHandler) WithAttrs(attrs []expslog.Attr) expslog.Handler {
	out := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		out[i] = convertAttr(a)
	}
	return tusHandler{h: t.h.WithAttrs(out)}
}

func (t tusHandler) WithGroup(name string) expslog.Handler {
	return tusHandler{h: t.h.WithGroup(name)}
}

// convertAttr converts one attribute, including groups.
func convertAttr(a expslog.Attr) slog.Attr {
	v := a.Value.Resolve()
	if v.Kind() == expslog.KindGroup {
		group := v.Group()
		attrs := make([]any, len(group))
		for i, g := range group {
			attrs[i] = convertAttr(g)
		}
		return slog.Group(a.Key, attrs...)
	}
	return slog.Any(a.Key, v.Any())
}
