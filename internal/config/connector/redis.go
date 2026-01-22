package connector

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
)

func ConnectToRedis(env string, redisUri string) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var client *redis.Client
	var err error

	switch env {
	case "production":
		client, err = createSecureRedisClient(redisUri)
	default:
		client = redis.NewClient(&redis.Options{
			Addr: redisUri,
			DB:   0,
		})
	}

	if _, err = client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return client, nil
}

func createSecureRedisClient(redisURI string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURI)
	if err != nil {
		return nil, err
	}

	caCertPEM, err := os.ReadFile("/certs/server-ca.pem")
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCertPEM) {
		return nil, fmt.Errorf("append CA cert failed")
	}

	opt.TLSConfig = &tls.Config{
		RootCAs:            pool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	return redis.NewClient(opt), nil
}
