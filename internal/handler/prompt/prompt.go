package prompt

import (
	"ai-orchestrator/internal/util"
	"net/http"
)

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
}

type Service interface {
}

type Handler struct {
	logger Logger
}

type CreateRequest struct {
	UserID string `json:"user_id"`
	Prompt string `json:"prompt"`
}

func NewHandler(l Logger) *Handler {
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
