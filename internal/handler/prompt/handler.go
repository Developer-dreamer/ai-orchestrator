package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/domain"
	"ai-orchestrator/internal/util"
	"context"
	"fmt"
	"net/http"
)

type Service interface {
	PostPrompt(ctx context.Context, prompt domain.Prompt) error
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

	userPrompt := &CreateRequest{}
	err := util.FromJSON(r.Body, userPrompt)
	if err != nil {
		h.logger.Warn("failed to decode request body", "error", err, "request_body", r.Body, "handler", "promptHandler.PostPrompt")
		util.WriteJSONError(rw, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	domainPrompt := userPrompt.ToDomain()
	err = h.service.PostPrompt(r.Context(), domainPrompt)
	response := FromDomain(domainPrompt, Accepted, "Processing started")
	if err != nil {
		h.logger.Warn("failed to post prompt", "error", err, "domainPrompt", domainPrompt)
		response.Status = Failed
		response.Message = fmt.Sprintf("failed to start processing prompt: %v", err)
	}

	util.WriteJSONResponse(rw, http.StatusAccepted, response)
}
