package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/domain"
	"ai-orchestrator/internal/util"
	"context"
	"github.com/opentracing/opentracing-go"
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

	tracer := opentracing.GlobalTracer()

	spanContext, _ := tracer.Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header),
	)

	span := tracer.StartSpan(
		"post_prompt",
		opentracing.ChildOf(spanContext),
	)
	defer span.Finish()

	ctxWithTrace := opentracing.ContextWithSpan(r.Context(), span)

	userPrompt := &CreateRequest{}
	err := util.FromJSON(r.Body, userPrompt)
	if err != nil {
		h.logger.Warn("failed to decode request body", "error", err, "request_body", r.Body, "handler", "promptHandler.PostPrompt")
		util.WriteJSONError(rw, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	domainPrompt := userPrompt.ToDomain()
	err = h.service.PostPrompt(ctxWithTrace, domainPrompt)
	if err != nil {
		h.logger.Warn("failed to post prompt", "error", err, "domainPrompt", domainPrompt)
		util.WriteJSONError(rw, http.StatusInternalServerError, "failed to post prompt", err)
		return
	}

	response := FromDomain(domainPrompt, "Processing started")
	util.WriteJSONResponse(rw, http.StatusAccepted, response)
}
