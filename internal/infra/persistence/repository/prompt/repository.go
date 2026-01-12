package prompt

import (
	"ai-orchestrator/internal/common"
	"ai-orchestrator/internal/domain/model"
	"context"
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
	dbPrompt.CreatedAt = time.Now()
	dbPrompt.UpdatedAt = dbPrompt.CreatedAt

	query := `
		INSERT INTO prompts (id, user_id, text, status, created_at, updated_at)
		VALUES (:id, :user_id, :text, :status, :created_at, :updated_at)
	`

	r.logger.Info("executing query to insert new prompt", "query", query, "repository", "promptRepository")

	_, err := r.db.NamedExecContext(ctx, query, dbPrompt)
	if err != nil {
		r.logger.Error("failed to insert new prompt", "error", err)
		return err
	}

	return nil
}
