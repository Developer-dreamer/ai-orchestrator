package main

import (
	"ai-orchestrator/internal/app"
	config "ai-orchestrator/internal/config/app"
	"log"
	"log/slog"
)

func main() {
	cfg, err := config.LoadWorkerConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := cfg.ConfigureLogger(slog.LevelDebug)

	workers, closer := app.SetupWorkers(cfg, logger)
	defer func() {
		err := closer.Close()
		if err != nil {
			logger.Error("Failed to close tracer.", "error", err)
		}
	}()
	app.StartWorkers(logger, workers)
}
