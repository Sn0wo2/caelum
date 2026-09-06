package caelum

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/muesli/termenv"
)

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
	default:
		return NoColor
	}
}

type Handler struct {
	out         io.Writer
	mu          *sync.Mutex
	style       Style
	depth       ColorDepth
	timeFormat  string
	addSource   bool
	pathWidth   int
	sourceRoot  string
	level       slog.Leveler
	replaceAttr func(groups []string, a slog.Attr) slog.Attr

	goas []groupOrAttrs
}

var _ slog.Handler = (*Handler)(nil)

type groupOrAttrs struct {
	group string
	attrs []slog.Attr
}

type HandlerOptions struct {
	slog.HandlerOptions

	Color      ColorDepth
	Style      Style
	PathWidth  int
	TimeFormat string
}

func NewHandler(w io.Writer, opts *HandlerOptions) *Handler {
	if w == nil {
		w = os.Stdout
	}

	var o HandlerOptions
	if opts != nil {
		o = *opts
	}
	if o.Level == nil {
		o.Level = slog.LevelInfo
	}
	if o.TimeFormat == "" {
		o.TimeFormat = "15:04:05"
	}
	if o.PathWidth <= 0 {
		o.PathWidth = 40
	}
	if o.Color == ColorAuto {
		o.Color = detectDepth(w)
	}

	return &Handler{
		out:         w,
		mu:          &sync.Mutex{},
		style:       mergeStyle(o.Style),
		depth:       o.Color,
		timeFormat:  o.TimeFormat,
		addSource:   o.AddSource,
		pathWidth:   o.PathWidth,
		sourceRoot:  findSourceRoot(),
		level:       o.Level,
		replaceAttr: o.ReplaceAttr,
	}
}

func newHandler(t Target, level slog.Leveler) *Handler {
	opts := HandlerOptions{
		HandlerOptions: t.slogOptions(level),
		Color:          t.Color,
		PathWidth:      t.PathWidth,
		TimeFormat:     t.TimeFormat,
	}
	if t.Style != nil {
		opts.Style = *t.Style
	}
	return NewHandler(t.Writer, &opts)
}

func findSourceRoot() string {
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
	return sourceRoot
}

func (h *Handler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	h2 := *h
	attrsCopy := append([]slog.Attr(nil), attrs...)
	h2.goas = append(append([]groupOrAttrs(nil), h.goas...), groupOrAttrs{attrs: attrsCopy})
	return &h2
}

func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := *h
	h2.goas = append(append([]groupOrAttrs(nil), h.goas...), groupOrAttrs{group: name})
	return &h2
}

func (h *Handler) levelColor(st *Style, l slog.Level) (RGB, string) {
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

func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	st := &h.style
	d := h.depth
	var b strings.Builder

	accent := st.Theme.Accent
	text := st.Theme.Text
	levelRGB, defaultLabel := h.levelColor(st, r.Level)

	var timeAttr slog.Attr
	hasTime := !r.Time.IsZero()
	if hasTime {
		timeAttr, hasTime = h.replaceBuiltIn(slog.Time(slog.TimeKey, r.Time.Round(0)), nil)
	}
	levelAttr, hasLevel := h.replaceBuiltIn(slog.Any(slog.LevelKey, r.Level), nil)

	if hasTime {
		d.paint(&b, accent, false, st.Icons.TimeBracketOpen)
		d.paint(&b, text, false, formatTime(timeAttr.Value, h.timeFormat))
		b.WriteByte(' ')
		d.paint(&b, accent.dim(), false, st.Icons.Separator)
		if hasLevel {
			b.WriteByte(' ')
		}
	}

	if hasLevel {
		bracketBg := st.Icons.Name != "nerd"
		label := levelLabel(levelAttr.Value, defaultLabel)
		d.paint(&b, levelRGB, bracketBg, st.Icons.BracketOpen)
		d.paint(&b, levelRGB, true, label)
		d.paint(&b, levelRGB, bracketBg, st.Icons.BracketClose)
	}
	if hasTime {
		d.paint(&b, accent, false, st.Icons.TimeBracketClose)
	}
	needsSpace := hasTime || hasLevel

	if h.addSource {
		sourceValue := r.Source()
		if sourceValue == nil {
			sourceValue = &slog.Source{}
		}
		if source, ok := h.replaceBuiltIn(slog.Any(slog.SourceKey, sourceValue), nil); ok {
			if file, line, ok := sourcePath(source.Value); ok {
				if needsSpace {
					b.WriteByte(' ')
				}
				d.paint(&b, text.dim(), false, formatPath(file, line, h.pathWidth, h.sourceRoot))
				b.WriteByte(' ')
				d.paint(&b, accent, false, st.Icons.Arrow)
				needsSpace = true
			}
		}
	}

	if msg, ok := h.replaceBuiltIn(slog.String(slog.MessageKey, r.Message), nil); ok {
		message := compactMessage(msg.Value)
		if message != "" {
			if needsSpace {
				b.WriteByte(' ')
			}
			d.paint(&b, text, false, message)
			needsSpace = true
		}
	}

	groups := []string(nil)
	for _, goa := range h.goas {
		if goa.group != "" {
			groups = append(groups, goa.group)
			continue
		}
		h.writeAttrs(&b, d, st, groups, goa.attrs, &needsSpace)
	}
	r.Attrs(func(a slog.Attr) bool {
		h.writeAttr(&b, d, st, groups, a, &needsSpace)
		return true
	})

	b.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.out, b.String())
	return err
}

func (h *Handler) replaceBuiltIn(a slog.Attr, groups []string) (slog.Attr, bool) {
	a.Value = a.Value.Resolve()
	if h.replaceAttr != nil {
		a = h.replaceAttr(groups, a)
		a.Value = a.Value.Resolve()
	}
	return a, !a.Equal(slog.Attr{})
}

func (h *Handler) writeAttrs(b *strings.Builder, d ColorDepth, st *Style, groups []string, attrs []slog.Attr, needsSpace *bool) {
	for _, a := range attrs {
		h.writeAttr(b, d, st, groups, a, needsSpace)
	}
}

func (h *Handler) writeAttr(b *strings.Builder, d ColorDepth, st *Style, groups []string, a slog.Attr, needsSpace *bool) {
	a.Value = a.Value.Resolve()
	if a.Value.Kind() != slog.KindGroup && h.replaceAttr != nil {
		a = h.replaceAttr(groups, a)
		a.Value = a.Value.Resolve()
	}
	if a.Equal(slog.Attr{}) {
		return
	}

	if a.Value.Kind() == slog.KindGroup {
		nextGroups := groups
		if a.Key != "" {
			nextGroups = append(append([]string(nil), groups...), a.Key)
		}
		h.writeAttrs(b, d, st, nextGroups, a.Value.Group(), needsSpace)
		return
	}

	key := strings.Join(append(append([]string(nil), groups...), a.Key), ".")
	if *needsSpace {
		b.WriteByte(' ')
	}
	d.paint(b, st.Theme.Secondary, false, compactToken(key))
	d.paint(b, st.Theme.Accent, false, "=")
	d.paint(b, st.Theme.Text, false, compactValue(a.Value))
	*needsSpace = true
}

func formatTime(v slog.Value, layout string) string {
	if v.Kind() == slog.KindTime {
		return v.Time().Format(layout)
	}
	return compactValue(v)
}

func levelLabel(v slog.Value, fallback string) string {
	v = v.Resolve()
	if v.Kind() == slog.KindString {
		return v.String()
	}
	if v.Kind() == slog.KindAny {
		if _, ok := v.Any().(slog.Level); ok {
			return fallback
		}
	}
	if value := compactValue(v); value != "" {
		return value
	}
	return fallback
}

func sourcePath(v slog.Value) (string, int, bool) {
	v = v.Resolve()
	if v.Kind() == slog.KindAny {
		switch source := v.Any().(type) {
		case *slog.Source:
			if source == nil {
				return "", 0, false
			}
			return source.File, source.Line, source.File != ""
		case slog.Source:
			return source.File, source.Line, source.File != ""
		}
	}
	if v.Kind() == slog.KindString {
		return v.String(), 0, v.String() != ""
	}
	return "", 0, false
}

func compactMessage(v slog.Value) string {
	v = v.Resolve()
	if v.Kind() == slog.KindString {
		return escapeControl(v.String())
	}
	return escapeControl(compactValue(v))
}

func compactValue(v slog.Value) string {
	v = v.Resolve()
	if v.Kind() == slog.KindTime {
		return v.Time().Format(time.RFC3339Nano)
	}
	return compactToken(v.String())
}

func compactToken(s string) string {
	if s == "" || strings.IndexFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r) || strings.ContainsRune(`=\\"`, r)
	}) >= 0 {
		return strconv.Quote(s)
	}
	return s
}

func escapeControl(s string) string {
	if strings.IndexFunc(s, func(r rune) bool { return unicode.IsControl(r) }) >= 0 {
		return strconv.Quote(s)
	}
	return s
}

func formatPath(file string, line, width int, sourceRoot string) string {
	p := filepath.ToSlash(file)
	if rel, err := filepath.Rel(sourceRoot, file); err == nil {
		rel = filepath.ToSlash(rel)
		if rel != "." && rel != ".." && !strings.HasPrefix(rel, "../") {
			p = rel
		}
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
