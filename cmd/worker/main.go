package main

import (
	"ai-orchestrator/internal/app"
	"ai-orchestrator/internal/config/setup"
	"ai-orchestrator/internal/config/worker"
	"log"
	"log/slog"
)

func main() {
	logger := setup.NewLogger(slog.LevelDebug)

	configPath := setup.LoadCfgFilesDir()
	logger.Info("Loading path", "path", configPath)

	cfg, err := setup.Load[worker.Config](configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger.Info("Loading cfg", "redisURI", cfg.Redis.URI)

	workers, tracerShutdown := app.SetupWorkers(cfg, logger)
	app.StartWorkers(logger, cfg, workers, tracerShutdown)
}
