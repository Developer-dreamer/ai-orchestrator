package stream

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/config/shared"
	"ai-orchestrator/internal/infra/telemetry/tracing"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"math/rand"
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
}

func NewConsumer(workerID int, l logger.Logger, client *redis.Client, usecase UseCase, streamCfg *shared.StreamConfig, backoffCfg *shared.BackoffConfig, propagator *tracing.PropagationConfig) (*Consumer, error) {
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
	}, nil
}

func (c *Consumer) Consume(ctx context.Context) error {
	c.logger.Info("Worker started", "id", c.WorkerID)

	currentBackoff := c.backoffCfg.Min

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
			headers, messageID, entity, err := c.consume(ctx)
			if err != nil {
				c.logger.Error("Error consuming message from stream", "error", err, "stream_id", c.streamCfg.ID, "group_id", c.streamCfg.Group.ID, "worker_id", c.WorkerID)

				jitter := time.Duration(rand.Int63n(int64(currentBackoff) / 5))
				sleepTime := currentBackoff + jitter

				c.logger.Info("Backoff active", "sleep_time", sleepTime.String())

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(sleepTime):
					// Time's up - continuing work
				}

				currentBackoff *= time.Duration(c.backoffCfg.Factor)
				if currentBackoff > c.backoffCfg.Max {
					currentBackoff = c.backoffCfg.Max
				}

				continue
			}

			currentBackoff = c.backoffCfg.Min
			if entity == "" {
				continue
			}

			parentCtx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(headers))

			tracer := otel.Tracer(c.contextPropagator.AppID)
			ctx, span := tracer.Start(parentCtx, c.contextPropagator.ProcessID,
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(attribute.String("redis.message_id", messageID)),
			)
			ctx = logger.WithMessageID(ctx, messageID)
			c.logger.InfoContext(ctx, "Received message from stream")

			err = c.usecase.Use(ctx, entity)
			span.End()
			if err == nil {
				ackCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				ackErr := c.ack(ackCtx, c.streamCfg.ID, c.streamCfg.Group.ID, messageID)
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
func (c *Consumer) consume(ctx context.Context) (map[string]string, string, string, error) {

	const undeliveredMessages = ">" // Redis specific alias: starts from the unconsumed message

	res, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Streams:  []string{c.streamCfg.ID, undeliveredMessages},
		Group:    c.streamCfg.Group.ID,
		Consumer: c.WorkerID,
		Count:    c.streamCfg.ReadCount,
		Block:    c.streamCfg.BlockTime,
	}).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, "", "", nil
		}
		if errors.Is(err, context.Canceled) {
			return nil, "", "", nil
		}

		c.logger.Error("Failed to consume message.", "error", err, "stream", c.streamCfg.ID, "group", c.streamCfg.Group.ID, "consumer", c.WorkerID)
		return nil, "", "", err
	}

	if len(res) == 0 || len(res[0].Messages) == 0 {
		return nil, "", "", nil
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
	return headers, message.ID, data, nil
}

func (c *Consumer) ack(ctx context.Context, stream, group, messageId string) error {
	_, err := c.client.XAck(ctx, stream, group, messageId).Result()
	if err != nil {
		c.logger.Error("Acknowledgment failed", "stream", stream, "group", group, "messageId", messageId, "error", err)
		return err
	}
	return nil
}
