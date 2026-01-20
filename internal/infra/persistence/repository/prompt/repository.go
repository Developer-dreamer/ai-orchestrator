package prompt

import (
	"ai-orchestrator/internal/common/logger"
	"ai-orchestrator/internal/domain/model"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
)

var ErrPromptNotFound = errors.New("prompt not found")

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

func (r *Repository) GetPromptByID(ctx context.Context, id uuid.UUID) (*model.Prompt, error) {
	var prompt Prompt
	query := `
		SELECT id, user_id, model_id, text, response, status, error, created_at, updated_at 
		FROM prompts 
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &prompt, query, id)
	if err != nil {
		return nil, err
	}

	domainPrompt := prompt.ToDomain()
	return &domainPrompt, err
}

func (r *Repository) InsertPrompt(ctx context.Context, prompt model.Prompt) error {
	dbPrompt := FromDomain(prompt)
	dbPrompt.CreatedAt = time.Now().UTC()
	dbPrompt.UpdatedAt = dbPrompt.CreatedAt

	query := `
		INSERT INTO prompts (id, user_id, model_id, text, response, status, error, created_at, updated_at)
		VALUES (:id, :user_id, :model_id, :text, :response, :status, :error, :created_at, :updated_at)
	`

	r.logger.InfoContext(ctx, "executing query to insert new prompt", "query", query, "repository", "promptRepository")

	_, err := r.db.NamedExecContext(ctx, query, dbPrompt)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to insert new prompt", "error", err)
		return err
	}

	return nil
}

func (r *Repository) UpdatePrompt(ctx context.Context, prompt model.Prompt) error {
	dbPrompt := FromDomain(prompt)
	dbPrompt.UpdatedAt = time.Now().UTC()

	query := `
        UPDATE prompts 
        SET response = :response, 
            status = :status,
            error = :error,
            updated_at = :updated_at
        WHERE id = :id
    `

	r.logger.InfoContext(ctx, "executing query to update prompt", "query", query, "prompt_id", dbPrompt.ID)

	result, err := r.db.NamedExecContext(ctx, query, dbPrompt)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to update prompt", "error", err, "id", dbPrompt.ID)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrPromptNotFound
	}

	return nil
}
