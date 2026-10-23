package caelum

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/muesli/termenv"
	"github.com/rivo/uniseg"
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
	case termenv.Ascii:
		return NoColor
	default:
		return NoColor
	}
}

type Handler struct {
	out         io.Writer
	mu          *sync.Mutex
	style       Style
	depth       ColorDepth
	colors      palette
	timeFormat  string
	addSource   bool
	pathWidth   int
	sourceRoot  string
	level       slog.Leveler
	replaceAttr func(groups []string, a slog.Attr) slog.Attr
	stacktrace  []slog.Level
	pretty      bool
	prettyOpts  PrettyOptions

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

	PrettyOptions *PrettyOptions
	Stacktrace    []slog.Level
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

	style := mergeStyle(o.Style)

	h := &Handler{
		out:         w,
		mu:          &sync.Mutex{},
		style:       style,
		depth:       o.Color,
		colors:      makePalette(o.Color, style.Theme),
		timeFormat:  o.TimeFormat,
		addSource:   o.AddSource,
		pathWidth:   o.PathWidth,
		sourceRoot:  findSourceRoot(),
		level:       o.Level,
		replaceAttr: o.ReplaceAttr,
		stacktrace:  resolveStacktrace(o.Stacktrace),
	}
	if o.PrettyOptions != nil {
		h.pretty = true
		h.prettyOpts = *o.PrettyOptions
		h.addSource = h.addSource || h.prettyOpts.Target || h.prettyOpts.File || h.prettyOpts.LineNumber
	}

	return h
}

func newHandler(t Target, level slog.Leveler) *Handler {
	opts := HandlerOptions{
		HandlerOptions: t.slogOptions(level),
		Color:          t.Color,
		PathWidth:      t.PathWidth,
		TimeFormat:     t.TimeFormat,
		Stacktrace:     t.Stacktrace,
	}
	if t.Format == Pretty {
		opts.PrettyOptions = t.PrettyOptions
		if opts.PrettyOptions == nil {
			opts.PrettyOptions = &PrettyOptions{Target: true, File: true, LineNumber: true}
		}
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

	for _, attr := range attrsCopy {
		if isStacktraceAttr(attr) {
			h2.stacktrace = nil

			break
		}
	}

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

func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	st := &h.style
	p := &h.colors

	var b strings.Builder

	var timeAttr slog.Attr

	hasTime := !r.Time.IsZero()
	if hasTime {
		timeAttr, hasTime = h.replaceBuiltIn(slog.Time(slog.TimeKey, r.Time.Round(0)))
	}

	levelAttr, hasLevel := h.replaceBuiltIn(slog.Any(slog.LevelKey, r.Level))
	h.writeHeader(&b, timeAttr, hasTime, levelAttr, hasLevel, r.Level)
	needsSpace := hasTime || hasLevel

	var sourceAttr slog.Attr

	hasSource := false

	if h.addSource {
		sourceValue := r.Source()
		if sourceValue == nil {
			sourceValue = &slog.Source{}
		}

		sourceAttr, hasSource = h.replaceBuiltIn(slog.Any(slog.SourceKey, sourceValue))
		if hasSource && !h.pretty {
			if file, line, ok := sourcePath(sourceAttr.Value); ok {
				if needsSpace {
					b.WriteByte(' ')
				}

				if st.Icons.Name == "nerd" {
					p.accentDimmed.paint(&b, "")
					b.WriteByte(' ')
				}

				p.textDimmed.paint(&b, formatPath(file, line, h.pathWidth, h.sourceRoot))
				b.WriteByte(' ')
				p.secondary.paint(&b, st.Icons.Arrow)

				needsSpace = true
			}
		}
	}

	if msg, ok := h.replaceBuiltIn(slog.String(slog.MessageKey, r.Message)); ok {
		message := compactMessage(msg.Value)
		if message != "" {
			if needsSpace {
				b.WriteByte(' ')
			}

			p.text.paint(&b, message)

			needsSpace = true
		}
	}

	groups := []string(nil)

	for _, goa := range h.goas {
		if goa.group != "" {
			groups = append(groups, goa.group)

			continue
		}

		h.writeAttrs(&b, groups, goa.attrs, &needsSpace)
	}

	r.Attrs(func(a slog.Attr) bool {
		h.writeAttr(&b, groups, a, &needsSpace)

		return true
	})

	if h.pretty && hasSource {
		h.writePrettySource(&b, sourceAttr.Value, needsSpace)
	}

	b.WriteByte('\n')

	if stack := stacktraceFor(h.stacktrace, r); stack != "" {
		if attr, ok := h.replaceBuiltIn(slog.String("stacktrace", stack)); ok {
			stack = attr.Value.String()

			b.WriteString("Stacktrace:\n")
			b.WriteString(stack)

			if !strings.HasSuffix(stack, "\n") {
				b.WriteByte('\n')
			}
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	n, err := io.WriteString(h.out, b.String())
	if err == nil && n != b.Len() {
		err = io.ErrShortWrite
	}

	return err
}

func (h *Handler) writeHeader(b *strings.Builder, time slog.Attr, hasTime bool, level slog.Attr, hasLevel bool, l slog.Level) {
	icons := h.style.Icons
	p := &h.colors
	index := levelIndex(l)
	label := levelLabel(level.Value, [5]string{
		h.style.Labels.Error, h.style.Labels.Warn, h.style.Labels.Info,
		h.style.Labels.Debug, h.style.Labels.Trace,
	}[index])
	colors := p.levels[index]
	nerd := icons.Name == "nerd"
	frameSpace, iconSpace := "", ""
	clockIcon, levelIcon := "", ""

	if icons.Name == "unicode" || nerd {
		frameSpace = " "
	}

	if nerd {
		iconSpace = " "
		clockIcon = ""
		levelIcon = [5]string{"", "", "", "", ""}[index]
	}

	if nerd && h.colors.timeBody != "" && hasTime && hasLevel {
		p.timeOutline.paint(b, icons.TimeBracketOpen)
		p.timeBody.paint(b, " "+clockIcon+" "+formatTime(time.Value, h.timeFormat)+" ")
		colors.transition.paint(b, icons.Separator)
		colors.body.paint(b, " "+levelIcon+" "+label+" ")
		colors.close.paint(b, icons.TimeBracketClose)
		b.WriteByte(' ')

		return
	}

	timeOpen, timeClose, separator := icons.TimeBracketOpen, icons.TimeBracketClose, icons.Separator
	if nerd {
		timeOpen, timeClose, separator = "❬", "❭", "┊"
	}

	if hasTime {
		p.accent.paint(b, timeOpen)
		b.WriteString(frameSpace)
		p.text.paint(b, clockIcon+iconSpace+formatTime(time.Value, h.timeFormat))

		if hasLevel {
			b.WriteByte(' ')
			p.accentDimmed.paint(b, separator)
			b.WriteByte(' ')
		}
	}

	if hasLevel {
		colors.outline.paint(b, icons.BracketOpen)
		colors.body.paint(b, levelIcon+iconSpace+label)
		colors.outline.paint(b, icons.BracketClose)
	}

	if hasTime {
		b.WriteByte(' ')
		p.accent.paint(b, timeClose)
		b.WriteString(frameSpace)
	}
}

func (h *Handler) writePrettySource(b *strings.Builder, value slog.Value, needsSpace bool) {
	var parts []string

	if file, line, ok := sourcePath(value); ok {
		if h.prettyOpts.File {
			if !h.prettyOpts.LineNumber {
				line = -1
			}

			parts = append(parts, strings.TrimLeft(formatPath(file, line, h.pathWidth, h.sourceRoot), " "))
		} else if h.prettyOpts.LineNumber {
			parts = append(parts, "line "+strconv.Itoa(line))
		}
	}

	if h.prettyOpts.Target && value.Kind() == slog.KindAny {
		var function string

		switch source := value.Any().(type) {
		case *slog.Source:
			if source != nil {
				function = source.Function
			}
		case slog.Source:
			function = source.Function
		}

		if function != "" {
			parts = append(parts, function)
		}
	}

	if len(parts) == 0 {
		return
	}

	if needsSpace {
		b.WriteByte('\n')
	}

	b.WriteString("  ")
	h.colors.textDimmed.paint(b, "at "+strings.Join(parts, " in "))
}

func (h *Handler) replaceBuiltIn(a slog.Attr) (slog.Attr, bool) {
	a.Value = a.Value.Resolve()
	if h.replaceAttr != nil {
		a = h.replaceAttr(nil, a)
		a.Value = a.Value.Resolve()
	}

	return a, !a.Equal(slog.Attr{})
}

func (h *Handler) writeAttrs(b *strings.Builder, groups []string, attrs []slog.Attr, needsSpace *bool) {
	for _, a := range attrs {
		h.writeAttr(b, groups, a, needsSpace)
	}
}

func (h *Handler) writeAttr(b *strings.Builder, groups []string, a slog.Attr, needsSpace *bool) {
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

		h.writeAttrs(b, nextGroups, a.Value.Group(), needsSpace)

		return
	}

	key := strings.Join(append(append([]string(nil), groups...), a.Key), ".")

	if *needsSpace {
		if h.pretty {
			b.WriteString("\n  ")
		} else {
			b.WriteByte(' ')
		}
	} else if h.pretty {
		b.WriteString("  ")
	}

	h.colors.secondary.paint(b, compactToken(key))
	h.colors.accent.paint(b, "=")
	h.colors.text.paint(b, compactValue(a.Value))

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
	if strings.IndexFunc(s, unicode.IsControl) >= 0 {
		return strconv.Quote(s)
	}

	return s
}

func formatPath(file string, line, width int, sourceRoot string) string {
	if width <= 0 {
		return ""
	}

	p := strings.ReplaceAll(file, `\`, "/")
	if sourceRoot != "" {
		if rel, err := filepath.Rel(sourceRoot, file); err == nil {
			rel = strings.ReplaceAll(rel, `\`, "/")
			if rel != "." && rel != ".." && !strings.HasPrefix(rel, "../") {
				p = rel
			}
		}
	}

	full := p
	if line >= 0 {
		full += ":" + strconv.Itoa(line)
	}

	start := 0
	ellipsis := ""

	textWidth := uniseg.StringWidth(full)
	if textWidth > width {
		start = strings.LastIndexByte(full, '/') + 1
		filenameStart := start

		textWidth = uniseg.StringWidth(full[start:])
		for start > 0 {
			previous := strings.LastIndexByte(full[:start-1], '/') + 1

			candidateWidth := textWidth + uniseg.StringWidth(full[previous:start])
			if candidateWidth > width {
				break
			}

			start, textWidth = previous, candidateWidth
		}

		if start == filenameStart {
			ellipsis = "…"
			start = 0
			textWidth = uniseg.StringWidth(full)

			graphemes := uniseg.NewGraphemes(full)
			for textWidth > width-1 && graphemes.Next() {
				_, start = graphemes.Positions()
				textWidth -= graphemes.Width()
			}
		}
	}

	return strings.Repeat(" ", max(width-textWidth-uniseg.StringWidth(ellipsis), 0)) + ellipsis + full[start:]
}
