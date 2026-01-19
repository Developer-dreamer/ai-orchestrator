package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/domain/model"
	"ai-orchestrator/internal/infra/persistence/repository/outbox"
	"context"
)

type Transactor interface {
	WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error
}

type OutboxRepository interface {
	CreateEvent(ctx context.Context, event outbox.Event) error
}

type SavePromptUsecase struct {
	logger common.Logger
	repo   Repository
	tx     Transactor
	outbox OutboxRepository
}

func NewSavePromptUsecase(l common.Logger, repository Repository, tx Transactor, or OutboxRepository) *SavePromptUsecase {
	return &SavePromptUsecase{
		logger: l,
		repo:   repository,
		tx:     tx,
		outbox: or}
}

func (s *SavePromptUsecase) PostPrompt(ctx context.Context, prompt model.Prompt) error {
	prompt.Status = model.Accepted

	return s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		err := s.repo.InsertPrompt(ctx, prompt)
		if err != nil {
			s.logger.ErrorContext(ctx, "saving prompt failed", "error", err)
			return err
		}

		err = s.outbox.CreateEvent(ctx, outbox.FromPromptDomain(prompt, "PostPrompt"))
		if err != nil {
			s.logger.ErrorContext(ctx, "saving event failed", "error", err)
			return err
		}

		return nil
	})
}
