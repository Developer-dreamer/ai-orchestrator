package app

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/config/env"
	"ai-orchestrator/internal/infra/broker"
	"ai-orchestrator/internal/infra/persistence"
	outbox2 "ai-orchestrator/internal/infra/persistence/repository/outbox"
	promptRepo "ai-orchestrator/internal/infra/persistence/repository/prompt"
	promptHandler "ai-orchestrator/internal/transport/http/handler/prompt"
	"ai-orchestrator/internal/transport/http/helper"
	"ai-orchestrator/internal/transport/middleware"
	savePromptUsecase "ai-orchestrator/internal/use_case/prompt"
	"context"
	"errors"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func SetupHttpServer(cfg *env.APIConfig, logger *slog.Logger) (*http.Server, func(context.Context) error) {
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
	pr := promptRepo.NewRepository(logger, postgresClient)
	outbox := outbox2.NewRepository(logger, postgresClient)
	transactor := persistence.NewTransactor(logger, postgresClient)
	savePromptUsecase := savePromptUsecase.NewSavePromptUsecase(logger, producer, pr, transactor, outbox)
	ph := promptHandler.NewHandler(logger, savePromptUsecase)

	r := registerRoutes(ph, logger)
	logger.Info("Starting server")

	return &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}, closer
}

func registerRoutes(handler *promptHandler.Handler, logger common.Logger) *mux.Router {
	r := mux.NewRouter()

	recoveryManager := middleware.NewRecoveryManager(logger)
	r.Use(recoveryManager.Recovery)
	r.Use(middleware.TracingMiddleware)

	r.HandleFunc("/ask", handler.PostPrompt).Methods(http.MethodPost)
	r.HandleFunc("/health", healthCheck).Methods(http.MethodGet)

	return r
}

func healthCheck(rw http.ResponseWriter, _ *http.Request) {
	helper.WriteJSONResponse(rw, http.StatusOK, nil)
}

func GracefulShutdown(server *http.Server, logger *slog.Logger, tracerShutdown func(context.Context) error) {
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

	if err := tracerShutdown(ctx); err != nil {
		logger.Error("error occurred when shutting down tracer", "error", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("error during server shutdown", "error", err, "addr", server.Addr)
	}
}

func startRelay() {

}
