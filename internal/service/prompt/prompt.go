package prompt

import (
	"ai-orchestrator/internal/domain/prompt"
	"context"
	"github.com/google/uuid"
)

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
}

type Service struct {
	logger Logger
}

func NewService(l Logger) *Service {
	return &Service{l}
}

func (s *Service) PostPrompt(ctx context.Context, prompt *prompt.Prompt) error {
	pID, err := uuid.NewUUID()
	if err != nil {
		// TODO : add logger
		return err
	}

	prompt.ID = pID

	// TODO: Save to repository

	// TODO: cache

	// TODO: push to queue

	return nil
}
