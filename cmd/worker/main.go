package main

import (
	"ai-orchestrator/internal/app"
	"ai-orchestrator/internal/config/setup"
	"ai-orchestrator/internal/config/worker"
	"log"
	"log/slog"
)

func main() {
	cfg, err := setup.Load[worker.Config]()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := setup.NewLogger(slog.LevelDebug)

	workers, tracerShutdown := app.SetupWorkers(cfg, logger)
	app.StartWorkers(logger, workers, tracerShutdown)
}
