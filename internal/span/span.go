package span

import (
	"context"
	"errors"

	"github.com/Sn0wo2/caelum/internal/event"
)

var ErrSpanMismatch = errors.New("caelum: span belongs to a different logger")

type SpanLogger interface {
	RecordSpan(id uint32, fields map[string]any) error
	CloseSpan(id uint32) error
}

type Span struct {
	logger SpanLogger
	id     uint32
}

type spanContextKey struct{}

func NewSpan(logger SpanLogger, id uint32) *Span {
	return &Span{logger: logger, id: id}
}

func ContextWithSpan(ctx context.Context, span *Span) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, spanContextKey{}, span)
}

func SpanFromContext(ctx context.Context) (*Span, bool) {
	if ctx == nil {
		return nil, false
	}

	span, ok := ctx.Value(spanContextKey{}).(*Span)

	return span, ok && span != nil
}

func (s *Span) ID() uint32 {
	if s == nil {
		return 0
	}

	return s.id
}

func (s *Span) Context(ctx context.Context) context.Context {
	return ContextWithSpan(ctx, s)
}

func (s *Span) Record(fields map[string]any) error {
	if s == nil || s.logger == nil {
		return errors.New("caelum: invalid span")
	}

	return s.logger.RecordSpan(s.id, fields)
}

func (s *Span) Close() error {
	if s == nil || s.logger == nil {
		return errors.New("caelum: invalid span")
	}

	return s.logger.CloseSpan(s.id)
}

func WithParent(ctx context.Context, logger SpanLogger, value event.Event) (event.Event, error) {
	if value.Parent != nil {
		return value, nil
	}

	span, ok := SpanFromContext(ctx)
	if !ok {
		return value, nil
	}

	if span.logger != logger {
		return event.Event{}, ErrSpanMismatch
	}

	value.Parent = &span.id

	return value, nil
}
