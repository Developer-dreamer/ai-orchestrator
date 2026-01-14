package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/domain/gateway"
	"ai-orchestrator/internal/domain/model"
	"context"
	"encoding/json"
)

type SendPromptUsecase struct {
	logger     common.Logger
	aiProvider gateway.AIProvider
}

func NewSendPrompUsecase(l common.Logger, provider gateway.AIProvider) *SendPromptUsecase {
	return &SendPromptUsecase{
		logger:     l,
		aiProvider: provider,
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
	return nil
}
