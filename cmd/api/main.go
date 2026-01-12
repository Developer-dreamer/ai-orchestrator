package main

import (
	"ai-orchestrator/internal/app"
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/config/env"
	"log"
	"log/slog"
)

func main() {
	cfg, err := env.LoadAPIConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := config.NewLogger(slog.LevelDebug)

	server, tracerShutdown := app.SetupHttpServer(cfg, logger)
	app.GracefulShutdown(server, logger, tracerShutdown)
}
