package caelum

import (
	"io"
	"log/slog"
	"os"
)

type Format int

const (
	Compact Format = iota
	JSON
	Text
	Pretty
)

type PrettyOptions struct {
	Target     bool
	File       bool
	LineNumber bool
}

type Target struct {
	Writer  io.Writer
	Format  Format
	Handler slog.Handler
	Levels  []slog.Level

	HandlerOptions slog.HandlerOptions

	AddSource bool

	ShowSource bool

	Color      ColorDepth
	Style      *Style
	PathWidth  int
	TimeFormat string

	PrettyOptions *PrettyOptions
	Stacktrace    []slog.Level
}

type Config struct {
	Level   slog.Level
	Targets []Target
}

const (
	LevelError = slog.LevelError
	LevelWarn  = slog.LevelWarn
	LevelInfo  = slog.LevelInfo
	LevelDebug = slog.LevelDebug
	LevelTrace = slog.Level(-8)
)

func (t Target) resolve() Target {
	if t.Writer == nil {
		t.Writer = os.Stdout
	}

	t.AddSource = t.AddSource || t.ShowSource

	return t
}

func (t Target) slogOptions(level slog.Leveler) slog.HandlerOptions {
	opts := t.HandlerOptions
	if opts.Level == nil {
		opts.Level = level
	} else if level != nil {
		opts.Level = maxLeveler{global: level, target: opts.Level}
	}

	opts.AddSource = opts.AddSource || t.AddSource

	return opts
}

type maxLeveler struct {
	global slog.Leveler
	target slog.Leveler
}

func (l maxLeveler) Level() slog.Level {
	global, target := l.global.Level(), l.target.Level()
	if global > target {
		return global
	}

	return target
}
