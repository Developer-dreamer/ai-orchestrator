package server

import (
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/infra/redis"
	"ai-orchestrator/internal/service/prompt"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func SetupWorkers(cfg *config.Config, logger *slog.Logger) []*prompt.Consumer {
	redisClient, err := config.ConnectToRedis(cfg)
	if err != nil {
		logger.Error("Failed to initiate redis. Server shutdown.", "error", err)
		os.Exit(1)
	}

	streamOptions := &redis.StreamConfig{
		MaxBacklog:   1000,
		UseDelApprox: true,
		ReadCount:    1,
		BlockTime:    5 * time.Second,
	}
	rds := redis.NewService(logger, redisClient, streamOptions)

	var workers []*prompt.Consumer

	for i := 0; i < cfg.GetNumberOfWorkers(); i++ {
		workerName := fmt.Sprintf("worker-%d", i)

		w := prompt.NewConsumer(logger, rds, "ai_tasks_group", workerName, cfg)
		workers = append(workers, w)
	}

	return workers
}

func StartWorkers(logger *slog.Logger, workers []*prompt.Consumer) {
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

		go func(w *prompt.Consumer) {
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
