package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/config"
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
	workerID string
	groupID  string
}

func NewConsumer(logger common.Logger, tasks TaskConsumer, group, worker string, cfg *config.Config) *Consumer {
	if worker == "" {
		worker = "worker-" + uuid.New().String()
	}
	if group == "" {
		group = "ai_tasks_group"
	}

	return &Consumer{
		logger:   logger,
		tasks:    tasks,
		streamID: cfg.RedisStreamID,
		groupID:  group,
		workerID: worker,
	}
}

func (c *Consumer) Consume(ctx context.Context) error {
	err := c.tasks.CreateGroup(context.Background(), c.streamID, c.groupID)
	if err != nil {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping consumer", "worker_id", c.workerID)
			return ctx.Err()
		default:
			messageID, entity, err := c.tasks.Consume(ctx, c.streamID, c.workerID, c.groupID)
			if err != nil {
				c.logger.Error("Read error", "error", err)
				time.Sleep(time.Second)
				continue
			}

			if entity == "" {
				continue
			}

			// TODO: process
			c.logger.Info("Read message", "message_id", messageID)

			c.tasks.Ack(ctx, c.streamID, c.groupID, messageID)
		}
	}
}
