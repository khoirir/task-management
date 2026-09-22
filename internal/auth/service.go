package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/khoirirrosikin/task-management/internal/response"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
}

type service struct {
	repo Repository
	jwtSecret []byte
}

func NewService(repo Repository, jwtSecret string) Service {
	return &service{
		repo: repo,
		jwtSecret: []byte(jwtSecret),
	}
}

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, response.ErrBadRequest("Email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Failed to hash password: %v", err)
	}

	user, err := s.repo.CreateUser(ctx, db.CreateUserParams{
		Name: req.Name,
		Email: req.Email,
		PasswordHash: string(hashedPassword),
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to create user: %w", err)
	}

	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User: UserResponse{
			ID: user.ID.String(),
			Name: user.Name,
			Email: user.Email,
			CreatedAt: user.CreatedAt.Time,
		},
	}, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, response.ErrUnauthorized("Email or password is wrong")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, response.ErrUnauthorized("Email or password is wrong")
	}

	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User: UserResponse{
			ID: user.ID.String(),
			Name: user.Name,
			Email: user.Email,
			CreatedAt: user.CreatedAt.Time,
		},
	}, nil
}

func (s *service) generateToken(userID uuid.UUID, email string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}