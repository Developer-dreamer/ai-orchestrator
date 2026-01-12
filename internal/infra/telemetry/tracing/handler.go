package tracing

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
)

type TraceHandler struct {
	slog.Handler
}

func (h TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		r.AddAttrs(slog.String("trace_id", span.SpanContext().TraceID().String()))
	}

	return h.Handler.Handle(ctx, r)
}
