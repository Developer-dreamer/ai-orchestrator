package model

import "github.com/google/uuid"

type Status string

var (
	Accepted   Status = "Accepted"
	Discarded  Status = "Discarded"
	Processing Status = "Processing"
	Completed  Status = "Completed"
	Failed     Status = "Failed"
)

type Prompt struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	ModelID string
	Text    string
	Status  Status
}
