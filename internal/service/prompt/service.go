package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/domain"
	"context"
	"fmt"
	"time"
)

type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, cacheKey string, data any, ttl time.Duration) error
	Del(ctx context.Context, key string) error
}

type Service struct {
	logger common.Logger
	cache  CacheService
	ttl    time.Duration
}

func NewService(l common.Logger, r CacheService, cfg *config.Config) *Service {
	return &Service{logger: l, cache: r, ttl: cfg.GetCacheTTL()}
}

func (s *Service) PostPrompt(ctx context.Context, prompt domain.Prompt) error {
	// TODO: Save to repository

	// TODO: cache
	err := s.cache.Set(ctx, fmt.Sprintf("prompt:%s", prompt.ID), prompt, s.ttl)
	if err != nil {
		s.logger.Error("Failed to set cache", "error", err)
	}
	return nil
}
