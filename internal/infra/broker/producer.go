package broker

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/config/shared"
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type Producer struct {
	logger logger.Logger
	client *redis.Client
	config *shared.StreamConfig
}

func NewProducer(l logger.Logger, client *redis.Client, streamCfg *shared.StreamConfig) (*Producer, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}
	if client == nil {
		return nil, errors.New("redis client is nil")
	}
	if streamCfg == nil {
		return nil, errors.New("redis stream config is nil")
	}

	return &Producer{
		logger: l,
		client: client,
		config: streamCfg,
	}, nil
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
		Stream: p.config.ID,
		Values: values,
	}).Result()

	if err != nil {
		p.logger.ErrorContext(ctx, "Failed to publish message.", "error", err, "stream", p.config.ID, "data", data)
		return err
	}

	p.logger.DebugContext(ctx, "Published message", "stream", p.config.ID, "data", data)
	return nil
}
