package gemini

import (
	"ai-orchestrator/internal/common"
	"context"
	"google.golang.org/genai"
)

type Client struct {
	logger common.Logger
	client *genai.Client
}

func NewClient(logger common.Logger, client *genai.Client) *Client {
	return &Client{
		logger: logger,
		client: client,
	}
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
