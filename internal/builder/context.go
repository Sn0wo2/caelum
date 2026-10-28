package builder

import (
	"context"

	"github.com/Sn0wo2/caelum/internal/config"
	"github.com/Sn0wo2/caelum/internal/event"
	"github.com/Sn0wo2/caelum/internal/span"
)

func (l *Logger) ErrorContext(ctx context.Context, message string) error {
	return l.LogContext(ctx, config.LevelError, message)
}

func (l *Logger) WarnContext(ctx context.Context, message string) error {
	return l.LogContext(ctx, config.LevelWarn, message)
}

func (l *Logger) InfoContext(ctx context.Context, message string) error {
	return l.LogContext(ctx, config.LevelInfo, message)
}

func (l *Logger) DebugContext(ctx context.Context, message string) error {
	return l.LogContext(ctx, config.LevelDebug, message)
}

func (l *Logger) TraceContext(ctx context.Context, message string) error {
	return l.LogContext(ctx, config.LevelTrace, message)
}

func (l *Logger) EventContext(ctx context.Context, value event.Event) error {
	value, err := span.WithParent(ctx, l, value)
	if err != nil {
		return err
	}

	return l.Event(value)
}

func (l *Logger) CreateSpanContext(ctx context.Context, value event.Event) (*span.Span, error) {
	value, err := span.WithParent(ctx, l, value)
	if err != nil {
		return nil, err
	}

	id, err := l.CreateSpan(value)
	if err != nil {
		return nil, err
	}

	return span.NewSpan(l, id), nil
}
