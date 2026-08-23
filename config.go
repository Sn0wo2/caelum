package caelum

import (
	"io"
	"log/slog"
	"os"
)

// Format selects how each record is rendered.
type Format int

const (
	// Compact is the themed, colorized single-line format (the default).
	Compact Format = iota
	// JSON emits one JSON object per line via slog's built-in JSONHandler.
	JSON
	// Text emits slog's built-in key=value TextHandler.
	Text
)

// Target describes where a writer sends its output and how it is rendered.
type Target struct {
	// Writer is the destination. Defaults to os.Stdout when nil.
	Writer io.Writer
	// Format selects the rendering. Defaults to Compact.
	Format Format
	// Color selects the depth. The zero value ColorAuto detects it from the
	// writer (TrueColor for a capable TTY, NoColor otherwise).
	Color ColorDepth
	// Style overrides the visual style for this target. A partially-filled
	// Style keeps defaults for its zero-value sections, so
	// &Style{Theme: ThemeNord} works. Nil falls back to DefaultStyle().
	Style *Style
	// ShowSource adds a right-aligned source column (path:line). For JSON and
	// Text targets it maps to slog's AddSource.
	ShowSource bool
	// PathWidth is the source column width in the Compact format. Defaults to 40.
	PathWidth int
	// TimeFormat is a Go reference-time layout for the UTC timestamp.
	// Defaults to "15:04:05".
	TimeFormat string
}

// Config is the top-level configuration passed to New / Init.
type Config struct {
	// Level is the minimum severity that will be logged. Defaults to LevelInfo.
	Level slog.Level
	// Targets are the destinations. When empty, a single stdout Compact target
	// is used.
	Targets []Target
}

// Level constants re-exported so callers don't need to import log/slog just to
// set a level. They are plain slog.Level values.
const (
	LevelError = slog.LevelError
	LevelWarn  = slog.LevelWarn
	LevelInfo  = slog.LevelInfo
	LevelDebug = slog.LevelDebug
	// LevelTrace sits below Debug, following slog's convention of spacing
	// custom levels four apart.
	LevelTrace = slog.Level(-8)
)

// resolve fills in defaults for a target.
func (t Target) resolve() Target {
	if t.Writer == nil {
		t.Writer = os.Stdout
	}
	if t.TimeFormat == "" {
		t.TimeFormat = "15:04:05"
	}
	if t.PathWidth == 0 {
		t.PathWidth = 40
	}
	if t.Color == ColorAuto {
		t.Color = detectDepth(t.Writer)
	}
	s := DefaultStyle()
	if t.Style != nil {
		if t.Style.Theme != (Theme{}) {
			s.Theme = t.Style.Theme
		}
		if t.Style.Icons != (Icons{}) {
			s.Icons = t.Style.Icons
		}
		if t.Style.Labels != (LevelLabels{}) {
			s.Labels = t.Style.Labels
		}
	}
	t.Style = &s
	return t
}
