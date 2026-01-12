package env

import (
	"fmt"
	"go-simpler.org/env"
	"strconv"
)

type WorkerConfig struct {
	AppID           string `env:"APP_ID,required"`
	RedisUri        string `env:"REDIS_URI,required"`
	JaegerUri       string `env:"JAEGER_URI,required"`
	RedisStreamID   string `env:"REDIS_STREAM_ID" envDefault:"tasks"`
	NumberOfWorkers string `env:"NUMBER_OF_WORKERS" envDefault:"1"`
}

func LoadWorkerConfig() (*WorkerConfig, error) {
	cfg := &WorkerConfig{}
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

func (cfg *WorkerConfig) GetNumberOfWorkers() int {
	// Error is deliberately ignored. All checks pass at the level of loading
	workers, _ := strconv.Atoi(cfg.NumberOfWorkers)
	return workers
}
