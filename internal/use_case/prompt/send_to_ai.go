package prompt

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/domain/gateway"
	"context"
	"encoding/json"
	"errors"
)

type Producer interface {
	Publish(ctx context.Context, data json.RawMessage) error
}

type SendPromptUsecase struct {
	logger     logger.Logger
	aiProvider gateway.AIProvider
	producer   Producer
}

func NewSendPromptUsecase(l logger.Logger, provider gateway.AIProvider, producer Producer) (*SendPromptUsecase, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}
	if provider == nil {
		return nil, gateway.ErrNilProvider
	}
	if producer == nil {
		return nil, errors.New("producer is nil")
	}

	return &SendPromptUsecase{
		logger:     l,
		aiProvider: provider,
		producer:   producer,
	}, nil
}

func (uc *SendPromptUsecase) Use(ctx context.Context, entity string) error {
	uc.logger.InfoContext(ctx, "Processing message")

	userPrompt := &TaskPayload{}
	err := json.Unmarshal([]byte(entity), userPrompt)
	if err != nil {
		uc.logger.WarnContext(ctx, "failed to decode incoming entity", "error", err)
		return err
	}

	res, err := uc.aiProvider.Generate(ctx, userPrompt.ModelID, userPrompt.Text)

	uc.logger.InfoContext(ctx, "Received the result", "response", res)
	resultPayload := &ResultPayload{
		ID:       userPrompt.ID,
		Response: res,
	}
	if err != nil {
		resultPayload.Error = err.Error()
	}

	resultJson, err := json.Marshal(resultPayload)
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
