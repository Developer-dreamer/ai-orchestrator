package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/domain/model"
	"context"
	"encoding/json"
)

type SocketProvider interface {
	SendToClient(ctx context.Context, userID string, data json.RawMessage) error
}

type SaveResponse struct {
	logger common.Logger
	socket SocketProvider
	repo   Repository
}

func NewSaveResponse(logger common.Logger, socket SocketProvider, repo Repository) *SaveResponse {
	return &SaveResponse{
		logger: logger,
		socket: socket,
		repo:   repo,
	}
}

func (sr *SaveResponse) Use(ctx context.Context, messageID, entity string) error {
	userPrompt := &model.Prompt{}
	err := json.Unmarshal([]byte(entity), userPrompt)
	if err != nil {
		sr.logger.WarnContext(ctx, "failed to decode incoming entity", "error", err)
		return err
	}

	userPrompt.Status = model.Completed

	err = sr.repo.UpdatePrompt(ctx, *userPrompt)
	if err != nil {
		sr.logger.WarnContext(ctx, "failed to save prompt", "error", err)
		return err
	}

	err = sr.socket.SendToClient(ctx, userPrompt.UserID.String(), []byte(entity))
	if err != nil {
		sr.logger.WarnContext(ctx, "failed to save prompt", "error", err)
		return err
	}

	return nil
}
