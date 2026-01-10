package stream

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config/api"
	"context"
)

type TaskProducer interface {
	Publish(ctx context.Context, stream string, data any) error
}

type Producer struct {
	logger   common.Logger
	tasks    TaskProducer
	streamID string
}

func NewProducer(l common.Logger, s TaskProducer, cfg *api.Config) *Producer {
	return &Producer{logger: l, tasks: s, streamID: cfg.RedisStreamID}
}

func (p *Producer) Publish(ctx context.Context, data any) error {
	return p.tasks.Publish(ctx, p.streamID, data)
}
