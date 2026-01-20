package gemini

import (
	"ai-orchestrator/internal/common/logger"
	"context"
	"errors"
	"google.golang.org/genai"
)

type Client struct {
	logger logger.Logger
	client *genai.Client
}

func NewClient(l logger.Logger, client *genai.Client) (*Client, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}
	if client == nil {
		return nil, errors.New("client is nil")
	}

	return &Client{
		logger: l,
		client: client,
	}, nil
}

func (c *Client) Generate(ctx context.Context, model, prompt string) (string, error) {
	if model == "" {
		model = "gemini-3-flash-preview"
	}

	// TODO implement retry policy
	result, err := c.client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		c.logger.ErrorContext(ctx, "Prompt to model failed.", "err", err, "model", model)
		return "", err
	}

	return result.Text(), nil
}
