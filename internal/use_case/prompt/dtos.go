package prompt

import (
	"ai-orchestrator/internal/domain/model"
	"ai-orchestrator/internal/infra/persistence/repository/outbox"
	"encoding/json"
	"github.com/google/uuid"
)

type TaskPayload struct {
	ID      uuid.UUID `json:"id"`
	UserID  uuid.UUID `json:"user_id"`
	ModelID string    `json:"model_id"`
	Text    string    `json:"text"`
}

type ResultPayload struct {
	ID       uuid.UUID `json:"id"`
	Response string    `json:"response"`
	Error    string    `json:"error,omitempty"`
}

type WebSocketResult struct {
	ID       uuid.UUID    `json:"id"`
	UserID   uuid.UUID    `json:"user_id"`
	ModelID  string       `json:"model_id"`
	Text     string       `json:"text"`
	Response string       `json:"response"`
	Status   model.Status `json:"status"`
	Error    string       `json:"error,omitempty"`
}

func (tp *TaskPayload) ToEvent(eventType string) outbox.Event {
	data, _ := json.Marshal(tp)

	return outbox.Event{
		ID:            uuid.New(),
		AggregateType: "prompt",
		AggregateID:   tp.ID,
		EventType:     eventType,
		Payload:       data,
		Status:        outbox.Pending,
		RetryCount:    0,
	}
}

func DomainToWebsocket(d *model.Prompt) WebSocketResult {
	return WebSocketResult{
		ID:       d.ID,
		UserID:   d.UserID,
		ModelID:  d.ModelID,
		Text:     d.Text,
		Response: d.Response,
		Status:   d.Status,
		Error:    d.Error,
	}
}
