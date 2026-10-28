package logger

import (
	"context"

	"github.com/Sn0wo2/caelum/internal/config"
)

func (l *Logger) ErrorContext(ctx context.Context, message string, fields ...Field) error {
	return l.LogContext(ctx, config.LevelError, message, fields...)
}

func (l *Logger) WarnContext(ctx context.Context, message string, fields ...Field) error {
	return l.LogContext(ctx, config.LevelWarn, message, fields...)
}

func (l *Logger) InfoContext(ctx context.Context, message string, fields ...Field) error {
	return l.LogContext(ctx, config.LevelInfo, message, fields...)
}

func (l *Logger) DebugContext(ctx context.Context, message string, fields ...Field) error {
	return l.LogContext(ctx, config.LevelDebug, message, fields...)
}

func (l *Logger) TraceContext(ctx context.Context, message string, fields ...Field) error {
	return l.LogContext(ctx, config.LevelTrace, message, fields...)
}

func (l *Logger) EventContext(ctx context.Context, value Event) error {
	return l.Event(withParent(ctx, value))
}

type spanContextKey struct{}

func ContextWithSpan(ctx context.Context, value *Span) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, spanContextKey{}, value)
}

func SpanFromContext(ctx context.Context) (*Span, bool) {
	if ctx == nil {
		return nil, false
	}

	value, ok := ctx.Value(spanContextKey{}).(*Span)

	return value, ok && value != nil
}

func withParent(ctx context.Context, value Event) Event {
	if value.Parent == nil {
		value.Parent, _ = SpanFromContext(ctx)
	}

	return value
}
