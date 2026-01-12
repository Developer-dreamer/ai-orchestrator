package tracing

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
)

// TraceHandler is a slog.Handler wrapper that injects OpenTelemetry trace IDs
// from the context into log records before delegating to the underlying handler.
type TraceHandler struct {
	slog.Handler
}

// Handle enriches the provided slog.Record with the current trace ID, if a valid
// OpenTelemetry span is found in the context, and then forwards the record to
// the wrapped slog.Handler as part of the logging pipeline.
func (h TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		r.AddAttrs(slog.String("trace_id", span.SpanContext().TraceID().String()))
	}

	return h.Handler.Handle(ctx, r)
}
