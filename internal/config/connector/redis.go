package connector

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

func ConnectToRedis(redisUri string) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr: redisUri,
		DB:   0,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return client, nil
}
