package main

import (
	"ai-orchestrator/internal/config/api"
	"ai-orchestrator/internal/server"
	"log"
	"log/slog"
)

func main() {
	cfg, err := api.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := cfg.ConfigureLogger(slog.LevelDebug)

	srvr, closer := server.SetupHttpServer(cfg, logger)
	defer func() {
		err := closer.Close()
		if err != nil {
			logger.Error("Failed to close tracer.", "error", err)
		}
	}()
	server.GracefulShutdown(srvr, logger)
}
