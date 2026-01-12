package app

import (
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/config/env"
	promptService "ai-orchestrator/internal/domain/service/prompt"
	"ai-orchestrator/internal/infra/broker"
	"ai-orchestrator/internal/infra/persistence/repository/prompt"
	promptHandler "ai-orchestrator/internal/transport/http/handler/prompt"
	"ai-orchestrator/internal/transport/http/helper"
	"ai-orchestrator/internal/transport/middleware"
	"context"
	"errors"
	"github.com/gorilla/mux"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func SetupHttpServer(cfg *env.APIConfig, logger *slog.Logger) (*http.Server, io.Closer) {
	redisClient, err := config.ConnectToRedis(cfg.RedisUri)
	if err != nil {
		logger.Error("Failed to initiate redis. Server shutdown.", "error", err)
		os.Exit(1)
	}
	closer, err := config.InitTracer(cfg.AppID, cfg.JaegerUri)
	if err != nil {
		logger.Error("Failed to initiate tracer.", "error", err)
		os.Exit(1)
	}
	postgresClient, err := config.ConnectToPostgres(cfg.PostgresUri)
	if err != nil {
		logger.Error("Failed to initiate postgres. Server shutdown.", "error", err)
		os.Exit(1)
	}
	if err = config.RunMigrations(postgresClient, cfg.MigrationsDir); err != nil {
		logger.Error("error while running migrations", "error", err)
		os.Exit(1)
	}

	streamOptions := &config.Stream{
		MaxBacklog:   1000,
		UseDelApprox: true,
		ReadCount:    1,
		BlockTime:    5 * time.Second,
	}
	producer := broker.NewProducer(logger, redisClient, streamOptions, cfg)
	pr := prompt.NewRepository(logger, postgresClient)
	ps := promptService.NewService(logger, producer, pr)
	ph := promptHandler.NewHandler(logger, ps)

	r := registerRoutes(ph)
	logger.Info("Starting server")

	return &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}, closer
}

func registerRoutes(handler *promptHandler.Handler) *mux.Router {
	r := mux.NewRouter()

	r.Use(middleware.TracingMiddleware)

	r.HandleFunc("/ask", handler.PostPrompt).Methods(http.MethodPost)
	r.HandleFunc("/health", healthCheck).Methods(http.MethodGet)

	return r
}

func healthCheck(rw http.ResponseWriter, _ *http.Request) {
	helper.WriteJSONResponse(rw, http.StatusOK, nil)
}

func GracefulShutdown(server *http.Server, logger *slog.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("starting HTTP server", "addr", server.Addr, "idle_timeout", server.IdleTimeout, "read_timeout", server.ReadTimeout, "write_timeout", server.WriteTimeout)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error occurred when starting server", "error", err)
		}
	}()

	sig := <-sigChan
	logger.Info("received terminate, graceful shutdown", "sig", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("error during server shutdown", "error", err, "addr", server.Addr)
	}
}
