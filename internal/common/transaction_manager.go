package common

import (
	"context"
	"errors"
)

type TransactionManager interface {
	WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error
}

var (
	ErrNilTransactionManager = errors.New("nil transaction manager")
)
