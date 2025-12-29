package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/util"
	"net/http"
)

type Handler struct {
	logger common.Logger
}

func NewHandler(l common.Logger) *Handler {
	return &Handler{l}
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

	// process with service

	util.WriteJSONResponse(rw, http.StatusAccepted, &struct {
		Message string
	}{
		Message: "Prompt accepted",
	})
}
