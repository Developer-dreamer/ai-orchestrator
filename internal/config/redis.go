package config

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

func ConnectToRedis(cfg *Config) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var client *redis.Client
	var err error

	client = redis.NewClient(&redis.Options{
		Addr: cfg.RedisUri,
		DB:   0,
	})

	if _, err = client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return client, nil
}
