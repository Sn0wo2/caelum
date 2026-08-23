// Package caelum is a small, slog-based logging library with themed, colorized
// console output, JSON/text formats, file rotation, and a runtime-adjustable
// level.
//
// It is a Go reimagining of the Rust "acta" library, deliberately scoped to
// what fits the slog ecosystem cleanly: it does not attempt tracing-style span
// trees or per-package dynamic filter directives.
package caelum

import (
	"io"
	"log/slog"
)

// Logger wraps an *slog.Logger together with the live controls returned by New.
type Logger struct {
	*slog.Logger
	level   *slog.LevelVar
	closers []io.Closer
}

// New builds a Logger from a Config. The returned Logger embeds an *slog.Logger,
// so it has all the usual Info/Warn/Debug/Error methods, plus SetLevel for
// runtime level changes.
func New(cfg Config) *Logger {
	lv := new(slog.LevelVar)
	lv.Set(cfg.Level)

	targets := cfg.Targets
	if len(targets) == 0 {
		targets = []Target{{}}
	}

	handlers := make([]slog.Handler, 0, len(targets))
	var closers []io.Closer
	for _, t := range targets {
		t = t.resolve()
		handlers = append(handlers, buildHandler(t, lv))
		// A writer that is also an io.Closer owns resources (async workers,
		// rotating files); collect it so Close can flush them uniformly.
		if c, ok := t.Writer.(io.Closer); ok {
			closers = append(closers, c)
		}
	}

	h := handlers[0]
	if len(handlers) > 1 {
		h = &fanoutHandler{handlers: handlers}
	}

	return &Logger{Logger: slog.New(h), level: lv, closers: closers}
}

// Close flushes and releases every target writer that implements io.Closer —
// async queues drain, rotating files sync. Call it before the program exits
// (e.g. defer log.Close()); this is caelum's equivalent of dropping acta's
// TracingGuard. It returns the first error encountered.
func (l *Logger) Close() error {
	var firstErr error
	for _, c := range l.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Init builds a Logger from cfg and installs it as the slog default, so plain
// slog.Info / slog.Error calls anywhere in the program are routed through it.
// It returns the Logger for runtime control.
func Init(cfg Config) *Logger {
	l := New(cfg)
	slog.SetDefault(l.Logger)
	return l
}

// SetLevel changes the minimum severity at runtime. It is safe to call from any
// goroutine and takes effect immediately for every target.
func (l *Logger) SetLevel(level slog.Level) {
	l.level.Set(level)
}

func buildHandler(t Target, level slog.Leveler) slog.Handler {
	opts := &slog.HandlerOptions{Level: level, AddSource: t.ShowSource}
	switch t.Format {
	case JSON:
		return slog.NewJSONHandler(t.Writer, opts)
	case Text:
		return slog.NewTextHandler(t.Writer, opts)
	default:
		return newHandler(t, level)
	}
}
