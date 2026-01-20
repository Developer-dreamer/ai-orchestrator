package outbox

import (
	"ai-orchestrator/internal/common/logger"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type Repository struct {
	logger logger.Logger
	db     *sqlx.DB
}

func NewRepository(l logger.Logger, db *sqlx.DB) (*Repository, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}
	if db == nil {
		return nil, errors.New("db is nil")
	}
	return &Repository{
		logger: l,
		db:     db,
	}, nil
}

func (r *Repository) GetAllPendingEvents(ctx context.Context, count int) ([]Event, error) {
	query := `
        SELECT * FROM outbox 
        WHERE status = 'pending' 
        ORDER BY created_at ASC 
        LIMIT $1 
        FOR UPDATE SKIP LOCKED
    `

	var events []Event

	err := r.db.SelectContext(ctx, &events, query, count)
	if err != nil {
		r.logger.Error("failed to select pending events", "error", err)
		return nil, err
	}

	return events, nil
}

func (r *Repository) CreateEvent(ctx context.Context, event Event) error {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	if event.Status == "" {
		event.Status = "pending"
	}

	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()

	event.TraceID = traceID

	query := `
       INSERT INTO outbox (
           id, aggregate_type, aggregate_id, event_type, 
           payload, status, trace_id, retry_count, error_message, 
           created_at, processed_at
       )
       VALUES (
           :id, :aggregate_type, :aggregate_id, :event_type, 
           :payload, :status, :trace_id, :retry_count, :error_message, 
           :created_at, :processed_at
       )
    `

	_, err := r.db.NamedExecContext(ctx, query, event)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to insert outbox event", "error", err, "event_id", event.ID)
		return err
	}

	return nil
}

func (r *Repository) ChangeEventStatus(ctx context.Context, eventID uuid.UUID, eventStatus Status) error {
	query := `
        UPDATE outbox 
        SET status = $2, 
            processed_at = NOW() 
        WHERE id = $1
    `

	r.logger.InfoContext(ctx, "marking event as processed", "event_id", eventID)

	result, err := r.db.ExecContext(ctx, query, eventID, eventStatus)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to mark event processed", "error", err, "event_id", eventID)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		r.logger.Warn("event not found or already processed", "event_id", eventID)
	}

	return nil
}

func (r *Repository) IncrementRetryCount(ctx context.Context, eventID uuid.UUID, errorMessage string) error {
	query := `
        UPDATE outbox 
        SET retry_count = retry_count + 1,
            error_message = $1
        WHERE id = $2
    `

	_, err := r.db.ExecContext(ctx, query, errorMessage, eventID)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to increment retry count", "error", err, "event_id", eventID)
		return err
	}

	return nil
}
