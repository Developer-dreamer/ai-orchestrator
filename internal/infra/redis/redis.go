package redis

import (
	"ai-orchestrator/internal/common"
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
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
}

func NewService(logger common.Logger, client *redis.Client) *Service {
	return &Service{
		logger: logger,
		client: client,
	}
}

func (s *Service) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		s.logger.Warn("empty key provided", "key", key, "service", "cacheService")
		return "", ErrInvalidKey
	}
	result, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			s.logger.Info("cache miss", "operation", "Get", "key", key, "service", "cacheService")
			return "", ErrCacheMiss
		}
		s.logger.Warn("failed to get cache", "key", key, "error", err, "service", "cacheService")
		return "", err
	}

	s.logger.Info("cache hit", "key", key, "service", "cacheService")
	return result, nil
}

func (s *Service) Set(ctx context.Context, cacheKey string, data any, ttl time.Duration) error {
	if cacheKey == "" {
		s.logger.Warn("empty cache key provided", "key", cacheKey, "service", "cacheService")
		return ErrInvalidKey
	}

	var value []byte
	switch v := data.(type) {
	case string:
		value = []byte(v)
	case []byte:
		value = v
	default:
		var err error
		value, err = json.Marshal(v)
		if err != nil {
			s.logger.Error("failed to marshal cache value", "key", cacheKey, "error", err)
			return ErrMarshalFailed
		}
	}

	if err := s.client.Set(ctx, cacheKey, value, ttl).Err(); err != nil {
		s.logger.Error("failed to set cache", "key", cacheKey, "error", err)
		return err
	}

	s.logger.Info("cache set", "key", cacheKey, "data", data, "service", "cacheService")

	return nil
}

func (s *Service) Del(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}
