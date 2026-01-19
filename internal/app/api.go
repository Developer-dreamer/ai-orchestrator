package app

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/config/env"
	"ai-orchestrator/internal/infra/broker"
	"ai-orchestrator/internal/infra/manager"
	"ai-orchestrator/internal/infra/persistence"
	outbox2 "ai-orchestrator/internal/infra/persistence/repository/outbox"
	promptRepo "ai-orchestrator/internal/infra/persistence/repository/prompt"
	"ai-orchestrator/internal/infra/websocket"
	promptHandler "ai-orchestrator/internal/transport/http/handler/prompt"
	"ai-orchestrator/internal/transport/http/helper"
	"ai-orchestrator/internal/transport/middleware"
	"ai-orchestrator/internal/transport/stream"
	savePromptUsecase "ai-orchestrator/internal/use_case/prompt"
	"context"
	"errors"
	"github.com/gorilla/mux"
	wslib "github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func SetupHttpServer(cfg *env.APIConfig, logger *slog.Logger) (*http.Server, *manager.Relay, *stream.ResConsumer, func(context.Context) error) {
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

	backoffOptions := &config.Backoff{
		Min:          1 * time.Second,
		Max:          60 * time.Second,
		Factor:       2,
		PollInterval: 50 * time.Millisecond,
	}

	transactor := persistence.NewTransactor(logger, postgresClient)
	outbox := outbox2.NewRepository(logger, postgresClient)

	producer := broker.NewProducer(logger, redisClient, streamOptions, cfg.RedisPubStreamID)
	relay := manager.NewRelayService(logger, transactor, outbox, producer, backoffOptions)

	upgrader := &wslib.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // Restrict in production!
	}

	hub := websocket.NewHub()
	socket := websocket.NewManager(logger, upgrader, hub)
	pr := promptRepo.NewRepository(logger, postgresClient)

	savePrompt := savePromptUsecase.NewSavePromptUsecase(logger, pr, transactor, outbox)
	ph := promptHandler.NewHandler(logger, savePrompt)

	saveResponse := savePromptUsecase.NewSaveResponse(logger, socket, pr)
	consumer := stream.NewResConsumer(logger, redisClient, streamOptions, saveResponse, cfg.RedisResStreamID, "ai_tasks_group", "worker-1")
	r := registerRoutes(ph, socket, logger)

	logger.Info("Starting server")

	return &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}, relay, consumer, closer
}

func registerRoutes(handler *promptHandler.Handler, socketManager *websocket.Manager, logger common.Logger) *mux.Router {
	r := mux.NewRouter()

	recoveryManager := middleware.NewRecoveryManager(logger)
	r.Use(recoveryManager.Recovery)
	r.Use(middleware.TracingMiddleware)

	r.HandleFunc("/ask", handler.PostPrompt).Methods(http.MethodPost)
	r.HandleFunc("/health", healthCheck).Methods(http.MethodGet)

	r.HandleFunc("/ws", socketManager.ServeWS).Methods(http.MethodGet)

	return r
}

func healthCheck(rw http.ResponseWriter, _ *http.Request) {
	helper.WriteJSONResponse(rw, http.StatusOK, nil)
}

func GracefulShutdown(server *http.Server, relay *manager.Relay, consumer *stream.ResConsumer, logger *slog.Logger, tracerShutdown func(context.Context) error) {
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
		logger.Info("Starting relay")
		if err := relay.Start(appCtx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error("error occurred in relay", "error", err)
			}
		}
	}()

	go func() {
		logger.Info("Starting consumer")
		if err := consumer.ConsumeResult(appCtx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error("error occurred in consumer", "error", err)
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
