package prompt

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/domain/model"
	"ai-orchestrator/internal/infra/persistence/repository/outbox"
	"context"
	"errors"
)

var ErrNilTransactor = errors.New("transactor is nil")

type Transactor interface {
	WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error
}

var ErrNilOutbox = errors.New("outbox is nil")

type OutboxRepository interface {
	CreateEvent(ctx context.Context, event outbox.Event) error
}

type SavePromptUsecase struct {
	logger logger.Logger
	repo   Repository
	tx     Transactor
	outbox OutboxRepository
}

func NewSavePromptUsecase(l logger.Logger, repository Repository, tx Transactor, or OutboxRepository) (*SavePromptUsecase, error) {
	if l == nil {
		return nil, errors.New("logger is nil")
	}
	if repository == nil {
		return nil, ErrNilRepository
	}
	if tx == nil {
		return nil, ErrNilTransactor
	}
	if or == nil {
		return nil, ErrNilOutbox
	}

	return &SavePromptUsecase{
		logger: l,
		repo:   repository,
		tx:     tx,
		outbox: or,
	}, nil
}

func (s *SavePromptUsecase) PostPrompt(ctx context.Context, prompt model.Prompt) error {
	prompt.Status = model.Accepted

	payload := TaskPayload{
		ID:      prompt.ID,
		UserID:  prompt.UserID,
		ModelID: prompt.ModelID,
		Text:    prompt.Text,
	}

	return s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		err := s.repo.InsertPrompt(ctx, prompt)
		if err != nil {
			s.logger.ErrorContext(ctx, "saving prompt failed", "error", err)
			return err
		}

		err = s.outbox.CreateEvent(ctx, payload.ToEvent("PostPrompt"))
		if err != nil {
			s.logger.ErrorContext(ctx, "saving event failed", "error", err)
			return err
		}

		return nil
	})
}
