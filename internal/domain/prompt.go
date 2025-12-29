package prompt

import "github.com/google/uuid"

type Prompt struct {
	ID     uuid.UUID
	UserID string
	Text   string
}
