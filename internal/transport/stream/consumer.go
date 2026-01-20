package stream

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/config/shared"
	"ai-orchestrator/internal/infra/manager"
	"ai-orchestrator/internal/infra/telemetry/tracing"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"strings"
	"time"
)

type UseCase interface {
	Use(ctx context.Context, entity string) error
}

type Consumer struct {
	WorkerID string

	logger logger.Logger
	client *redis.Client

	usecase UseCase

	streamCfg         *shared.StreamConfig
	backoffCfg        *shared.BackoffConfig
	contextPropagator *tracing.PropagationConfig

	backoff manager.Backoff
}

type ConsumerResult struct {
	Headers   map[string]string
	MessageID string
	Entity    string
}

func NewConsumer(workerID int, l logger.Logger, client *redis.Client, usecase UseCase, streamCfg *shared.StreamConfig, backoffCfg *shared.BackoffConfig, propagator *tracing.PropagationConfig, backoff manager.Backoff) (*Consumer, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}
	if client == nil {
		return nil, errors.New("redis client is nil")
	}
	if streamCfg == nil {
		return nil, errors.New("streamCfg is nil")
	}
	if usecase == nil {
		return nil, errors.New("usecase is nil")
	}
	if propagator == nil {
		return nil, errors.New("propagator is nil")
	}
	var workerFullID string
	if workerID == 0 {
		workerFullID = streamCfg.Group.ConsumerPrimarilyID
	} else {
		workerFullID = fmt.Sprintf("%s-%d", streamCfg.Group.ConsumerPrimarilyID, workerID)
	}

	return &Consumer{
		logger:            l,
		usecase:           usecase,
		client:            client,
		streamCfg:         streamCfg,
		backoffCfg:        backoffCfg,
		WorkerID:          workerFullID,
		contextPropagator: propagator,
		backoff:           backoff,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context) error {
	c.logger.Info("Worker started", "id", c.WorkerID)

	err := c.createGroup(ctx, c.streamCfg.ID, c.streamCfg.Group.ID)
	if err != nil {
		c.logger.Error("Failed to create group", "id", c.WorkerID, "err", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping consumer", "worker_id", c.WorkerID)
			return ctx.Err()
		default:
			res, err := manager.WithBackoff[ConsumerResult](
				ctx,
				&c.backoff,
				func(ctx context.Context) (ConsumerResult, error) {
					return c.consume(ctx)
				},
				func(callErr error) bool {
					return true
				})

			if err != nil {
				return err
			}

			if res.Entity == "" {
				continue
			}

			parentCtx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(res.Headers))

			tracer := otel.Tracer(c.contextPropagator.AppID)
			ctx, span := tracer.Start(parentCtx, c.contextPropagator.ProcessID,
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(attribute.String("redis.message_id", res.MessageID)),
			)
			ctx = logger.WithMessageID(ctx, res.MessageID)
			c.logger.InfoContext(ctx, "Received message from stream")

			err = c.usecase.Use(ctx, res.Entity)
			span.End()
			if err == nil {
				ackCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				ackErr := c.ack(ackCtx, c.streamCfg.ID, c.streamCfg.Group.ID, res.MessageID)
				cancel()

				if ackErr != nil {
					c.logger.Error("Failed to ack message", "error", ackErr)
				}
				continue
			}
			c.logger.ErrorContext(ctx, "Failed to process message", "error", err)
		}
	}
}

func (c *Consumer) createGroup(ctx context.Context, stream, group string) error {

	const EarliestMessage = "0" // Redis specific alias: start from the beginning of the stream

	_, err := c.client.XGroupCreateMkStream(ctx, stream, group, EarliestMessage).Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		c.logger.Warn("Failed to create consumer group.", "error", err, "stream", stream, "group", group)
		return err
	}
	c.logger.Info("Created consumer group", "stream", stream, "group", group)

	return nil
}

// consume Consumes a message from the specified stream. Returns Headers, MessageID, Data, Error
func (c *Consumer) consume(ctx context.Context) (ConsumerResult, error) {

	const undeliveredMessages = ">" // Redis specific alias: starts from the unconsumed message

	res, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Streams:  []string{c.streamCfg.ID, undeliveredMessages},
		Group:    c.streamCfg.Group.ID,
		Consumer: c.WorkerID,
		Count:    c.streamCfg.ReadCount,
		Block:    c.streamCfg.BlockTime,
	}).Result()

	var result ConsumerResult
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return result, nil
		}
		if errors.Is(err, context.Canceled) {
			return result, nil
		}

		c.logger.Error("Failed to consume message.", "error", err, "stream", c.streamCfg.ID, "group", c.streamCfg.Group.ID, "consumer", c.WorkerID)
		return result, err
	}

	if len(res) == 0 || len(res[0].Messages) == 0 {
		return result, nil
	}

	message := res[0].Messages[0]

	headers := make(map[string]string)

	var data string
	for k, v := range message.Values {
		if strVal, ok := v.(string); ok {
			if k == "data" {
				data = strVal
			} else {
				headers[k] = strVal
			}
		}
	}

	c.logger.Debug("Received message", "stream", c.streamCfg.ID, "group", c.streamCfg.Group.ID, "consumer", c.WorkerID, "data", data)
	result.Headers = headers
	result.MessageID = message.ID
	result.Entity = data

	return result, nil
}

func (c *Consumer) ack(ctx context.Context, stream, group, messageId string) error {
	_, err := c.client.XAck(ctx, stream, group, messageId).Result()
	if err != nil {
		c.logger.Error("Acknowledgment failed", "stream", stream, "group", group, "messageId", messageId, "error", err)
		return err
	}
	return nil
}

// TODO add unified backoff process
