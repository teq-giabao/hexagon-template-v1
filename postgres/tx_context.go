package postgres

import (
	"context"

	"gorm.io/gorm"
)

type txContextKey struct{}

func withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func txFromContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return nil
	}

	tx, ok := ctx.Value(txContextKey{}).(*gorm.DB)
	if !ok {
		return nil
	}

	return tx
}
