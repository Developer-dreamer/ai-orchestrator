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

type Status string

const (
	Accepted Status = "accepted"
	Failed   Status = "failed"
)

type ResultResponse struct {
	PromptID uuid.UUID `json:"prompt_id"`
	UserID   uuid.UUID `json:"user_id"`
	Status   Status    `json:"status"`
	Message  string    `json:"message"`
}

func FromDomain(domain domain.Prompt, status Status, message string) ResultResponse {
	return ResultResponse{
		PromptID: domain.ID,
		UserID:   domain.UserID,
		Status:   status,
		Message:  message,
	}
}
