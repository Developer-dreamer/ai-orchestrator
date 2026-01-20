package prompt

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/domain/model"
	"context"
	"encoding/json"
	"errors"
)

var ErrNilSocket = errors.New("socket is nil")

type SocketProvider interface {
	SendToClient(ctx context.Context, userID string, data json.RawMessage) error
}

type SaveResponse struct {
	logger logger.Logger
	socket SocketProvider
	repo   Repository
}

func NewSaveResponse(l logger.Logger, socket SocketProvider, repo Repository) (*SaveResponse, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}
	if socket == nil {
		return nil, ErrNilSocket
	}
	if repo == nil {
		return nil, ErrNilRepository
	}

	return &SaveResponse{
		logger: l,
		socket: socket,
		repo:   repo,
	}, nil
}

func (sr *SaveResponse) Use(ctx context.Context, entity string) error {
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
