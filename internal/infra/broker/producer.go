package broker

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/config/env"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
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
		p.logger.ErrorContext(ctx, "Failed to convert data to json.", "error", err, "data", data)
		return err
	}

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	_, err = p.client.XAdd(ctx, &redis.XAddArgs{
		MaxLen: p.config.MaxBacklog,
		Approx: p.config.UseDelApprox,
		Stream: p.streamID,
		Values: map[string]interface{}{
			"data":     dataStr,
			"trace_id": carrier.Get("uber-trace-id"),
		},
	}).Result()

	if err != nil {
		p.logger.ErrorContext(ctx, "Failed to publish message.", "error", err, "stream", p.streamID, "data", data)
		return err
	}

	p.logger.DebugContext(ctx, "Published message", "stream", p.streamID, "data", data)
	return nil
}
