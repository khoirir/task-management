package project

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/khoirirrosikin/task-management/internal/response"
)

type Service interface {
	CreateProject(ctx context.Context, ownerID uuid.UUID, req CreateProjectRequest) (*ProjectResponse, error)
	GetProjectByID(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (*ProjectResponse, error)
	ListProjects(ctx context.Context, ownerID uuid.UUID) ([]ProjectResponse, error)
	UpdateProject(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, req UpdateProjectRequest) (*ProjectResponse, error)
	DeleteProject(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateProject(ctx context.Context, ownerID uuid.UUID, req CreateProjectRequest) (*ProjectResponse, error) {
	arg := db.CreateProjectParams{
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid: req.Description != "",
		},
		OwnerID: ownerID,
	}

	project, err := s.repo.CreateProject(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("Failed to create project: %w", err)
	}

	return toProjectResponse(project), nil
}

func (s *service) GetProjectByID(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (*ProjectResponse, error) {
	project, err := s.getProjectAndVerifyOwner(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	return toProjectResponse(project), nil
}

func (s *service) ListProjects(ctx context.Context, ownerID uuid.UUID) ([]ProjectResponse, error) {
	projects, err := s.repo.ListProjectsByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("Failed to list projects: %w", err)
	}

	res := make([]ProjectResponse, 0, len(projects))
	for _, p := range projects {
		res = append(res, *toProjectResponse(p))
	}

	return res, nil
}

func (s *service) UpdateProject(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, req UpdateProjectRequest) (*ProjectResponse, error) {
	if _, err := s.getProjectAndVerifyOwner(ctx, projectID, ownerID); err != nil {
		return nil, err
	}

	arg := db.UpdateProjectParams{
		ID: projectID,
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid: req.Description != "",
		},
		OwnerID: ownerID,
	}

	updated, err := s.repo.UpdateProject(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("Failed to update project")
	}

	return toProjectResponse(updated), nil
}

func (s *service) DeleteProject(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID) error {
	if _, err := s.getProjectAndVerifyOwner(ctx, projectID, ownerID); err != nil {
		return err
	}

	arg := db.DeleteProjectParams{
		ID: projectID,
		OwnerID: ownerID,
	}

	if err := s.repo.DeleteProject(ctx, arg); err != nil {
		return fmt.Errorf("Failed to delete project: %w", err)
	}

	return nil
}

func (s *service) getProjectAndVerifyOwner(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID) (db.Project, error) {
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return db.Project{}, response.ErrNotFound("Project not found")
	}

	if project.OwnerID != ownerID {
		return db.Project{}, response.ErrForbidden("Unauthorized to access this project")
	}

	return project, nil
}

func toProjectResponse(p db.Project) *ProjectResponse {
	var desc string
	if p.Description.Valid {
		desc = p.Description.String
	}

	return &ProjectResponse{
		ID: p.ID.String(),
		Name: p.Name,
		Description: desc,
		OwnerID: p.OwnerID.String(),
		CreatedAt: p.CreatedAt.Time,
		UpdatedAt: p.UpdatedAt.Time,
	}
}