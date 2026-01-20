package manager

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/config/shared"
	"ai-orchestrator/internal/infra/persistence/repository/outbox"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"math/rand"
	"time"
)

var ErrNilOutbox = errors.New("outbox is nil")

type Outbox interface {
	GetAllPendingEvents(ctx context.Context, count int) ([]outbox.Event, error)
	ChangeEventStatus(ctx context.Context, eventID uuid.UUID, eventStatus outbox.Status) error
	IncrementRetryCount(ctx context.Context, eventID uuid.UUID, errorMessage string) error
}

var ErrNilProducer = errors.New("producer is nil")

type Producer interface {
	Publish(ctx context.Context, data json.RawMessage) error
}

var ErrNilConfig = errors.New("backoff config is nil")

type Relay struct {
	logger     logger.Logger
	tx         common.TransactionManager
	repo       Outbox
	producer   Producer
	backoffCfg *shared.BackoffConfig
}

func NewRelayService(l logger.Logger, tx common.TransactionManager, repo Outbox, producer Producer, backoffCfg *shared.BackoffConfig) (*Relay, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}
	if tx == nil {
		return nil, common.ErrNilTransactionManager
	}
	if repo == nil {
		return nil, ErrNilOutbox
	}
	if producer == nil {
		return nil, ErrNilProducer
	}
	if backoffCfg == nil {
		return nil, ErrNilConfig
	}

	return &Relay{
		logger:     l,
		tx:         tx,
		repo:       repo,
		producer:   producer,
		backoffCfg: backoffCfg,
	}, nil
}

func (r *Relay) Start(ctx context.Context) error {
	r.logger.Info("Publisher started")

	currentBackoff := r.backoffCfg.Min
	maxEvents := 10

	ticker := time.NewTicker(r.backoffCfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Stopping producer")
			return ctx.Err()
		default:
			events, err := r.repo.GetAllPendingEvents(ctx, maxEvents)
			if err != nil {
				newBackOff, backOffErr := r.backOff(ctx, currentBackoff)
				currentBackoff = newBackOff

				if backOffErr != nil {
					return errors.Join(err, backOffErr)
				}

				continue
			}

			currentBackoff = r.backoffCfg.Min
			if len(events) == 0 {
				continue
			}

			for _, event := range events {
				r.processSingleEvent(ctx, event)
			}
		}
	}
}

func (r *Relay) processSingleEvent(ctx context.Context, event outbox.Event) {
	ctx, span := r.restoreTraceContext(ctx, &event)
	defer span.End()

	r.logger.InfoContext(ctx, "Sending message to stream", "message_id", event.ID)

	err := r.producer.Publish(ctx, event.Payload)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to publish message", "message_id", event.ID, "error", err)

		_ = r.saveProcessingError(ctx, event, err)
		return
	}

	if err := r.repo.ChangeEventStatus(ctx, event.ID, outbox.Processed); err != nil {
		r.logger.ErrorContext(ctx, "Failed to mark event as processed", "error", err)
	}
}

func (r *Relay) saveProcessingError(ctx context.Context, event outbox.Event, err error) error {
	return r.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if event.RetryCount > 5 {
			if dbErr := r.repo.ChangeEventStatus(ctx, event.ID, outbox.Failed); dbErr != nil {
				r.logger.ErrorContext(ctx, "Failed to mark event as processed", "error", err)

				return dbErr
			}
		} else {
			if dbErr := r.repo.IncrementRetryCount(ctx, event.ID, err.Error()); dbErr != nil {
				r.logger.ErrorContext(ctx, "Failed to increment retry count", "error", err)

				return dbErr
			}

		}
		return nil
	})
}

func (r *Relay) restoreTraceContext(ctx context.Context, event *outbox.Event) (context.Context, trace.Span) {
	traceID, _ := trace.TraceIDFromHex(event.TraceID)

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		Remote:     true,
		TraceFlags: trace.FlagsSampled,
	})

	parentCtx := trace.ContextWithRemoteSpanContext(ctx, spanContext)

	return otel.Tracer("outbox-relay").Start(parentCtx, "relay_process",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attribute.String("event_id", event.ID.String())),
	)
}

func (r *Relay) backOff(ctx context.Context, currentBackoff time.Duration) (time.Duration, error) {
	jitter := time.Duration(rand.Int63n(int64(currentBackoff) / 5))
	sleepTime := currentBackoff + jitter

	r.logger.Info("Backoff active", "sleep_time", sleepTime.String())

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-time.After(sleepTime):
		// Time's up - continuing work
	}

	currentBackoff *= time.Duration(r.backoffCfg.Factor)
	if currentBackoff > r.backoffCfg.Max {
		currentBackoff = r.backoffCfg.Max
	}

	return currentBackoff, nil
}
