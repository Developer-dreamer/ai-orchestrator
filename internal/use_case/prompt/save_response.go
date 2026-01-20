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
	result := &ResultPayload{}
	err := json.Unmarshal([]byte(entity), result)
	if err != nil {
		sr.logger.ErrorContext(ctx, "failed to decode incoming entity", "error", err)
		return err
	}

	domainPrompt, err := sr.repo.GetPromptByID(ctx, result.ID)
	if err != nil {
		sr.logger.ErrorContext(ctx, "failed to get prompt by id", "error", err)
		return err
	}

	domainPrompt.Response = result.Response
	if result.Error != "" {
		domainPrompt.Status = model.Failed
		domainPrompt.Error = result.Error
	} else {
		domainPrompt.Status = model.Completed
	}

	err = sr.repo.UpdatePrompt(ctx, *domainPrompt)
	if err != nil {
		sr.logger.WarnContext(ctx, "failed to save prompt", "error", err)
		return err
	}

	wsResult := DomainToWebsocket(domainPrompt)
	wsJson, err := json.Marshal(wsResult)
	if err != nil {
		sr.logger.ErrorContext(ctx, "failed to marshal user prompt", "error", err)
		return err
	}
	err = sr.socket.SendToClient(ctx, domainPrompt.UserID.String(), wsJson)
	if err != nil {
		sr.logger.WarnContext(ctx, "failed to send result to client", "error", err)
		return err
	}

	return nil
}
