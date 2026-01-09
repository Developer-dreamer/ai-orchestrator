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

	srvr := server.SetupHttpServer(cfg, logger)
	server.GracefulShutdown(srvr, logger)
}
