package prompt

import (
	"ai-orchestrator/internal/domain"
	"github.com/google/uuid"
)

type CreateRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Prompt string    `json:"prompt"`
}

func (r *CreateRequest) ToDomain() domain.Prompt {
	return domain.Prompt{
		ID:     uuid.New(),
		UserID: r.UserID,
		Text:   r.Prompt,
	}
}

type ResultResponse struct {
	PromptID uuid.UUID `json:"prompt_id"`
	UserID   uuid.UUID `json:"user_id"`
	Message  string    `json:"message"`
}

func FromDomain(domain domain.Prompt, message string) ResultResponse {
	return ResultResponse{
		PromptID: domain.ID,
		UserID:   domain.UserID,
		Message:  message,
	}
}
