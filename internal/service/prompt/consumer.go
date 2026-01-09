package prompt

import (
	"ai-orchestrator/internal/common"
	"context"
	"github.com/google/uuid"
	"time"
)

type TaskConsumer interface {
	CreateGroup(ctx context.Context, stream, group string) error
	Consume(ctx context.Context, stream, consumer, group string) (string, string, error)
	Ack(ctx context.Context, stream, group, messageId string) error
}

type Consumer struct {
	logger   common.Logger
	tasks    TaskConsumer
	streamID string
	WorkerID string
	groupID  string
}

func NewConsumer(logger common.Logger, tasks TaskConsumer, redisStreamID, group, worker string) *Consumer {
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
			messageID, entity, err := c.tasks.Consume(ctx, c.streamID, c.WorkerID, c.groupID)
			if err != nil {
				c.logger.Error("Error consuming message from stream", "error", err, "stream_id", c.streamID, "group_id", c.groupID, "worker_id", c.WorkerID)
				// TODO implement exponential backoff
				time.Sleep(time.Second)
				continue
			}

			if entity == "" {
				continue
			}

			// Process the message before acknowledging it.
			if err := c.processMessage(ctx, messageID, entity); err != nil {
				c.logger.Error("Failed to process message", "message_id", messageID, "error", err)
				continue
			}

			if err := c.tasks.Ack(ctx, c.streamID, c.groupID, messageID); err != nil {
				c.logger.Error("Failed to ack message", "message_id", messageID, "error", err)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, messageID, entity string) error {
	// TODO: replace this placeholder with actual business logic for processing the message entity.
	c.logger.Info("Processing message", "message_id", messageID, "entity", entity)
	return nil
}
