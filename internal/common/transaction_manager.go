package common

import "context"

type TransactionManager interface {
	WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error
}
