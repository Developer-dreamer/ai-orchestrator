package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/domain/model"
	"ai-orchestrator/internal/infra/telemetry/tracing"
	"ai-orchestrator/internal/transport/http/helper"
	"context"
	"net/http"
)

type Service interface {
	PostPrompt(ctx context.Context, prompt model.Prompt) error
}

type Handler struct {
	logger  common.Logger
	service Service
}

func NewHandler(l common.Logger, s Service) *Handler {
	return &Handler{
		logger:  l,
		service: s,
	}
}

func (h *Handler) PostPrompt(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Incoming request:", "path", "promptHandler.PostPrompt")

	span, ctx := tracing.InitContextFromHttp(r, "post_prompt")
	defer span.Finish()

	userPrompt := &CreateRequest{}
	err := helper.FromJSON(r.Body, userPrompt)
	if err != nil {
		h.logger.Warn("failed to decode request body", "error", err, "request_body", r.Body, "handler", "promptHandler.PostPrompt")
		helper.WriteJSONError(rw, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	domainPrompt := userPrompt.ToDomain()
	err = h.service.PostPrompt(ctx, domainPrompt)
	if err != nil {
		h.logger.Warn("failed to post prompt", "error", err, "domainPrompt", domainPrompt)
		helper.WriteJSONError(rw, http.StatusInternalServerError, "failed to post prompt", err)
		return
	}

	response := FromDomain(domainPrompt, "Processing started")
	helper.WriteJSONResponse(rw, http.StatusAccepted, response)
}
