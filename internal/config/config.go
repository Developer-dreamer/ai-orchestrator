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
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Load(cfg, nil)
	if err != nil {
		return nil, err
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
	minutes, err := strconv.Atoi(cfg.CacheTTLMinutes)
	if err != nil {
		fmt.Errorf("consider using numeric value for CACHE_TTL_MINUTES: %s", cfg.CacheTTLMinutes)
	}
	if minutes <= 0 {
		minutes = 5
	}
	return time.Duration(minutes) * time.Minute
}
