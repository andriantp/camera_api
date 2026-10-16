package user

import (
	"context"

	"camera_api/pkg/adapter/database/postgres"

	"gorm.io/gorm"
)

type Repository interface {
	Transaction
	Role
	User
	Session
	Activity
}

type Transaction interface {
	DoTransaction(ctx context.Context, fn func(ctxWithTx context.Context, _ *gorm.DB) error) error
}

type repo struct {
	provider postgres.Database
}

func NewRepository(provider postgres.Database) Repository {
	return &repo{
		provider: provider,
	}
}

func (r *repo) DoTransaction(ctx context.Context, fn func(ctxWithTx context.Context, _ *gorm.DB) error) error {
	return r.provider.WithTransaction(ctx, fn)
}
