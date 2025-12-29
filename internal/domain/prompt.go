package domain

import "github.com/google/uuid"

type Prompt struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Text   string
}
