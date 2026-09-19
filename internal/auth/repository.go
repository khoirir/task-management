package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoirir/task-management/internal/database/db"
)

type Repository interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error)
}

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{
		queries: queries,
	}
}

func (r *repository) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return r.queries.CreateUser(ctx, arg)
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	return r.queries.GetUserByID(ctx, id)
}
