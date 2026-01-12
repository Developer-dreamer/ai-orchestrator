package broker

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config"
	"ai-orchestrator/internal/config/env"
	"ai-orchestrator/internal/infra/persistence/repository/outbox"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"math/rand"
	"time"
)

const (
	minBackoff    = 1 * time.Second
	maxBackoff    = 60 * time.Second
	backoffFactor = 2
	pollInterval  = 50 * time.Millisecond
)

type Outbox interface {
	GetAllPendingEvents(ctx context.Context, count int) ([]outbox.Event, error)
	ChangeEventStatus(ctx context.Context, eventID uuid.UUID, eventStatus outbox.Status) error
	IncrementRetryCount(ctx context.Context, eventID uuid.UUID, errorMessage string) error
}

type Producer struct {
	logger   common.Logger
	client   *redis.Client
	config   *config.Stream
	streamID string
	repo     Outbox
	tx       common.TransactionManager
}

func NewProducer(l common.Logger, client *redis.Client, streamCfg *config.Stream, cfg *env.APIConfig, outbox Outbox, tx common.TransactionManager) *Producer {
	return &Producer{
		logger:   l,
		client:   client,
		config:   streamCfg,
		streamID: cfg.RedisStreamID,
		repo:     outbox,
		tx:       tx,
	}
}

func (p *Producer) Start(ctx context.Context) error {
	p.logger.Info("Publisher started")

	currentBackoff := minBackoff
	maxEvents := 10

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Stopping producer")
			return ctx.Err()
		default:
			events, err := p.repo.GetAllPendingEvents(ctx, maxEvents)
			if err != nil {
				newBackOff, backOffErr := p.backOff(ctx, currentBackoff)
				currentBackoff = newBackOff

				if backOffErr != nil {
					return errors.Join(err, backOffErr)
				}

				continue
			}

			currentBackoff = minBackoff
			if len(events) == 0 {
				continue
			}

			for _, event := range events {
				p.processSingleEvent(ctx, event)
			}
		}
	}
}

func (p *Producer) processSingleEvent(ctx context.Context, event outbox.Event) {
	ctx, span := p.restoreTraceContext(ctx, &event)
	defer span.End()

	p.logger.InfoContext(ctx, "Sending message to stream", "message_id", event.ID)

	err := p.publish(ctx, event.Payload)
	if err != nil {
		p.logger.ErrorContext(ctx, "Failed to publish message", "message_id", event.ID, "error", err)

		_ = p.tx.WithinTransaction(ctx, func(ctx context.Context) error {
			if event.RetryCount > 5 {
				if dbErr := p.repo.ChangeEventStatus(ctx, event.ID, outbox.Failed); dbErr != nil {
					p.logger.ErrorContext(ctx, "Failed to mark event as processed", "error", err)

					return dbErr
				}
			} else {
				if dbErr := p.repo.IncrementRetryCount(ctx, event.ID, err.Error()); dbErr != nil {
					p.logger.ErrorContext(ctx, "Failed to increment retry count", "error", err)

					return dbErr
				}

			}
			return nil
		})

		return
	}

	if err := p.repo.ChangeEventStatus(ctx, event.ID, outbox.Processed); err != nil {
		p.logger.ErrorContext(ctx, "Failed to mark event as processed", "error", err)
	}
}

func (p *Producer) restoreTraceContext(ctx context.Context, event *outbox.Event) (context.Context, trace.Span) {
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

func (p *Producer) backOff(ctx context.Context, currentBackoff time.Duration) (time.Duration, error) {
	jitter := time.Duration(rand.Int63n(int64(currentBackoff) / 5))
	sleepTime := currentBackoff + jitter

	p.logger.Info("Backoff active", "sleep_time", sleepTime.String())

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-time.After(sleepTime):
		// Time's up - continuing work
	}

	currentBackoff *= time.Duration(backoffFactor)
	if currentBackoff > maxBackoff {
		currentBackoff = maxBackoff
	}

	return currentBackoff, nil
}

func (p *Producer) publish(ctx context.Context, data []byte) error {
	headers := make(map[string]string)

	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))

	values := map[string]interface{}{
		"data": data,
	}

	for k, v := range headers {
		values[k] = v
	}

	_, err := p.client.XAdd(ctx, &redis.XAddArgs{
		MaxLen: p.config.MaxBacklog,
		Approx: p.config.UseDelApprox,
		Stream: p.streamID,
		Values: values,
	}).Result()

	if err != nil {
		p.logger.ErrorContext(ctx, "Failed to publish message.", "error", err, "stream", p.streamID, "data", data)
		return err
	}

	p.logger.DebugContext(ctx, "Published message", "stream", p.streamID, "data", data)
	return nil
}
