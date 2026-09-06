package caelum

import (
	"log/slog"
)

type Logger struct {
	*slog.Logger
	level *slog.LevelVar
}

func New(cfg Config) *Logger {
	lv := new(slog.LevelVar)
	lv.Set(cfg.Level)

	targets := cfg.Targets
	if len(targets) == 0 {
		targets = []Target{{}}
	}

	handlers := make([]slog.Handler, 0, len(targets))
	for _, t := range targets {
		t = t.resolve()
		handlers = append(handlers, buildHandler(t, lv))
	}

	h := handlers[0]
	if len(handlers) > 1 {
		h = &fanoutHandler{handlers: handlers}
	}

	return &Logger{Logger: slog.New(h), level: lv}
}

func Init(cfg Config) *Logger {
	l := New(cfg)
	slog.SetDefault(l.Logger)
	return l
}

func (l *Logger) SetLevel(level slog.Level) {
	l.level.Set(level)
}

func buildHandler(t Target, level slog.Leveler) slog.Handler {
	opts := t.slogOptions(level)
	switch t.Format {
	case JSON:
		return slog.NewJSONHandler(t.Writer, &opts)
	case Text:
		return slog.NewTextHandler(t.Writer, &opts)
	default:
		return newHandler(t, level)
	}
}
