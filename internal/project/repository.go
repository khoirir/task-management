package project

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoirirrosikin/task-management/internal/database/db"
)

type Repository interface {
	CreateProject(ctx context.Context, arg db.CreateProjectParams) (db.Project, error)
	GetProjectByID(ctx context.Context, id uuid.UUID) (db.Project, error)
	ListProjectsByOwner(ctx context.Context, ownerID uuid.UUID) ([]db.Project, error)
	UpdateProject(ctx context.Context, arg db.UpdateProjectParams) (db.Project, error)
	DeleteProject(ctx context.Context, arg db.DeleteProjectParams) error
}

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{
		queries: queries,
	}
}

func (r *repository) CreateProject(ctx context.Context, arg db.CreateProjectParams) (db.Project, error) {
	return r.queries.CreateProject(ctx, arg)
}

func (r *repository) GetProjectByID(ctx context.Context, id uuid.UUID) (db.Project, error) {
	return r.queries.GetProjectByID(ctx, id)
}

func (r *repository) ListProjectsByOwner(ctx context.Context, ownerID uuid.UUID) ([]db.Project, error) {
	return r.queries.ListProjectsByOwner(ctx, ownerID)
}

func (r *repository) UpdateProject(ctx context.Context, arg db.UpdateProjectParams) (db.Project, error) {
	return r.queries.UpdateProject(ctx, arg)
}

func (r *repository) DeleteProject(ctx context.Context, arg db.DeleteProjectParams) error {
	return r.queries.DeleteProject(ctx, arg)
}