package setup

import (
	"context"
	"errors"

	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var ErrInitTracer = errors.New("could not initialize opentelemetry tracer")

func InitTracer(appID, jaegerUri string) (func(context.Context) error, error) {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(jaegerUri),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, errors.Join(ErrInitTracer, err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(appID),
		),
	)
	if err != nil {
		return nil, errors.Join(ErrInitTracer, err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(jaeger.Jaeger{})

	return tp.Shutdown, nil
}
