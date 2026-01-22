package worker

import (
	"ai-orchestrator/internal/config/shared"
)

type Config struct {
	App   AppConfig          `yaml:"app"`
	Redis shared.RedisConfig `yaml:"redis"`
	OTEL  shared.OtelConfig  `yaml:"otel"`
}

type AppConfig struct {
	ID              string `yaml:"id" env:"APP_ID" envDefault:"api"`
	Port            string `yaml:"port" env:"APP_PORT" envDefault:"8080"`
	Environment     string `yaml:"env" env:"APP_ENV" envDefault:"development"`
	NumberOfWorkers int    `yaml:"number_of_workers" env:"NUMBER_OF_WORKERS" envDefault:"1"`

	Backoff shared.BackoffConfig `yaml:"backoff"`
}
