package prompt

import (
	"ai-orchestrator/internal/domain/model"
	"github.com/google/uuid"
)

type CreateRequest struct {
	UserID  uuid.UUID `json:"user_id"`
	ModelID string    `json:"model_id"`
	Prompt  string    `json:"prompt"`
}

func (r *CreateRequest) ToDomain() model.Prompt {
	return model.Prompt{
		ID:      uuid.New(),
		UserID:  r.UserID,
		ModelID: r.ModelID,
		Text:    r.Prompt,
	}
}

type ResultResponse struct {
	PromptID uuid.UUID `json:"prompt_id"`
	UserID   uuid.UUID `json:"user_id"`
	Message  string    `json:"message"`
}

func FromDomain(domain model.Prompt, message string) ResultResponse {
	return ResultResponse{
		PromptID: domain.ID,
		UserID:   domain.UserID,
		Message:  message,
	}
}
