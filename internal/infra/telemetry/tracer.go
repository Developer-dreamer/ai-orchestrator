package telemetry

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"net/http"
)

func InitContextFromHttp(r *http.Request, operationName string) (opentracing.Span, context.Context) {
	tracer := opentracing.GlobalTracer()

	spanContext, _ := tracer.Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header),
	)

	span := tracer.StartSpan(
		operationName,
		opentracing.ChildOf(spanContext),
	)

	return span, opentracing.ContextWithSpan(r.Context(), span)
}

func InitContext(ctx context.Context, traceId, operationId string) (opentracing.Span, context.Context) {
	tracer := opentracing.GlobalTracer()
	spanContext, _ := tracer.Extract(
		opentracing.TextMap,
		opentracing.TextMapCarrier{"uber-trace-id": traceId},
	)

	span := tracer.StartSpan(
		operationId,
		opentracing.FollowsFrom(spanContext),
	)

	workerCtx := opentracing.ContextWithSpan(ctx, span)

	return span, workerCtx
}
