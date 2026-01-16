package prompt

import (
	"ai-orchestrator/internal/domain/model"
	"github.com/google/uuid"
	"time"
)

type Prompt struct {
	Id        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Text      string    `db:"text"`
	Response  string    `db:"response"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func FromDomain(d model.Prompt) Prompt {
	return Prompt{
		Id:       d.ID,
		UserID:   d.UserID,
		Text:     d.Text,
		Response: d.Response,
		Status:   string(d.Status),
	}
}
