package prompt

import (
	"github.com/google/uuid"
	"time"
)

type Prompt struct {
	Id        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Text      string    `db:"text"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
