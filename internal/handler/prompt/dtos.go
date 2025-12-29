package prompt

import (
	"ai-orchestrator/internal/domain"
	"github.com/google/uuid"
)

type CreateRequest struct {
	UserID string `json:"user_id"`
	Prompt string `json:"prompt"`
}

func (r *CreateRequest) ToDomain() domain.Prompt {
	return domain.Prompt{
		ID:     uuid.New(),
		UserID: r.UserID,
		Text:   r.Prompt,
	}
}
