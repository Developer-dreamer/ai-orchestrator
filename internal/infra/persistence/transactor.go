package persistence

import (
	"ai-orchestrator/internal/common"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Transactor struct {
	logger common.Logger
	db     *sqlx.DB
}

func NewTransactor(logger common.Logger, db *sqlx.DB) *Transactor {
	return &Transactor{
		logger: logger,
		db:     db,
	}
}

func (t *Transactor) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	txCtx := context.WithValue(ctx, "tx", tx)
	err = tFunc(txCtx)

	return err
}
