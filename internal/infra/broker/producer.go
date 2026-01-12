package broker

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/config/env"
	"context"
	"encoding/json"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
)

type Producer struct {
	logger   common.Logger
	client   *redis.Client
	config   *config.Stream
	streamID string
}

func NewProducer(l common.Logger, client *redis.Client, streamCfg *config.Stream, cfg *env.APIConfig) *Producer {
	return &Producer{logger: l, client: client, config: streamCfg, streamID: cfg.RedisStreamID}
}

func (p *Producer) Publish(ctx context.Context, data any) error {
	dataStr, err := json.Marshal(data)
	if err != nil {
		p.logger.Error("Failed to convert data to json.", "error", err, "data", data)
		return err
	}

	carrier := make(map[string]string)
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		err := opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.TextMap,
			opentracing.TextMapCarrier(carrier),
		)
		if err != nil {
			p.logger.Warn("failed to inject tracing context", "error", err)
		}
	}

	_, err = p.client.XAdd(ctx, &redis.XAddArgs{
		MaxLen: p.config.MaxBacklog,
		Approx: p.config.UseDelApprox,
		Stream: p.streamID,
		Values: map[string]interface{}{
			"data":     dataStr,
			"trace_id": carrier["uber-trace-id"],
		},
	}).Result()

	if err != nil {
		p.logger.Error("Failed to publish message.", "error", err, "stream", p.streamID, "data", data)
		return err
	}

	p.logger.Debug("Published message", "stream", p.streamID, "data", data)
	return nil
}
