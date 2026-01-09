package main

import (
	"ai-orchestrator/internal/config/worker"
	"ai-orchestrator/internal/server"
	"log"
	"log/slog"
)

func main() {
	cfg, err := worker.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := cfg.ConfigureLogger(slog.LevelDebug)

	workers := server.SetupWorkers(cfg, logger)
	server.StartWorkers(logger, workers)
}
