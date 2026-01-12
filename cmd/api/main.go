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
		log.Fatalf("failed to load appConfig: %v", err)
	}
	logger := config.NewLogger(slog.LevelDebug)

	server, closer := app.SetupHttpServer(cfg, logger)
	defer func() {
		err := closer.Close()
		if err != nil {
			logger.Error("Failed to close tracer.", "error", err)
		}
	}()
	app.GracefulShutdown(server, logger)
}
