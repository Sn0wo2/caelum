package caelum

import (
	"context"
	"errors"
	"log/slog"
)

type Logger struct {
	*slog.Logger
	level    *slog.LevelVar
	flushers []interface{ Flush() error }
}

func New(cfg Config) *Logger {
	lv := new(slog.LevelVar)
	lv.Set(cfg.Level)

	targets := cfg.Targets
	if len(targets) == 0 {
		targets = []Target{{}}
	}

	handlers := make([]slog.Handler, 0, len(targets))

	var flushers []interface{ Flush() error }

	for _, t := range targets {
		t = t.resolve()
		handlers = append(handlers, buildHandler(t, lv))

		var output any = t.Writer
		if t.Handler != nil {
			output = t.Handler
		}

		if f, ok := output.(interface{ Flush() error }); ok {
			flushers = append(flushers, f)
		}
	}

	h := handlers[0]
	if len(handlers) > 1 {
		h = slog.NewMultiHandler(handlers...)
	}

	return &Logger{Logger: slog.New(h), level: lv, flushers: flushers}
}

func Init(cfg Config) *Logger {
	l := New(cfg)
	slog.SetDefault(l.Logger)

	return l
}

func (l *Logger) SetLevel(level slog.Level) {
	l.level.Set(level)
}

func (l *Logger) Flush() error {
	var errs []error

	for _, f := range l.flushers {
		if err := f.Flush(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func buildHandler(t Target, level slog.Leveler) slog.Handler {
	opts := t.slogOptions(level)
	if t.Handler != nil {
		return withStacktrace(&levelHandler{Handler: t.Handler, level: opts.Level}, t.Stacktrace)
	}

	switch t.Format {
	case JSON:
		return withStacktrace(slog.NewJSONHandler(t.Writer, &opts), t.Stacktrace)
	case Text:
		return withStacktrace(slog.NewTextHandler(t.Writer, &opts), t.Stacktrace)
	case Compact, Pretty:
		return newHandler(t, level)
	default:
		return newHandler(t, level)
	}
}

type levelHandler struct {
	slog.Handler
	level slog.Leveler
}

func (h *levelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level.Level() && h.Handler.Enabled(ctx, level)
}

func (h *levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelHandler{Handler: h.Handler.WithAttrs(attrs), level: h.level}
}

func (h *levelHandler) WithGroup(name string) slog.Handler {
	return &levelHandler{Handler: h.Handler.WithGroup(name), level: h.level}
}
