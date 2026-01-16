package app

import (
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/config/env"
	"ai-orchestrator/internal/infra/ai/gemini"
	"ai-orchestrator/internal/infra/broker"
	"ai-orchestrator/internal/transport/stream"
	"ai-orchestrator/internal/use_case/prompt"
	"context"
	"errors"
	"fmt"
	"google.golang.org/genai"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func SetupWorkers(cfg *env.WorkerConfig, logger *slog.Logger) ([]*stream.Consumer, func(context.Context) error) {
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
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		logger.Error("Failed to initiate client.", "error", err)
		os.Exit(1)
	}

	streamOptions := &config.Stream{
		MaxBacklog:   1000,
		UseDelApprox: true,
		ReadCount:    1,
		BlockTime:    5 * time.Second,
	}

	producer := broker.NewProducer(logger, redisClient, streamOptions, cfg.RedisPubStreamID)

	aiProvider := gemini.NewClient(logger, client)
	sendPromptUsecase := prompt.NewSendPrompUsecase(logger, aiProvider, producer)
	var workers []*stream.Consumer

	for i := 0; i < cfg.GetNumberOfWorkers(); i++ {
		workerName := fmt.Sprintf("worker-%d", i)

		w := stream.NewConsumer(logger, sendPromptUsecase, redisClient, streamOptions, cfg.RedisSubStreamID, "ai_tasks_group", workerName)
		workers = append(workers, w)
	}

	return workers, closer
}

func StartWorkers(logger *slog.Logger, workers []*stream.Consumer, tracerShutdown func(context.Context) error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received termination signal. Stopping worker...", "signal", sig)
		cancel()
	}()

	var wg sync.WaitGroup

	for _, worker := range workers {
		wg.Add(1)

		go func(w *stream.Consumer) {
			defer wg.Done()

			if err := w.Consume(ctx); err != nil {
				if !errors.Is(err, context.Canceled) {
					logger.Error("WorkerConfig failed with error", "id", w.WorkerID, "error", err)
				}
			}

			logger.Info("WorkerConfig stopped gracefully", "id", w.WorkerID)
		}(worker)
	}

	logger.Info("All workers are running. Waiting for tasks...")
	wg.Wait()

	if err := tracerShutdown(ctx); err != nil {
		logger.Error("error occurred when shutting down tracer", "error", err)
	}
	logger.Info("System shutdown complete.")
}
