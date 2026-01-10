package config

import (
	"errors"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
)

var ErrInitTracer = errors.New("could not initialize jaeger tracer")

func InitTracer(appID, jaegerUri string) (io.Closer, error) {
	jaegerCfg := &config.Configuration{
		ServiceName: appID,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: jaegerUri,
		},
	}

	tracer, closer, err := jaegerCfg.NewTracer()
	if err != nil {
		return nil, errors.Join(ErrInitTracer, err)
	}
	opentracing.SetGlobalTracer(tracer)
	return closer, nil
}
