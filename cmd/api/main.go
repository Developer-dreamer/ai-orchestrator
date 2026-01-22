package main

import (
	"ai-orchestrator/internal/app"
	"ai-orchestrator/internal/config/api"
	"ai-orchestrator/internal/config/setup"
	"log"
	"log/slog"
)

func main() {
	logger := setup.NewLogger(slog.LevelDebug)

	configPath := setup.LoadCfgFilesDir()
	logger.Info("Loading path", "path", configPath)

	cfg, err := setup.Load[api.Config](configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger.Info("Loading cfg", "redisURI", cfg.Redis.URI)

	server, producer, consumer, tracerShutdown := app.SetupHttpServer(cfg, logger)
	app.GracefulShutdown(server, producer, consumer, logger, tracerShutdown)
}
