package stream

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/infra/jaeger"
	"ai-orchestrator/internal/service/prompt"
	"context"
	"github.com/google/uuid"
	"time"
)

type BrokerConsumer interface {
	CreateGroup(ctx context.Context, stream, group string) error
	Consume(ctx context.Context, stream, consumer, group string) (string, string, string, error)
	Ack(ctx context.Context, stream, group, messageId string) error
}

type Consumer struct {
	logger   common.Logger
	tasks    BrokerConsumer
	streamID string
	WorkerID string
	groupID  string
}

func NewConsumer(logger common.Logger, tasks BrokerConsumer, redisStreamID, group, worker string) *Consumer {
	if worker == "" {
		worker = "worker-" + uuid.New().String()
	}
	if group == "" {
		group = "ai_tasks_group"
	}

	return &Consumer{
		logger:   logger,
		tasks:    tasks,
		streamID: redisStreamID,
		groupID:  group,
		WorkerID: worker,
	}
}

func (c *Consumer) Consume(ctx context.Context) error {
	c.logger.Info("Worker started", "id", c.WorkerID)

	err := c.tasks.CreateGroup(ctx, c.streamID, c.groupID)
	if err != nil {
		return nil // TODO refactor
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping consumer", "worker_id", c.WorkerID)
			return ctx.Err()
		default:
			traceId, messageID, entity, err := c.tasks.Consume(ctx, c.streamID, c.WorkerID, c.groupID)
			if err != nil {
				c.logger.Error("Error consuming message from stream", "error", err, "stream_id", c.streamID, "group_id", c.groupID, "worker_id", c.WorkerID)
				// TODO implement exponential backoff
				time.Sleep(time.Second)
				continue
			}

			if entity == "" {
				continue
			}

			c.logger.Info("Received message from stream", "trace_id", traceId, "message_id", messageID)

			span, traceContext := jaeger.InitContext(ctx, traceId, "worker_process_task")
			err = prompt.SendPromptUseCase(traceContext, messageID, entity)
			span.Finish()
			if err == nil {
				ackCtx, cancel := context.WithTimeout(traceContext, 2*time.Second)
				ackErr := c.tasks.Ack(ackCtx, c.streamID, c.groupID, messageID)
				cancel()

				if ackErr != nil {
					c.logger.Error("Failed to ack message", "message_id", messageID, "error", ackErr)
				}
				continue
			}
			c.logger.Error("Failed to process message", "message_id", messageID, "error", err)
		}
	}
}
