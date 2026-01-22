package app

import (
	"ai-orchestrator/internal/config/connector"
	"ai-orchestrator/internal/config/setup"
	"ai-orchestrator/internal/config/worker"
	"ai-orchestrator/internal/infra/ai/gemini"
	"ai-orchestrator/internal/infra/broker"
	"ai-orchestrator/internal/infra/manager"
	"ai-orchestrator/internal/infra/telemetry/tracing"
	prompt2 "ai-orchestrator/internal/transport/stream"
	"ai-orchestrator/internal/use_case/prompt"
	"context"
	"errors"
	"google.golang.org/genai"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func SetupWorkers(cfg *worker.Config, l *slog.Logger) ([]*prompt2.Consumer, func(context.Context) error) {
	redisClient, err := connector.ConnectToRedis(cfg.App.Environment, cfg.Redis.URI)
	if err != nil {
		l.Error("Failed to initiate redis. Server shutdown.", "error", err)
		os.Exit(1)
	}
	closer, err := setup.InitTracer(cfg.App.ID, cfg.OTEL.URI)
	if err != nil {
		l.Error("Failed to initiate tracer.", "error", err)
		os.Exit(1)
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		l.Error("Failed to initiate client.", "error", err)
		os.Exit(1)
	}

	producer, err := broker.NewProducer(l, redisClient, &cfg.Redis.PubStream)
	if err != nil {
		l.Error("Failed to initiate producer.", "error", err)
		os.Exit(1)
	}

	backoffManager, err := manager.NewBackoff(l, &cfg.App.Backoff)
	if err != nil {
		l.Error("Failed to initiate backoffManager.", "error", err)
		os.Exit(1)
	}

	aiProvider, err := gemini.NewClient(l, client, *backoffManager)
	if err != nil {
		l.Error("Failed to initiate ai provider.", "error", err)
		os.Exit(1)
	}
	sendPromptUsecase, err := prompt.NewSendPromptUsecase(l, aiProvider, producer)
	if err != nil {
		l.Error("Failed to initiate sendPrompUsecase.", "error", err)
		os.Exit(1)
	}

	var workers []*prompt2.Consumer

	tracePropagator := &tracing.PropagationConfig{
		AppID:     cfg.App.ID,
		ProcessID: "send_to_ai",
	}

	for i := 1; i <= cfg.App.NumberOfWorkers; i++ {
		consumer, err := prompt2.NewConsumer(i, l, redisClient, sendPromptUsecase, &cfg.Redis.SubStream, &cfg.App.Backoff, tracePropagator, *backoffManager)
		if err != nil {
			l.Error("Failed to initiate consumer.", "error", err)
			os.Exit(1)
		}
		workers = append(workers, consumer)
	}

	return workers, closer
}

func StartWorkers(logger *slog.Logger, cfg *worker.Config, workers []*prompt2.Consumer, tracerShutdown func(context.Context) error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received termination signal. Stopping workers...", "signal", sig)
		cancel()
	}()

	go addHealthCheck(logger, &cfg.App)

	var wg sync.WaitGroup
	logger.Info("Starting workers...", "numberOfWorkers", len(workers))

	for _, w := range workers {
		wg.Add(1)

		go func(w *prompt2.Consumer) {
			defer wg.Done()

			if err := w.Consume(ctx); err != nil {
				if !errors.Is(err, context.Canceled) {
					logger.Error("Worker failed", "id", w.WorkerID, "error", err)
				}
			}
			logger.Info("Worker stopped gracefully", "id", w.WorkerID)
		}(w)
	}

	logger.Info("All workers are running. Waiting for tasks...")

	wg.Wait()

	logger.Info("Shutting down tracer...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := tracerShutdown(shutdownCtx); err != nil {
		logger.Error("Error occurred when shutting down tracer", "error", err)
	}

	logger.Info("System shutdown complete.")
}

func addHealthCheck(logger *slog.Logger, cfg *worker.AppConfig) {
	server := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Worker is running"))
		}),
	}

	logger.Info("Starting health check server", "port", cfg.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Failed to start health check server", "error", err)
		os.Exit(1)
	}
}
