package main

import (
	"ai-orchestrator/internal/app"
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/config/env"
	"log"
	"log/slog"
)

func main() {
	cfg, err := env.LoadWorkerConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := config.NewLogger(slog.LevelDebug)

	workers, tracerShutdown := app.SetupWorkers(cfg, logger)
	app.StartWorkers(logger, workers, tracerShutdown)
}
