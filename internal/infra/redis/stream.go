package redis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var (
	ErrCreateGroup          = errors.New("create group error")
	ErrPublishFailed        = errors.New("failed to publish message")
	ErrInvalidPublishEntity = errors.New("invalid publish entity")
	ErrReadFromGroup        = errors.New("read group error")
	ErrAcknowledgmentFailed = errors.New("acknowledgement error")
	ErrParsingMessage       = errors.New("error when parsing consumed message")
)

type StreamConfig struct {
	MaxBacklog   int64
	UseDelApprox bool
	ReadCount    int64
	BlockTime    time.Duration
}

// CreateGroup Creates consumer group and stream (if the last one does not exist).
func (s *Service) CreateGroup(ctx context.Context, stream, group string) error {

	const EarliestMessage = "0" // Redis specific alias: start from the beginning of the stream

	_, err := s.client.XGroupCreateMkStream(ctx, stream, group, EarliestMessage).Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		s.logger.Warn("Failed to create consumer group.", "error", err, "stream", stream, "group", group)
		return ErrCreateGroup
	}
	s.logger.Info("Created consumer group", "stream", stream, "group", group)

	return nil
}

// Publish Publishes a message into the specified stream
func (s *Service) Publish(ctx context.Context, stream string, data any) error {
	dataStr, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("Failed to convert data to json.", "error", err, "data", data)
		return ErrInvalidPublishEntity
	}

	carrier := make(map[string]string)
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		err := opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.TextMap,
			opentracing.TextMapCarrier(carrier),
		)
		if err != nil {
			s.logger.Warn("failed to inject tracing context", "error", err)
		}
	}

	_, err = s.client.XAdd(ctx, &redis.XAddArgs{
		MaxLen: s.config.MaxBacklog,
		Approx: s.config.UseDelApprox,
		Stream: stream,
		Values: map[string]interface{}{
			"data":     dataStr,
			"trace_id": carrier["uber-trace-id"],
		},
	}).Result()

	if err != nil {
		s.logger.Error("Failed to publish message.", "error", err, "stream", stream, "data", data)
		return ErrPublishFailed
	}

	s.logger.Debug("Published message", "stream", stream, "data", data)
	return nil
}

// Consume Consumes a message from the specified stream. Returns TraceID, MessageID, Data, Error
func (s *Service) Consume(ctx context.Context, stream, consumer, group string) (string, string, string, error) {

	const undeliveredMessages = ">" // Redis specific alias: starts from the unconsumed message

	res, err := s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Streams:  []string{stream, undeliveredMessages},
		Group:    group,
		Consumer: consumer,
		Count:    s.config.ReadCount,
		Block:    s.config.BlockTime,
	}).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", "", "", nil
		}
		if errors.Is(err, context.Canceled) {
			return "", "", "", nil
		}

		s.logger.Error("Failed to consume message.", "error", err, "stream", stream, "group", group, "consumer", consumer)
		return "", "", "", errors.Join(ErrReadFromGroup, err)
	}

	if len(res) == 0 || len(res[0].Messages) == 0 {
		return "", "", "", nil
	}

	message := res[0].Messages[0]

	val, ok := message.Values["data"].(string)
	if !ok {
		s.logger.Error("Payload is not a string", "stream", stream, "group", group, "consumer", consumer)
		return "", "", "", ErrParsingMessage
	}

	traceID, _ := message.Values["trace_id"].(string)

	s.logger.Debug("Received message", "stream", stream, "group", group, "consumer", consumer, "data", val)
	return traceID, message.ID, val, nil
}

// Ack Acknowledges a processed message by ID
func (s *Service) Ack(ctx context.Context, stream, group, messageId string) error {
	_, err := s.client.XAck(ctx, stream, group, messageId).Result()
	if err != nil {
		s.logger.Error("Acknowledgment failed", "stream", stream, "group", group, "messageId", messageId, "error", err)
		return ErrAcknowledgmentFailed
	}
	return nil
}
