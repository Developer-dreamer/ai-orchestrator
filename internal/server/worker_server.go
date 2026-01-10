package server

import (
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/config/worker"
	"ai-orchestrator/internal/infra/redis"
	"ai-orchestrator/internal/infra/transport/stream"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func SetupWorkers(cfg *worker.Config, logger *slog.Logger) ([]*stream.Consumer, io.Closer) {
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

	streamOptions := &redis.StreamConfig{
		MaxBacklog:   1000,
		UseDelApprox: true,
		ReadCount:    1,
		BlockTime:    5 * time.Second,
	}
	rds := redis.NewService(logger, redisClient, streamOptions)

	var workers []*stream.Consumer

	for i := 0; i < cfg.GetNumberOfWorkers(); i++ {
		workerName := fmt.Sprintf("worker-%d", i)

		w := stream.NewConsumer(logger, rds, cfg.RedisStreamID, "ai_tasks_group", workerName)
		workers = append(workers, w)
	}

	return workers, closer
}

func StartWorkers(logger *slog.Logger, workers []*stream.Consumer) {
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
					logger.Error("Worker failed with error", "id", w.WorkerID, "error", err)
				}
			}

			logger.Info("Worker stopped gracefully", "id", w.WorkerID)
		}(worker)
	}

	logger.Info("All workers are running. Waiting for tasks...")
	wg.Wait()
	logger.Info("System shutdown complete.")
}
