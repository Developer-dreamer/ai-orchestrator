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

	workers, closer := app.SetupWorkers(cfg, logger)
	defer func() {
		err := closer.Close()
		if err != nil {
			logger.Error("Failed to close tracer.", "error", err)
		}
	}()
	app.StartWorkers(logger, workers)
}
