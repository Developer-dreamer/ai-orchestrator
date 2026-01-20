package connector

import (
	"ai-orchestrator/internal/config/api"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func ConnectToPostgres(cfg api.PostgresConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("postgres connection failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sqlx.DB, migrationDir string) error {
	return goose.Up(db.DB, migrationDir)
}
