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

	workers, closer := server.SetupWorkers(cfg, logger)
	defer func() {
		err := closer.Close()
		if err != nil {
			logger.Error("Failed to close tracer.", "error", err)
		}
	}()
	server.StartWorkers(logger, workers)
}
