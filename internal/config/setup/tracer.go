package setup

import (
	"context"
	"errors"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var ErrInitTracer = errors.New("could not initialize opentelemetry tracer")

func InitTracer(ctx context.Context, env, appID, otlpURI string) (func(context.Context) error, error) {
	switch env {
	case "production":
		return initProdTracer(ctx, appID)
	default:
		return initLocalTracer(ctx, appID, otlpURI)
	}
}

func initLocalTracer(ctx context.Context, appID, otlpURI string) (func(context.Context) error, error) {
	if otlpURI == "" {
		return nil, errors.New("you cannot pass empty parameter for otlpURI on local environment")
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(otlpURI),
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

func initProdTracer(ctx context.Context, appID string) (func(context.Context) error, error) {
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, errors.Join(ErrInitTracer, err)
	}

	res, err := resource.New(ctx,
		resource.WithDetectors(),
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

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
