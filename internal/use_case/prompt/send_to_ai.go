package prompt

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/domain/gateway"
	"ai-orchestrator/internal/domain/model"
	"context"
	"encoding/json"
)

type Producer interface {
	Publish(ctx context.Context, data json.RawMessage) error
}

type SendPromptUsecase struct {
	logger     logger.Logger
	aiProvider gateway.AIProvider
	producer   Producer
}

func NewSendPrompUsecase(l logger.Logger, provider gateway.AIProvider, producer Producer) (*SendPromptUsecase, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}
	if provider == nil {
		return nil, gateway.ErrNilProvider
	}

	return &SendPromptUsecase{
		logger:     l,
		aiProvider: provider,
		producer:   producer,
	}, nil
}

func (uc *SendPromptUsecase) Use(ctx context.Context, entity string) error {
	uc.logger.InfoContext(ctx, "Processing message")

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

	uc.logger.InfoContext(ctx, "Received the result", "response", res)
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
