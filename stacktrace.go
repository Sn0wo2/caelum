package caelum

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

func levelIndex(level slog.Level) int {
	switch {
	case level >= slog.LevelError:
		return 0
	case level >= slog.LevelWarn:
		return 1
	case level >= slog.LevelInfo:
		return 2
	case level >= slog.LevelDebug:
		return 3
	default:
		return 4
	}
}

func resolveStacktrace(levels []slog.Level) []slog.Level {
	if levels == nil {
		return []slog.Level{slog.LevelError}
	}

	return append([]slog.Level{}, levels...)
}

func stacktraceFor(levels []slog.Level, r slog.Record) string {
	bucket := levelIndex(r.Level)
	matched := false

	for _, level := range levels {
		if levelIndex(level) == bucket {
			matched = true

			break
		}
	}

	if !matched {
		return ""
	}

	r.Attrs(func(attr slog.Attr) bool {
		if isStacktraceAttr(attr) {
			matched = false

			return false
		}

		return true
	})

	if !matched {
		return ""
	}

	pcs := make([]uintptr, 64)
	n := runtime.Callers(0, pcs)
	frames := make([]runtime.Frame, 0, n)

	iter := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := iter.Next()
		frames = append(frames, frame)

		if !more {
			break
		}
	}

	start := -1

	source := r.Source()
	if source != nil {
		lineMatch := -1

		for i, frame := range frames {
			if frame.File == source.File && frame.Line == source.Line {
				if source.Function == "" || frame.Function == source.Function {
					start = i

					break
				}

				if lineMatch < 0 {
					lineMatch = i
				}
			}
		}

		if start < 0 {
			start = lineMatch
		}

		if start < 0 {
			pcFrames := runtime.CallersFrames([]uintptr{r.PC})

			frame, _ := pcFrames.Next()
			if frame.File != "" || frame.Function != "" {
				frames = []runtime.Frame{frame}
				start = 0
			}
		}
	}

	if start < 0 {
		for i, frame := range frames {
			if !internalStackFrame(frame.Function) {
				start = i

				break
			}
		}
	}

	if start < 0 {
		return ""
	}

	var b strings.Builder

	for _, frame := range frames[start:] {
		if internalStackFrame(frame.Function) || (frame.Function == "" && frame.File == "") {
			continue
		}

		if b.Len() > 0 {
			b.WriteByte('\n')
		}

		function := frame.Function
		if function == "" {
			function = "<unknown>"
		}

		fmt.Fprintf(&b, "%s\n\t%s:%d", function, frame.File, frame.Line)
	}

	return b.String()
}

func isStacktraceAttr(attr slog.Attr) bool {
	if attr.Key == "stacktrace" {
		return true
	}

	if attr.Key == "" && attr.Value.Kind() == slog.KindGroup {
		for _, child := range attr.Value.Group() {
			if isStacktraceAttr(child) {
				return true
			}
		}
	}

	return false
}

func internalStackFrame(function string) bool {
	return strings.HasPrefix(function, "runtime.") ||
		strings.HasPrefix(function, "log/slog.") ||
		function == "github.com/Sn0wo2/caelum.stacktraceFor" ||
		function == "github.com/Sn0wo2/caelum.(*stacktraceHandler).Handle" ||
		function == "github.com/Sn0wo2/caelum.(*Handler).Handle"
}

func withStacktrace(h slog.Handler, levels []slog.Level) slog.Handler {
	resolved := resolveStacktrace(levels)
	if len(resolved) == 0 {
		return h
	}

	return &stacktraceHandler{handler: h, levels: resolved}
}

type stacktraceHandler struct {
	handler slog.Handler
	levels  []slog.Level
}

func (h *stacktraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *stacktraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if stacktrace := stacktraceFor(h.levels, r); stacktrace != "" {
		r = r.Clone()
		r.AddAttrs(slog.String("stacktrace", stacktrace))
	}

	return h.handler.Handle(ctx, r)
}

func (h *stacktraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	for _, attr := range attrs {
		if isStacktraceAttr(attr) {
			return h.handler.WithAttrs(attrs)
		}
	}

	return &stacktraceHandler{handler: h.handler.WithAttrs(attrs), levels: h.levels}
}

func (h *stacktraceHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &stacktraceHandler{handler: h.handler.WithGroup(name), levels: h.levels}
}
