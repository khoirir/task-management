package project

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func setupTestRepository(t *testing.T) (pgxmock.PgxPoolIface, Repository) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)

	t.Cleanup(func(){
		mock.Close()
	})

	queries := db.New(mock)
	repo := NewRepository(queries)
	
	return mock, repo
}

var projectColumns = []string{"id", "name", "description", "owner_id", "created_at", "updated_at"}

func TestRepository_CreateProject_Success(t *testing.T) {
	mock, repo := setupTestRepository(t)

	ownerID := uuid.New()
	projectID := uuid.New()
	now := time.Now()

	arg := db.CreateProjectParams{
		Name: "Test project",
		Description: pgtype.Text{String: "Test project description", Valid: true},
		OwnerID: ownerID,
	}

	expectedCreatedAt := pgtype.Timestamptz{Time: now, Valid: true}
	expectedUpdatedAt := pgtype.Timestamptz{Time: now, Valid: true}
	
	mock.ExpectQuery("INSERT INTO projects").
		WithArgs(arg.Name, arg.Description, arg.OwnerID).
		WillReturnRows(
			pgxmock.NewRows(projectColumns).
				AddRow(projectID, arg.Name, arg.Description, arg.OwnerID, expectedCreatedAt, expectedUpdatedAt),
		)
	
	project, err := repo.CreateProject(context.Background(), arg)

	assert.NoError(t, err)
	assert.Equal(t, projectID, project.ID)
	assert.Equal(t, arg.Name, project.Name)
	assert.Equal(t, arg.Description.String, project.Description.String)
	assert.Equal(t, ownerID, project.OwnerID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CreateProject_Error(t *testing.T) {
	mock, repo := setupTestRepository(t)

	arg := db.CreateProjectParams{
		Name: "Test project",
		OwnerID: uuid.New(),
	}
	
	expectedErr := errors.New("database error")

	mock.ExpectQuery("INSERT INTO projects").
		WithArgs(arg.Name, arg.Description, arg.OwnerID).
		WillReturnError(expectedErr)
	
	project, err := repo.CreateProject(context.Background(), arg)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, db.Project{}, project)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetProjectByID_Success(t *testing.T) {
	mock, repo := setupTestRepository(t)

	ownerID := uuid.New()
	projectID := uuid.New()
	now := time.Now()
	desc := pgtype.Text{String: "Test project description", Valid: true}
	createdAt := pgtype.Timestamptz{Time: now, Valid: true}
	updatedAt := pgtype.Timestamptz{Time: now, Valid: true}
	
	mock.ExpectQuery("SELECT .* FROM projects WHERE id =").
		WithArgs(projectID).
		WillReturnRows(
			pgxmock.NewRows(projectColumns).
				AddRow(projectID, "Test project", desc, ownerID, createdAt, updatedAt),
		)
	
	project, err := repo.GetProjectByID(context.Background(), projectID)

	assert.NoError(t, err)
	assert.Equal(t, projectID, project.ID)
	assert.Equal(t, "Test project", project.Name)
	assert.Equal(t, "Test project description", project.Description.String)
	assert.Equal(t, ownerID, project.OwnerID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetProjectByID_NotFound(t *testing.T) {
	mock, repo := setupTestRepository(t)

	projectID := uuid.New()

	mock.ExpectQuery("SELECT .* FROM projects WHERE id =").
		WithArgs(projectID).
		WillReturnError(pgx.ErrNoRows)

	project, err := repo.GetProjectByID(context.Background(), projectID)

	assert.Error(t, err)
	assert.Equal(t, pgx.ErrNoRows, err)
	assert.Equal(t, db.Project{}, project)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestRepository_ListProjectsByOwner_Success(t *testing.T) {
	mock, repo := setupTestRepository(t)

	ownerID := uuid.New()
	nowTz := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	mockRows := pgxmock.NewRows(projectColumns)
	for i := 1; i <= 2; i++ {
		mockRows.AddRow(
			uuid.New(),
			fmt.Sprintf("Project %d", i),
			pgtype.Text{Valid: false},
			ownerID,
			nowTz,
			nowTz,
		)
	}
	
	mock.ExpectQuery("SELECT .* FROM projects WHERE owner_id =").
		WithArgs(ownerID).
		WillReturnRows(mockRows)

	projects, err := repo.ListProjectsByOwner(context.Background(), ownerID)

	assert.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Equal(t, "Project 1", projects[0].Name)
	assert.Equal(t, "Project 2", projects[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestRepository_UpdateProject_Success(t *testing.T) {
	mock, repo := setupTestRepository(t)

	projectID := uuid.New()
	ownerID := uuid.New()
	now := time.Now()

	arg := db.UpdateProjectParams{
		ID:          projectID,
		Name:        "Updated Name",
		Description: pgtype.Text{String: "Updated Desc", Valid: true},
		OwnerID:     ownerID,
	}
	
	mock.ExpectQuery("UPDATE projects").
		WithArgs(arg.ID, arg.Name, arg.Description, arg.OwnerID).
		WillReturnRows(
			pgxmock.NewRows(projectColumns).
				AddRow(arg.ID, arg.Name, arg.Description, arg.OwnerID, pgtype.Timestamptz{Time: now, Valid: true}, pgtype.Timestamptz{Time: now, Valid: true}),
		)

	project, err := repo.UpdateProject(context.Background(), arg)

	assert.NoError(t, err)
	assert.Equal(t, arg.ID, project.ID)
	assert.Equal(t, arg.Name, project.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestRepository_DeleteProject_Success(t *testing.T) {
	mock, repo := setupTestRepository(t)

	projectID := uuid.New()
	ownerID := uuid.New()

	arg := db.DeleteProjectParams{
		ID:      projectID,
		OwnerID: ownerID,
	}

	mock.ExpectExec("DELETE FROM projects").
		WithArgs(arg.ID, arg.OwnerID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.DeleteProject(context.Background(), arg)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}