package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/domain/model"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

type Repository struct {
	logger common.Logger
	db     *sqlx.DB
}

func NewRepository(logger common.Logger, db *sqlx.DB) *Repository {
	return &Repository{
		logger: logger,
		db:     db,
	}
}

func (r *Repository) InsertPrompt(ctx context.Context, prompt model.Prompt) error {
	dbPrompt := FromDomain(prompt)
	dbPrompt.CreatedAt = time.Now().UTC()
	dbPrompt.UpdatedAt = dbPrompt.CreatedAt

	query := `
		INSERT INTO prompts (id, user_id, text, response, status, created_at, updated_at)
		VALUES (:id, :user_id, :text, :response, :status, :created_at, :updated_at)
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
		return fmt.Errorf("prompt not found to update")
	}

	return nil
}
