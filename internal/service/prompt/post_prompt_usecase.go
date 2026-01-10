package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config/api"
	"ai-orchestrator/internal/domain"
	"context"
	"time"
)

type TaskProducer interface {
	Publish(ctx context.Context, data any) error
}

type Producer struct {
	logger common.Logger
	tasks  TaskProducer
	ttl    time.Duration
}

func NewProducer(l common.Logger, s TaskProducer, cfg *api.Config) *Producer {
	return &Producer{logger: l, tasks: s, ttl: cfg.GetCacheTTL()}
}

func (s *Producer) PostPrompt(ctx context.Context, prompt domain.Prompt) error {
	// TODO: Save to repository

	// TODO: publish
	err := s.tasks.Publish(ctx, prompt)
	if err != nil {
		return err
	}
	// Later there will be no return value, since error will be handled automatically via the Outbox pattern.
	return nil
}
