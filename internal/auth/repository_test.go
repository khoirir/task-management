package auth

import (
	"context"
	"errors"
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

	t.Cleanup(func ()  {
		mock.Close()
	})

	queries := db.New(mock)
	repo := NewRepository(queries)
	return mock, repo
}

func TestRepository_CreateUser_Success(t *testing.T) {
	mock, repo := setupTestRepository(t)

	arg := db.CreateUserParams{
		Name:         "Test User",
		Email:        "test@test.com",
		PasswordHash: "password123hashed",
	}

	expectedID := uuid.New()
	now := time.Now()
	expectedCreatedAt := pgtype.Timestamptz{Time: now, Valid: true}
	expectedUpdatedAt := pgtype.Timestamptz{Time: now, Valid: true}

	columns := []string{"id", "name", "email", "password_hash", "created_at", "updated_at"}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(arg.Name, arg.Email, arg.PasswordHash).
		WillReturnRows(
			pgxmock.NewRows(columns).
				AddRow(expectedID, arg.Name, arg.Email, arg.PasswordHash, expectedCreatedAt, expectedUpdatedAt),
		)

	user, err := repo.CreateUser(context.Background(), arg)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, user.ID)
	assert.Equal(t, arg.Name, user.Name)
	assert.Equal(t, arg.Email, user.Email)
	assert.Equal(t, arg.PasswordHash, user.PasswordHash)
	assert.True(t, user.CreatedAt.Valid)
	assert.True(t, user.UpdatedAt.Valid)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CreateUser_Error(t *testing.T) {
	mock, repo := setupTestRepository(t)

	arg := db.CreateUserParams{
		Name: "Test User",
		Email: "test@test.com",
		PasswordHash: "password123hashed",
	}

	expectedErr := errors.New("Database connection error")

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(arg.Name, arg.Email, arg.PasswordHash).
		WillReturnError(expectedErr)

	user, err := repo.CreateUser(context.Background(), arg)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, db.User{}, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetUserByEmail_Success(t *testing.T) {
	mock, repo := setupTestRepository(t)
	
	email := "test@test.com"
	expectedID := uuid.New()
	now := time.Now()
	expectedCreatedAt := pgtype.Timestamptz{Time: now, Valid: true}
	expectedUpdatedAt := pgtype.Timestamptz{Time: now, Valid: true}

	columns := []string{"id","name","email","password_hash","created_at","updated_at"}

	mock.ExpectQuery("SELECT .* FROM users WHERE email =").
		WithArgs(email).
		WillReturnRows(
			pgxmock.NewRows(columns).
				AddRow(expectedID, "Test User", email, "password123hashed", expectedCreatedAt, expectedUpdatedAt),
		)

	user, err := repo.GetUserByEmail(context.Background(), email)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, user.ID)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, "password123hashed", user.PasswordHash)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetUserByEmail_NotFound(t *testing.T) {
	mock, repo := setupTestRepository(t)

	email := "notfound@test.com"

	mock.ExpectQuery("SELECT .* FROM users WHERE email =").
		WithArgs(email).
		WillReturnError(pgx.ErrNoRows)

	user, err := repo.GetUserByEmail(context.Background(), email)

	assert.Error(t, err)
	assert.Equal(t, pgx.ErrNoRows, err)
	assert.Equal(t, db.User{}, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetUserByID_Success(t *testing.T) {
	mock, repo := setupTestRepository(t)

	userID := uuid.New()
	now := time.Now()
	expectedCreatedAt := pgtype.Timestamptz{Time: now, Valid: true}
	expectedUpdatedAt := pgtype.Timestamptz{Time: now, Valid: true}

	columns := []string{"id", "name", "email", "password_hash", "created_at", "updated_at"}

	mock.ExpectQuery("SELECT .* FROM users WHERE id =").
		WithArgs(userID).
		WillReturnRows(
			pgxmock.NewRows(columns).
				AddRow(userID, "Test User", "test@test.com", "password123hashed", expectedCreatedAt, expectedUpdatedAt),
		)

	user, err := repo.GetUserByID(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "test@test.com", user.Email)
	assert.Equal(t, "password123hashed", user.PasswordHash)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetUserByID_NotFound(t *testing.T) {
	mock, repo := setupTestRepository(t)

	userID := uuid.New()

	mock.ExpectQuery("SELECT .* FROM users WHERE id =").
		WithArgs(userID).
		WillReturnError(pgx.ErrNoRows)

	user, err := repo.GetUserByID(context.Background(), userID)

	assert.Error(t, err)
	assert.Equal(t, pgx.ErrNoRows, err)
	assert.Equal(t, db.User{}, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}
