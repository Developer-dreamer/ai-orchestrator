package prompt

import (
	"ai-orchestrator/internal/domain/model"
	"context"
	"errors"
	"github.com/google/uuid"
)

var ErrNilRepository = errors.New("repository is nil")

type Repository interface {
	GetPromptByID(ctx context.Context, id uuid.UUID) (*model.Prompt, error)
	InsertPrompt(ctx context.Context, prompt model.Prompt) error
	UpdatePrompt(ctx context.Context, prompt model.Prompt) error
}
