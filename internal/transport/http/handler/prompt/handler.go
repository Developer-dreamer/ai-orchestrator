package prompt

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/domain/model"
	"ai-orchestrator/internal/infra/telemetry/tracing"
	"ai-orchestrator/internal/transport/http/helper"
	"context"
	"errors"
	"net/http"
)

var ErrNilService = errors.New("service is nil")

type Service interface {
	PostPrompt(ctx context.Context, prompt model.Prompt) error
}

type Handler struct {
	logger  logger.Logger
	service Service
}

func NewHandler(l logger.Logger, s Service) (*Handler, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}
	if s == nil {
		return nil, ErrNilService
	}

	return &Handler{
		logger:  l,
		service: s,
	}, nil
}

func (h *Handler) PostPrompt(rw http.ResponseWriter, r *http.Request) {
	span, ctx := tracing.InitContextFromHttp(r, "post_prompt")
	defer span.End()
	h.logger.InfoContext(ctx, "Incoming request:", "path", "promptHandler.PostPrompt")

	userPrompt := &CreateRequest{}
	err := helper.FromJSON(r.Body, userPrompt)
	if err != nil {
		h.logger.WarnContext(ctx, "failed to decode request body", "error", err, "request_body", r.Body, "handler", "promptHandler.PostPrompt")
		helper.WriteJSONError(rw, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	domainPrompt := userPrompt.ToDomain()
	err = h.service.PostPrompt(ctx, domainPrompt)
	if err != nil {
		h.logger.WarnContext(ctx, "failed to post prompt", "error", err, "domainPrompt", domainPrompt)
		helper.WriteJSONError(rw, http.StatusInternalServerError, "failed to post prompt", err)
		return
	}

	response := FromDomain(domainPrompt, "Processing started")
	helper.WriteJSONResponse(rw, http.StatusAccepted, response)
}
