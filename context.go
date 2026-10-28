package caelum

import (
	"context"

	"github.com/Sn0wo2/caelum/internal/logger"
)

type Span = logger.Span

var ErrSpanMismatch = logger.ErrSpanMismatch

func ContextWithSpan(ctx context.Context, value *Span) context.Context {
	return logger.ContextWithSpan(ctx, value)
}

func SpanFromContext(ctx context.Context) (*Span, bool) {
	return logger.SpanFromContext(ctx)
}
