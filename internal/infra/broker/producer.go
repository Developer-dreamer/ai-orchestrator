package broker

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config"
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

func NewProducer(l common.Logger, client *redis.Client, streamCfg *config.Stream, redisStreamID string) *Producer {
	return &Producer{
		logger:   l,
		client:   client,
		config:   streamCfg,
		streamID: redisStreamID,
	}
}

func (p *Producer) Publish(ctx context.Context, data json.RawMessage) error {
	headers := make(map[string]string)

	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))

	values := map[string]interface{}{
		"data": string(data),
	}

	for k, v := range headers {
		values[k] = v
	}

	_, err := p.client.XAdd(ctx, &redis.XAddArgs{
		MaxLen: p.config.MaxBacklog,
		Approx: p.config.UseDelApprox,
		Stream: p.streamID,
		Values: values,
	}).Result()

	if err != nil {
		p.logger.ErrorContext(ctx, "Failed to publish message.", "error", err, "stream", p.streamID, "data", data)
		return err
	}

	p.logger.DebugContext(ctx, "Published message", "stream", p.streamID, "data", data)
	return nil
}
