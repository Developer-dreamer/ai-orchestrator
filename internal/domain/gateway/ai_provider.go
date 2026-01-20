package gateway

import (
	"context"
	"errors"
)

var ErrNilProvider = errors.New("provider is nil")

type AIProvider interface {
	Generate(ctx context.Context, model, prompt string) (string, error)
}
