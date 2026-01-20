package setup

import (
	"ai-orchestrator/internal/config/api"
	"ai-orchestrator/internal/config/worker"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

func Load[T api.Config | worker.Config]() (*T, error) {
	var cfg T

	if err := cleanenv.ReadConfig("config.yml", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("config error: %w", err)
		}
	}

	return &cfg, nil
}
