package config

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func ConnectToPostgres(postgresUri string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", postgresUri)
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
