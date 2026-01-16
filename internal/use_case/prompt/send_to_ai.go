package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/domain/gateway"
	"ai-orchestrator/internal/domain/model"
	"context"
	"encoding/json"
)

type Producer interface {
	Publish(ctx context.Context, data json.RawMessage) error
}

type SendPromptUsecase struct {
	logger     common.Logger
	aiProvider gateway.AIProvider
	producer   Producer
}

func NewSendPrompUsecase(l common.Logger, provider gateway.AIProvider, producer Producer) *SendPromptUsecase {
	return &SendPromptUsecase{
		logger:     l,
		aiProvider: provider,
		producer:   producer,
	}
}

func (uc *SendPromptUsecase) Use(ctx context.Context, messageID, entity string) error {
	uc.logger.InfoContext(ctx, "Processing message", "messageID", messageID)

	userPrompt := &model.Prompt{}
	err := json.Unmarshal([]byte(entity), userPrompt)
	if err != nil {
		uc.logger.WarnContext(ctx, "failed to decode incoming entity", "error", err)
		return err
	}

	res, err := uc.aiProvider.Generate(ctx, userPrompt.ModelID, userPrompt.Text)
	if err != nil {
		return err
	}

	uc.logger.InfoContext(ctx, "Received the result", "messageID", messageID, "response", res)
	userPrompt.Response = res

	resultJson, err := json.Marshal(userPrompt)
	if err != nil {
		uc.logger.WarnContext(ctx, "failed to marshal result", "error", err)
		return err
	}

	err = uc.producer.Publish(ctx, resultJson)
	if err != nil {
		uc.logger.WarnContext(ctx, "failed to publish result", "error", err)
		return err
	}

	return nil
}
