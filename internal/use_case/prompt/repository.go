package prompt

import (
	"ai-orchestrator/internal/domain/model"
	"context"
	"errors"
)

var ErrNilRepository = errors.New("repository is nil")

type Repository interface {
	InsertPrompt(ctx context.Context, prompt model.Prompt) error
	UpdatePrompt(ctx context.Context, prompt model.Prompt) error
}
