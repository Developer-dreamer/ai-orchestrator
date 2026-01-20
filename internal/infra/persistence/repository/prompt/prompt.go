package prompt

import (
	"ai-orchestrator/internal/domain/model"
	"github.com/google/uuid"
	"time"
)

type Prompt struct {
	ID        uuid.UUID    `db:"id"`
	UserID    uuid.UUID    `db:"user_id"`
	ModelID   string       `db:"model_id"`
	Text      string       `db:"text"`
	Response  string       `db:"response"`
	Status    model.Status `db:"status"`
	Error     string       `db:"error"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
}

func FromDomain(d model.Prompt) Prompt {
	return Prompt{
		ID:       d.ID,
		UserID:   d.UserID,
		ModelID:  d.ModelID,
		Text:     d.Text,
		Response: d.Response,
		Status:   d.Status,
		Error:    d.Error,
	}
}

func (p *Prompt) ToDomain() model.Prompt {
	return model.Prompt{
		ID:       p.ID,
		UserID:   p.UserID,
		ModelID:  p.ModelID,
		Text:     p.Text,
		Response: p.Response,
		Status:   p.Status,
		Error:    p.Error,
	}
}
