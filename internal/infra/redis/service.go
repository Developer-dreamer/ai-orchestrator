package redis

import (
	"ai-orchestrator/internal/common"
	"errors"
	"github.com/redis/go-redis/v9"
)

var (
	ErrCacheMiss      = errors.New("cache miss")
	ErrInvalidKey     = errors.New("invalid cache key")
	ErrMarshalFailed  = errors.New("failed to marshal data")
	ErrCacheSetFailed = errors.New("failed to set cache")
)

type Service struct {
	logger common.Logger
	client *redis.Client
	config *StreamConfig
}

func NewService(logger common.Logger, client *redis.Client, config *StreamConfig) *Service {
	return &Service{
		logger: logger,
		client: client,
		config: config,
	}
}
