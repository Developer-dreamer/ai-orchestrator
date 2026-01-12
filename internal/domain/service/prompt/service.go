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
	InsertPrompt(ctx context.Context, prompt model.Prompt) error
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
	prompt.Status = model.Accepted

	err := s.repo.InsertPrompt(ctx, prompt)
	if err != nil {
		return err
	}
	err = s.tasks.Publish(ctx, prompt)
	if err != nil {
		return err
	}
	return nil
}
