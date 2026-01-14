package stream

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"math/rand"
	"strings"
	"time"
)

const (
	minBackoff    = 1 * time.Second
	maxBackoff    = 60 * time.Second
	backoffFactor = 2
)

type UseCase interface {
	Use(ctx context.Context, messageID, entity string) error
}
type Consumer struct {
	logger   common.Logger
	usecase  UseCase
	client   *redis.Client
	config   *config.Stream
	streamID string
	WorkerID string
	groupID  string
}

func NewConsumer(logger common.Logger, usecase UseCase, client *redis.Client, cfg *config.Stream, redisStreamID, group, worker string) *Consumer {
	if worker == "" {
		worker = "worker-" + uuid.New().String()
	}
	if group == "" {
		group = "ai_tasks_group"
	}

	return &Consumer{
		logger:   logger,
		usecase:  usecase,
		client:   client,
		config:   cfg,
		streamID: redisStreamID,
		groupID:  group,
		WorkerID: worker,
	}
}

func (c *Consumer) Consume(ctx context.Context) error {
	c.logger.Info("Worker started", "id", c.WorkerID)

	currentBackoff := minBackoff

	err := c.createGroup(ctx, c.streamID, c.groupID)
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
			headers, messageID, entity, err := c.consume(ctx, c.streamID, c.WorkerID, c.groupID)
			if err != nil {
				c.logger.Error("Error consuming message from stream", "error", err, "stream_id", c.streamID, "group_id", c.groupID, "worker_id", c.WorkerID)

				jitter := time.Duration(rand.Int63n(int64(currentBackoff) / 5))
				sleepTime := currentBackoff + jitter

				c.logger.Info("Backoff active", "sleep_time", sleepTime.String())

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(sleepTime):
					// Time's up - continuing work
				}

				currentBackoff *= time.Duration(backoffFactor)
				if currentBackoff > maxBackoff {
					currentBackoff = maxBackoff
				}

				continue
			}

			currentBackoff = minBackoff
			if entity == "" {
				continue
			}

			parentCtx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(headers))

			tracer := otel.Tracer("worker")
			ctx, span := tracer.Start(parentCtx, "worker_process",
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(attribute.String("redis.message_id", messageID)),
			)
			c.logger.InfoContext(ctx, "Received message from stream", "message_id", messageID)

			err = c.usecase.Use(ctx, messageID, entity)
			span.End()
			if err == nil {
				ackCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				ackErr := c.ack(ackCtx, c.streamID, c.groupID, messageID)
				cancel()

				if ackErr != nil {
					c.logger.Error("Failed to ack message", "message_id", messageID, "error", ackErr)
				}
				continue
			}
			c.logger.ErrorContext(ctx, "Failed to process message", "message_id", messageID, "error", err)
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

// Consume Consumes a message from the specified stream. Returns Headers, MessageID, Data, Error
func (c *Consumer) consume(ctx context.Context, stream, consumer, group string) (map[string]string, string, string, error) {

	const undeliveredMessages = ">" // Redis specific alias: starts from the unconsumed message

	res, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Streams:  []string{stream, undeliveredMessages},
		Group:    group,
		Consumer: consumer,
		Count:    c.config.ReadCount,
		Block:    c.config.BlockTime,
	}).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, "", "", nil
		}
		if errors.Is(err, context.Canceled) {
			return nil, "", "", nil
		}

		c.logger.Error("Failed to consume message.", "error", err, "stream", stream, "group", group, "consumer", consumer)
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

	c.logger.Debug("Received message", "stream", stream, "group", group, "consumer", consumer, "data", data)
	return headers, message.ID, data, nil
}

// Ack Acknowledges a processed message by ID
func (c *Consumer) ack(ctx context.Context, stream, group, messageId string) error {
	_, err := c.client.XAck(ctx, stream, group, messageId).Result()
	if err != nil {
		c.logger.Error("Acknowledgment failed", "stream", stream, "group", group, "messageId", messageId, "error", err)
		return err
	}
	return nil
}
