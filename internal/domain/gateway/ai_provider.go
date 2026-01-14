package gateway

import "context"

type AIProvider interface {
	Generate(ctx context.Context, model, prompt string) (string, error)
}
