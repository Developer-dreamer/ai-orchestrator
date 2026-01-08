package config

import (
	"fmt"
	"go-simpler.org/env"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppPort         string `env:"PORT,required"`
	RedisUri        string `env:"REDIS_URI,required"`
	CacheTTLMinutes string `env:"CACHE_TTL_MINUTES" envDefault:"5"`
	RedisStreamID   string `env:"REDIS_STREAM_ID" envDefault:"tasks"`
	NumberOfWorkers string `env:"NUMBER_OF_WORKERS" envDefault:"1"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Load(cfg, nil)
	if err != nil {
		return nil, err
	}
	if _, err = strconv.Atoi(cfg.CacheTTLMinutes); err != nil {
		return nil, fmt.Errorf("invalid value for cache_ttl_minutes: %s", cfg.CacheTTLMinutes)
	}
	if _, err = strconv.Atoi(cfg.NumberOfWorkers); err != nil {
		return nil, fmt.Errorf("invalid value for number_of_workers: %s", cfg.NumberOfWorkers)
	}
	return cfg, nil
}

func (cfg *Config) ConfigureLogger(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}

func (cfg *Config) GetCacheTTL() time.Duration {
	// Error is deliberately ignored. All checks pass at the level of loading
	minutes, _ := strconv.Atoi(cfg.CacheTTLMinutes)
	return time.Duration(minutes) * time.Minute
}

func (cfg *Config) GetNumberOfWorkers() int {
	// Error is deliberately ignored. All checks pass at the level of loading
	workers, _ := strconv.Atoi(cfg.NumberOfWorkers)
	return workers
}
