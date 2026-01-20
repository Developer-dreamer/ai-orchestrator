package outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Status string

var (
	Pending   Status = "pending"
	Failed    Status = "failed"
	Processed Status = "processed"
)

type Event struct {
	ID            uuid.UUID `db:"id"`
	AggregateType string    `db:"aggregate_type"`
	AggregateID   uuid.UUID `db:"aggregate_id"`
	EventType     string    `db:"event_type"`

	// Payload maps to JSONB. json.RawMessage allows delaying parsing
	// until you know the specific struct type in the worker.
	Payload json.RawMessage `db:"payload"`

	Status Status `db:"status"`

	TraceID    string `db:"trace_id"`
	RetryCount int    `db:"retry_count"`

	// Use pointers for Nullable columns
	ErrorMessage *string `db:"error_message"`

	CreatedAt   time.Time  `db:"created_at"`
	ProcessedAt *time.Time `db:"processed_at"`
}
