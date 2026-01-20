package manager

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/config/shared"
	"context"
	"errors"
	"math/rand"
	"time"
)

type Backoff struct {
	logger         logger.Logger
	cfg            *shared.BackoffConfig
	currentBackoff time.Duration
}

func NewBackoff(l logger.Logger, cfg *shared.BackoffConfig) (*Backoff, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}

	return &Backoff{
		logger:         l,
		cfg:            cfg,
		currentBackoff: cfg.Min,
	}, nil
}

func WithBackoff[T any](
	ctx context.Context,
	backoff *Backoff,
	operation func(ctx context.Context) (T, error),
	isRetryable func(error) bool,
) (T, error) {

	var zero T

	for i := 0; i < backoff.cfg.MaxRetries; i++ {
		result, err := operation(ctx)
		if err == nil {
			backoff.currentBackoff = backoff.cfg.Min

			return result, err
		}

		if ctx.Err() != nil {
			return zero, ctx.Err()
		}

		if !isRetryable(err) {
			backoff.logger.ErrorContext(ctx, "Non-retryable error, stopping backoff", "err", err)
			return zero, err
		}

		jitter := time.Duration(rand.Int63n(int64(backoff.currentBackoff) / 5))
		sleepTime := backoff.currentBackoff + jitter

		backoff.logger.InfoContext(ctx, "Backoff active", "sleep_time", sleepTime.String(), "err", err)

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(sleepTime):

		}

		backoff.currentBackoff *= time.Duration(backoff.cfg.Factor)
		if backoff.currentBackoff > backoff.cfg.Max {
			backoff.currentBackoff = backoff.cfg.Max
		}
	}

	return zero, errors.New("reached max retries")
}
