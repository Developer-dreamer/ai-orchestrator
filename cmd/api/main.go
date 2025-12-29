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
	logger := cfg.ConfigureLogger(slog.LevelInfo)

	srvr := server.SetupServer(cfg, logger)
	server.GracefulShutdown(srvr, logger)
}
