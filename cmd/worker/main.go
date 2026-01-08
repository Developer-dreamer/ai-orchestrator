package main

import (
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/server"
	"log"
	"log/slog"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := cfg.ConfigureLogger(slog.LevelDebug)

	workers := server.SetupWorkers(cfg, logger)
	server.StartWorkers(logger, workers)
}
