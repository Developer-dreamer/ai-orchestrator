package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func getTracer() trace.Tracer {
	return otel.Tracer("default-tracer")
}

func InitContextFromHttp(r *http.Request, operationName string) (trace.Span, context.Context) {
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	ctx, span := getTracer().Start(ctx, operationName)

	return span, ctx
}

func InitContext(ctx context.Context, traceId, operationId string) (trace.Span, context.Context) {
	carrier := propagation.MapCarrier{
		"uber-trace-id": traceId,
	}

	extractedCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)
	newCtx, span := getTracer().Start(extractedCtx, operationId)

	return span, newCtx
}
