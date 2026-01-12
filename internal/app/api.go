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

func SetupHttpServer(cfg *env.APIConfig, logger *slog.Logger) (*http.Server, *broker.Producer, func(context.Context) error) {
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
	transactor := persistence.NewTransactor(logger, postgresClient)
	outbox := outbox2.NewRepository(logger, postgresClient)

	producer := broker.NewProducer(logger, redisClient, streamOptions, cfg, outbox, transactor)

	pr := promptRepo.NewRepository(logger, postgresClient)

	savePrompt := savePromptUsecase.NewSavePromptUsecase(logger, pr, transactor, outbox)
	ph := promptHandler.NewHandler(logger, savePrompt)

	r := registerRoutes(ph, logger)
	logger.Info("Starting server")

	return &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}, producer, closer
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

func GracefulShutdown(server *http.Server, producer *broker.Producer, logger *slog.Logger, tracerShutdown func(context.Context) error) {
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("starting HTTP server", "addr", server.Addr, "idle_timeout", server.IdleTimeout, "read_timeout", server.ReadTimeout, "write_timeout", server.WriteTimeout)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error occurred when starting server", "error", err)
			appCancel()
		}
	}()

	go func() {
		logger.Info("Starting producer")
		if err := producer.Start(appCtx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error("error occurred in producer", "error", err)
			}
		}
	}()

	select {
	case sig := <-sigChan:
		logger.Info("received terminate signal", "sig", sig)
	case <-appCtx.Done():
		logger.Info("application context canceled (internal error)")
	}

	appCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("error during server shutdown", "error", err, "addr", server.Addr)
	}

	if err := tracerShutdown(ctx); err != nil {
		logger.Error("error occurred when shutting down tracer", "error", err)
	}

	logger.Info("Graceful shutdown completed")
}
