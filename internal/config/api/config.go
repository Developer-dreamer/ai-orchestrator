package api

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
	AppID           string `env:"APP_ID,required"`
	RedisUri        string `env:"REDIS_URI,required"`
	JaegerUri       string `env:"JAEGER_URI,required"`
	CacheTTLMinutes string `env:"CACHE_TTL_MINUTES" envDefault:"5"`
	RedisStreamID   string `env:"REDIS_STREAM_ID" envDefault:"tasks"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Load(cfg, nil)
	if err != nil {
		return nil, err
	}
	minutes, err := strconv.Atoi(cfg.CacheTTLMinutes)
	if err != nil || minutes <= 0 {
		return nil, fmt.Errorf("invalid value for cache_ttl_minutes (must be positive integer): %s", cfg.CacheTTLMinutes)
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
