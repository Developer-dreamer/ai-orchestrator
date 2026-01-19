package stream

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"math/rand"
	"strings"
	"time"
)

type ResConsumer struct {
	logger   common.Logger
	client   *redis.Client
	config   *config.Stream
	streamID string
	WorkerID string
	groupID  string

	usecase UseCase
}

func NewResConsumer(logger common.Logger, client *redis.Client, cfg *config.Stream, usecase UseCase, redisStreamID, groupID, workerID string) *ResConsumer {
	return &ResConsumer{
		logger:   logger,
		client:   client,
		config:   cfg,
		usecase:  usecase,
		streamID: redisStreamID,
		groupID:  groupID,
		WorkerID: workerID,
	}
}

func (c *ResConsumer) ConsumeResult(ctx context.Context) error {
	c.logger.Info("Worker started")

	err := c.createGroupRes(ctx, c.streamID, c.groupID)
	if err != nil {
		c.logger.Error("Failed to create group", "id", c.WorkerID, "err", err)
		return err
	}

	currentBackoff := minBackoff

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping consumer")
			return ctx.Err()
		default:
			headers, messageID, entity, err := c.consumeResult(ctx, c.streamID)
			if err != nil {
				c.logger.Error("Error consuming message from stream", "error", err, "stream_id", c.streamID, "group_id")

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

			tracer := otel.Tracer("api")
			ctx, span := tracer.Start(parentCtx, "send_result",
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(attribute.String("redis.message_id", messageID)),
			)
			c.logger.InfoContext(ctx, "Received message from stream", "message_id", messageID)

			err = c.usecase.Use(ctx, messageID, entity)
			span.End()
			if err == nil {
				ackCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				ackErr := c.ackRes(ackCtx, messageID)
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

func (c *ResConsumer) createGroupRes(ctx context.Context, stream, group string) error {

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
func (c *ResConsumer) consumeResult(ctx context.Context, stream string) (map[string]string, string, string, error) {

	const undeliveredMessages = ">" // Redis specific alias: starts from the unconsumed message

	res, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Streams:  []string{stream, undeliveredMessages},
		Group:    c.groupID,
		Consumer: c.WorkerID,
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

		c.logger.Error("Failed to consume message.", "error", err, "stream", stream)
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

	c.logger.Debug("Received message", "stream", stream, "data", data)
	return headers, message.ID, data, nil
}

func (c *ResConsumer) ackRes(ctx context.Context, messageId string) error {
	_, err := c.client.XAck(ctx, c.streamID, c.groupID, messageId).Result()
	if err != nil {
		c.logger.Error("Acknowledgment failed", "stream", c.streamID, "group", c.groupID, "messageId", messageId, "error", err)
		return err
	}
	return nil
}
