package app

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/config/api"
	"ai-orchestrator/internal/config/connector"
	"ai-orchestrator/internal/config/setup"
	"ai-orchestrator/internal/infra/broker"
	"ai-orchestrator/internal/infra/manager"
	"ai-orchestrator/internal/infra/persistence"
	outboxRepo "ai-orchestrator/internal/infra/persistence/repository/outbox"
	promptRepo "ai-orchestrator/internal/infra/persistence/repository/prompt"
	"ai-orchestrator/internal/infra/telemetry/tracing"
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

func SetupHttpServer(cfg *api.Config, l *slog.Logger) (*http.Server, *manager.Relay, *stream.Consumer, func(context.Context) error) {
	redisClient, err := connector.ConnectToRedis(cfg.Redis.URI)
	if err != nil {
		l.Error("Failed to initiate redis. Server shutdown.", "error", err)
		os.Exit(1)
	}
	closer, err := setup.InitTracer(cfg.App.ID, cfg.OTEL.URI)
	if err != nil {
		l.Error("Failed to initiate tracer.", "error", err)
		os.Exit(1)
	}
	postgresClient, err := connector.ConnectToPostgres(cfg.Postgres)
	if err != nil {
		l.Error("Failed to initiate postgres. Server shutdown.", "error", err)
		os.Exit(1)
	}
	if err = connector.RunMigrations(postgresClient, cfg.App.MigrationsDir); err != nil {
		l.Error("error while running migrations", "error", err)
		os.Exit(1)
	}

	transactor, err := persistence.NewTransactor(l, postgresClient)
	if err != nil {
		l.Error("Failed to initiate transactor.", "error", err)
		os.Exit(1)
	}
	outbox, err := outboxRepo.NewRepository(l, postgresClient)
	if err != nil {
		l.Error("Failed to initiate outbox.", "error", err)
		os.Exit(1)
	}

	producer, err := broker.NewProducer(l, redisClient, &cfg.Redis.PubStream)
	if err != nil {
		l.Error("Failed to initiate producer.", "error", err)
		os.Exit(1)
	}
	relay, err := manager.NewRelayService(l, transactor, outbox, producer, &cfg.App.Backoff)
	if err != nil {
		l.Error("Failed to initiate relay.", "error", err)
		os.Exit(1)
	}

	upgrader := &wslib.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // Restrict in production!
	}

	hub := websocket.NewHub()
	socket, err := websocket.NewManager(l, upgrader, hub)
	if err != nil {
		l.Error("Failed to initiate websocket.", "error", err)
		os.Exit(1)
	}

	pr, err := promptRepo.NewRepository(l, postgresClient)
	if err != nil {
		l.Error("Failed to initiate prompt repository.", "error", err)
		os.Exit(1)
	}

	savePrompt, err := savePromptUsecase.NewSavePromptUsecase(l, pr, transactor, outbox)
	if err != nil {
		l.Error("Failed to initiate save prompt usecase.", "error", err)
		os.Exit(1)
	}

	ph, err := promptHandler.NewHandler(l, savePrompt)
	if err != nil {
		l.Error("Failed to initiate prompt handler.", "error", err)
		os.Exit(1)
	}

	saveResponse, err := savePromptUsecase.NewSaveResponse(l, socket, pr)
	if err != nil {
		l.Error("Failed to initiate save response.", "error", err)
		os.Exit(1)
	}

	tracePropagator := &tracing.PropagationConfig{
		AppID:     cfg.App.ID,
		ProcessID: "save_response",
	}
	backoffManager, err := manager.NewBackoff(l, &cfg.App.Backoff)
	if err != nil {
		l.Error("Failed to initiate backoffManager.", "error", err)
		os.Exit(1)
	}

	consumer, err := stream.NewConsumer(0, l, redisClient, saveResponse, &cfg.Redis.SubStream, &cfg.App.Backoff, tracePropagator, *backoffManager)
	if err != nil {
		l.Error("Failed to initiate consumer.", "error", err)
		os.Exit(1)
	}

	r := registerRoutes(ph, socket, l)

	l.Info("Starting server")

	return &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}, relay, consumer, closer
}

func registerRoutes(handler *promptHandler.Handler, socketManager *websocket.Manager, logger logger.Logger) *mux.Router {
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

func GracefulShutdown(server *http.Server, relay *manager.Relay, consumer *stream.Consumer, logger *slog.Logger, tracerShutdown func(context.Context) error) {
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
		if err := consumer.Consume(appCtx); err != nil {
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
