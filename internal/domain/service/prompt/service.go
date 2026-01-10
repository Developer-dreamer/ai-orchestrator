package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/domain/model"
	"context"
)

type TaskProducer interface {
	Publish(ctx context.Context, data any) error
}

type Repository interface {
}

type Service struct {
	logger common.Logger
	tasks  TaskProducer
	repo   Repository
}

func NewService(l common.Logger, s TaskProducer, repository Repository) *Service {
	return &Service{logger: l, tasks: s, repo: repository}
}

func (s *Service) PostPrompt(ctx context.Context, prompt model.Prompt) error {
	// TODO: Save to repository

	// TODO: publish
	err := s.tasks.Publish(ctx, prompt)
	if err != nil {
		return err
	}
	// Later there will be no return value, since error will be handled automatically via the Outbox pattern.
	return nil
}
