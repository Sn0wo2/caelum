package caelum

import (
	"context"

	"github.com/Sn0wo2/caelum/internal/span"
)

type Span = span.Span

var ErrSpanMismatch = span.ErrSpanMismatch

func ContextWithSpan(ctx context.Context, value *Span) context.Context {
	return span.ContextWithSpan(ctx, value)
}

func SpanFromContext(ctx context.Context) (*Span, bool) {
	return span.SpanFromContext(ctx)
}
