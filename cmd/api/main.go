package main

import (
	"ai-orchestrator/internal/app"
	config "ai-orchestrator/internal/config/app"
	"log"
	"log/slog"
)

func main() {
	cfg, err := config.LoadAPIConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := cfg.ConfigureLogger(slog.LevelDebug)

	server, closer := app.SetupHttpServer(cfg, logger)
	defer func() {
		err := closer.Close()
		if err != nil {
			logger.Error("Failed to close tracer.", "error", err)
		}
	}()
	app.GracefulShutdown(server, logger)
}
