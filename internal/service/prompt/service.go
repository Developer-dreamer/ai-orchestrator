package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/domain"
	"context"
)

type Service struct {
	logger common.Logger
}

func NewService(l common.Logger) *Service {
	return &Service{l}
}

func (s *Service) PostPrompt(ctx context.Context, prompt *domain.Prompt) error {
	// TODO: Save to repository

	// TODO: cache

	// TODO: push to queue

	return nil
}
