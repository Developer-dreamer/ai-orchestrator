package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/domain"
	"context"
	"time"
)

type TaskProducer interface {
	CreateGroup(ctx context.Context, stream, group string) error
	Publish(ctx context.Context, stream string, data any) error
}

type Producer struct {
	logger   common.Logger
	tasks    TaskProducer
	ttl      time.Duration
	streamID string
}

func NewProducer(l common.Logger, s TaskProducer, cfg *config.Config) *Producer {
	return &Producer{logger: l, tasks: s, ttl: cfg.GetCacheTTL(), streamID: cfg.RedisStreamID}
}

func (s *Producer) PostPrompt(ctx context.Context, prompt domain.Prompt) error {
	// TODO: Save to repository

	// TODO: publish
	err := s.tasks.Publish(ctx, s.streamID, prompt)
	if err != nil {
		return err
	}
	// Later there will be no return value, since error will be handled automatically via the Outbox pattern.
	return nil
}
