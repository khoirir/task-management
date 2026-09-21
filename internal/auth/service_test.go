package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

const testJWTSecret = "test_secret_key_portofolio_12345"

func TestRegister_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, testJWTSecret)

	req := RegisterRequest{
		Name: "Test User",
		Email: "test@test.com",
		Password: "password123",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).
		Return(db.User{}, errors.New("user not found"))

	mockUser := db.User{
		ID: uuid.New(),
		Name: req.Name,
		Email: req.Email,
		CreatedAt: pgtype.Timestamptz{
			Time: time.Now(),
			Valid: true,
		},
	}

	mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(arg db.CreateUserParams) bool {
		return arg.Name == req.Name && arg.Email == req.Email
	})).Return(mockUser, nil)

	res, err := service.Register(context.Background(), req)
	
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Token)
	assert.Equal(t, req.Email, res.User.Email)
	mockRepo.AssertExpectations(t)
}

func TestRegister_CreateUserError(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, testJWTSecret)

	req := RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).
		Return(db.User{}, errors.New("user not found"))

	mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(arg db.CreateUserParams) bool {
		return arg.Name == req.Name && arg.Email == req.Email
	})).Return(db.User{}, errors.New("database error"))

	res, err := service.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "Failed to create user")
	mockRepo.AssertExpectations(t)
}


func TestRegister_EmailAlreadyExists(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, testJWTSecret)

	req := RegisterRequest{
		Name:     "Test User",
		Email:    "existing@example.com",
		Password: "password123",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).
		Return(db.User{Email: req.Email}, nil)

	res, err := service.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "Email already exists", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, testJWTSecret)

	rawPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)

	mockUser := db.User{
		ID:           uuid.New(),
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	mockRepo.On("GetUserByEmail", mock.Anything, mockUser.Email).
		Return(mockUser, nil)

	req := LoginRequest{
		Email:    mockUser.Email,
		Password: rawPassword,
	}

	res, err := service.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Token)
	assert.Equal(t, mockUser.Email, res.User.Email)
	mockRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, testJWTSecret)

	req := LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).
		Return(db.User{}, errors.New("user not found"))

	res, err := service.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "Email or password is wrong", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, testJWTSecret)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)
	mockUser := db.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	mockRepo.On("GetUserByEmail", mock.Anything, mockUser.Email).
		Return(mockUser, nil)

	req := LoginRequest{
		Email:    mockUser.Email,
		Password: "wrong_password",
	}

	res, err := service.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "Email or password is wrong", err.Error())
	mockRepo.AssertExpectations(t)
}
