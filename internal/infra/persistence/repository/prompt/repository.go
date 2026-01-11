package prompt

import (
	"ai-orchestrator/internal/common"
	"github.com/jmoiron/sqlx"
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
