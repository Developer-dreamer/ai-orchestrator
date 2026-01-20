package gemini

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/infra/manager"
	"context"
	"errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/genai"
	"net/http"
)

type Client struct {
	logger logger.Logger
	client *genai.Client

	backoff manager.Backoff
}

func NewClient(l logger.Logger, client *genai.Client, backoff manager.Backoff) (*Client, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}
	if client == nil {
		return nil, errors.New("client is nil")
	}

	return &Client{
		logger:  l,
		client:  client,
		backoff: backoff,
	}, nil
}

func (c *Client) Generate(ctx context.Context, model, prompt string) (string, error) {
	if model == "" {
		model = "gemini-3-flash-preview"
	}

	res, err := manager.WithBackoff[*genai.GenerateContentResponse](
		ctx,
		&c.backoff,
		func(ctx context.Context) (*genai.GenerateContentResponse, error) {
			return c.client.Models.GenerateContent(
				ctx,
				model,
				genai.Text(prompt),
				nil,
			)
		},
		IsRetryable)

	if err != nil {
		c.logger.ErrorContext(ctx, "Prompt to model failed.", "err", err, "model", model)
		return "", err
	}

	c.logger.DebugContext(ctx, "Prompt to model completed.", "response", res.Text())
	return res.Text(), nil
}

func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var e *googleapi.Error
	if errors.As(err, &e) {
		switch e.Code {
		case http.StatusTooManyRequests, // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout:      // 504
			return true
		default:
			// 400, 401, 403, 404 should NOT be retried
			return false
		}
	}

	return false
}
