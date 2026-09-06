package caelum

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/muesli/termenv"
)

// detectDepth probes a writer's real terminal capabilities and degrades the
// color depth accordingly, mirroring acta's supports-color behavior across all
// four tiers. termenv inspects COLORTERM, TERM, and TTY-ness (and honors
// NO_COLOR); anything that isn't a terminal — files, pipes, buffers — collapses
// to NoColor. Async wrappers unwrap to their underlying writer so a wrapped
// stdout is still detected as a terminal.
func detectDepth(w io.Writer) ColorDepth {
	for {
		u, ok := w.(interface{ Unwrap() io.Writer })
		if !ok {
			break
		}
		w = u.Unwrap()
	}

	switch termenv.NewOutput(w).Profile {
	case termenv.TrueColor:
		return TrueColor
	case termenv.ANSI256:
		return Ansi256
	case termenv.ANSI:
		return Ansi16
	default: // termenv.Ascii
		return NoColor
	}
}

// handler is a slog.Handler that renders the compact themed line. Style is
// fixed per target at construction; the mutex only guards the final write so
// concurrent goroutines don't interleave bytes.
type handler struct {
	out        io.Writer
	mu         *sync.Mutex
	style      Style
	depth      ColorDepth
	timeFormat string
	showSource bool
	pathWidth  int
	sourceRoot string
	level      slog.Leveler

	// goas accumulates groups and attrs added via WithGroup / WithAttrs.
	goas []groupOrAttrs
}

type groupOrAttrs struct {
	group string
	attrs []slog.Attr
}

func newHandler(t Target, level slog.Leveler) *handler {
	sourceRoot, _ := os.Getwd()
	for dir := sourceRoot; dir != ""; {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			sourceRoot = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return &handler{
		out:        t.Writer,
		mu:         &sync.Mutex{},
		style:      *t.Style,
		depth:      t.Color,
		timeFormat: t.TimeFormat,
		showSource: t.ShowSource,
		pathWidth:  t.PathWidth,
		sourceRoot: sourceRoot,
		level:      level,
	}
}

func (h *handler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	h2 := *h
	h2.goas = append(append([]groupOrAttrs(nil), h.goas...), groupOrAttrs{attrs: attrs})
	return &h2
}

func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := *h
	h2.goas = append(append([]groupOrAttrs(nil), h.goas...), groupOrAttrs{group: name})
	return &h2
}

func (h *handler) levelColor(st *Style, l slog.Level) (RGB, string) {
	switch {
	case l >= slog.LevelError:
		return st.Theme.Error, st.Labels.Error
	case l >= slog.LevelWarn:
		return st.Theme.Warn, st.Labels.Warn
	case l >= slog.LevelInfo:
		return st.Theme.Info, st.Labels.Info
	case l >= slog.LevelDebug:
		return st.Theme.Debug, st.Labels.Debug
	default:
		return st.Theme.Trace, st.Labels.Trace
	}
}

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	st := &h.style
	d := h.depth
	var b strings.Builder

	accent := st.Theme.Accent
	text := st.Theme.Text
	levelRGB, label := h.levelColor(st, r.Level)

	// ｢ time ┇ ［LEVEL］｣ — timestamps in UTC, matching acta.
	d.paint(&b, accent, false, st.Icons.TimeBracketOpen)
	d.paint(&b, text, false, r.Time.UTC().Format(h.timeFormat))
	b.WriteByte(' ')
	d.paint(&b, accent.dim(), false, st.Icons.Separator)
	b.WriteByte(' ')

	// Nerd icons look better with a foreground-only bracket; the unicode set
	// uses an inverted (on-background) level chip, matching acta.
	bracketBg := st.Icons.Name != "nerd"
	d.paint(&b, levelRGB, bracketBg, st.Icons.BracketOpen)
	d.paint(&b, levelRGB, true, label)
	d.paint(&b, levelRGB, bracketBg, st.Icons.BracketClose)
	b.WriteByte(' ')
	d.paint(&b, accent, false, st.Icons.TimeBracketClose)
	b.WriteByte(' ')

	// Optional source column: right-aligned dimmed path + accent arrow.
	if h.showSource && r.PC != 0 {
		f, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
		file := f.File
		if file == "" {
			file = "?"
		}
		d.paint(&b, text.dim(), false, formatPath(file, f.Line, h.pathWidth, h.sourceRoot))
		b.WriteByte(' ')
		d.paint(&b, accent, false, st.Icons.Arrow)
		b.WriteByte(' ')
	}

	if r.Message != "" {
		d.paint(&b, text, false, r.Message)
	}

	var writeAttr func(prefix string, a slog.Attr)
	writeAttr = func(prefix string, a slog.Attr) {
		a.Value = a.Value.Resolve()
		if a.Value.Kind() == slog.KindGroup {
			if a.Key != "" {
				prefix += a.Key + "."
			}
			for _, ga := range a.Value.Group() {
				writeAttr(prefix, ga)
			}
			return
		}
		if a.Equal(slog.Attr{}) {
			return
		}
		b.WriteByte(' ')
		d.paint(&b, st.Theme.Secondary, false, prefix+a.Key)
		d.paint(&b, accent, false, "=")
		d.paint(&b, text, false, a.Value.String())
	}

	// Groups only prefix attrs added after them, per the slog contract.
	prefix := ""
	for _, goa := range h.goas {
		if goa.group != "" {
			prefix += goa.group + "."
			continue
		}
		for _, a := range goa.attrs {
			writeAttr(prefix, a)
		}
	}
	r.Attrs(func(a slog.Attr) bool {
		writeAttr(prefix, a)
		return true
	})

	b.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.out, b.String())
	return err
}

// formatPath makes source files project-relative before right-aligning them.
// Paths outside the project retain the existing src/tail fallback behavior.
func formatPath(file string, line, width int, sourceRoot string) string {
	p := filepath.ToSlash(file)
	if rel, err := filepath.Rel(sourceRoot, file); err == nil {
		rel = filepath.ToSlash(rel)
		if rel != "." && rel != ".." && !strings.HasPrefix(rel, "../") {
			p = rel
		}
	}
	if _, tail, ok := strings.Cut(p, "src/"); ok {
		p = tail
	}
	full := p + ":" + strconv.Itoa(line)
	if len(full) <= width {
		return fmt.Sprintf("%*s", width, full)
	}

	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		fileWithLine := p[i+1:] + ":" + strconv.Itoa(line)
		if len(fileWithLine)+2 <= width {
			dir := p[:i]
			start := max(len(dir)-(width-len(fileWithLine)-1), 0)
			for start < len(dir) && !utf8.RuneStart(dir[start]) {
				start++
			}
			tail := dir[start:]
			// Drop a partially-cut leading segment unless the cut landed on '/'.
			if start > 0 && dir[start-1] != '/' {
				if j := strings.IndexByte(tail, '/'); j >= 0 {
					tail = tail[j+1:]
				}
			}
			return fmt.Sprintf("%*s", width, tail+"/"+fileWithLine)
		}
	}

	adj := max(len(full)-(width-1), 0)
	for adj < len(full) && !utf8.RuneStart(full[adj]) {
		adj++
	}
	return "…" + full[adj:]
}
