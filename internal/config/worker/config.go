package worker

import (
	"fmt"
	"go-simpler.org/env"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	AppID           string `env:"APP_ID,required"`
	RedisUri        string `env:"REDIS_URI,required"`
	JaegerUri       string `env:"JAEGER_URI,required"`
	RedisStreamID   string `env:"REDIS_STREAM_ID" envDefault:"tasks"`
	NumberOfWorkers string `env:"NUMBER_OF_WORKERS" envDefault:"1"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Load(cfg, nil)
	if err != nil {
		return nil, err
	}
	workers, err := strconv.Atoi(cfg.NumberOfWorkers)
	if err != nil || workers < 1 {
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

func (cfg *Config) GetNumberOfWorkers() int {
	// Error is deliberately ignored. All checks pass at the level of loading
	workers, _ := strconv.Atoi(cfg.NumberOfWorkers)
	return workers
}
