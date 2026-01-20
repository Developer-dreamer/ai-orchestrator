package main

import (
	"ai-orchestrator/internal/app"
	"ai-orchestrator/internal/config/api"
	"ai-orchestrator/internal/config/setup"
	"log"
	"log/slog"
)

func main() {
	configPath := setup.LoadCfgFilesDir()
	cfg, err := setup.Load[api.Config](configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := setup.NewLogger(slog.LevelDebug)

	server, producer, consumer, tracerShutdown := app.SetupHttpServer(cfg, logger)
	app.GracefulShutdown(server, producer, consumer, logger, tracerShutdown)
}
