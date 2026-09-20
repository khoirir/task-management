package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/khoirir/task-management/internal/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResponse), args.Error(1)
}

func (m *MockService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) ==nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResponse), args.Error(1)
}

func setupTestRouter(service Service) (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	handler := NewHandler(service)

	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	return router, handler
}

func performRequest(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		switch b := body.(type) {
		case []byte:
			bodyReader = bytes.NewBuffer(b)
		default:
			jsonBytes, _ := json.Marshal(body)
			bodyReader = bytes.NewBuffer(jsonBytes)
		}
	}

	req, _ := http.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func TestHandler_Register_Success(t *testing.T) {
	mockService := new(MockService)
	router, _ := setupTestRouter(mockService)

	reqBody := RegisterRequest{
		Name: "Test User",
		Email: "test@test.com",
		Password: "password123",
	}

	expectedResponse := &AuthResponse{
		Token: "mock_jwt_token",
		User: UserResponse{
			ID:    "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
			Name:  reqBody.Name,
			Email: reqBody.Email,
		},
	}

	mockService.On("Register", mock.Anything, reqBody).
		Return(expectedResponse, nil)

	w := performRequest(router, http.MethodPost, "/api/v1/auth/register", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "User registered successfully", res.Message)
	assert.NotNil(t, res.Data)
	mockService.AssertExpectations(t)
}

func TestHandler_Register_InvalidJSON(t *testing.T) {
	mockService := new(MockService)
	router, _ := setupTestRouter(mockService)

	invalidJSON := []byte(`{"name": "Test", "email": }`)

	w := performRequest(router, http.MethodPost, "/api/v1/auth/register", invalidJSON)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Validation error", res.Message)
	mockService.AssertNotCalled(t, "Register")
}

func TestHandler_Register_ValidationError(t *testing.T) {
	mockService := new(MockService)
	router, _ := setupTestRouter(mockService)

	reqBody := RegisterRequest{
		Name: "Test User",
		Email: "test@test.com",
		Password: "123",
	}

	w := performRequest(router, http.MethodPost, "/api/v1/auth/register", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Validation error", res.Message)

	errData, ok := res.Error.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "password must be at least 6 characters", errData["password"])
	mockService.AssertNotCalled(t, "Register")
}

func TestHandler_Register_ServiceError(t *testing.T) {
	mockService := new(MockService)
	router, _ := setupTestRouter(mockService)

	reqBody := RegisterRequest{
		Name: "Test User",
		Email: "test@test.com",
		Password: "password123",
	}

	mockService.On("Register", mock.Anything, reqBody).
		Return(nil, errors.New("Email already exists"))

	w := performRequest(router, http.MethodPost, "/api/v1/auth/register", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Email already exists", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_Login_Success(t *testing.T) {
	mockService := new(MockService)
	router, _ := setupTestRouter(mockService)

	reqBody := LoginRequest{
		Email: "test@test.com",
		Password: "password123",
	}

	expectedResponse := &AuthResponse{
		Token: "mock_jwt_token",
		User: UserResponse{
			ID: "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
			Name: "Test user",
			Email: reqBody.Email,
		},
	}

	mockService.On("Login", mock.Anything, reqBody).
		Return(expectedResponse, nil)

	w := performRequest(router, http.MethodPost, "/api/v1/auth/login", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "User logged in successfully", res.Message)
	assert.NotNil(t, res.Data)
	mockService.AssertExpectations(t)
}

func TestHandler_Login_InvalidJSON(t *testing.T) {
	mockService := new(MockService)
	router, _ := setupTestRouter(mockService)

	invalidJSON := []byte(`{"email": "test@test.com", "password": }`)

	w := performRequest(router, http.MethodPost, "/api/v1/auth/login", invalidJSON)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Validation error", res.Message)
	mockService.AssertNotCalled(t, "Login")
}

func TestHandler_Login_ValidationError(t *testing.T) {
	mockService := new(MockService)
	router, _ := setupTestRouter(mockService)

	reqBody := LoginRequest{
		Email: "not-valid-email",
		Password: "",
	}

	w := performRequest(router, http.MethodPost, "/api/v1/auth/login", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Validation error", res.Message)
	mockService.AssertNotCalled(t, "Login")
}

func TestHandler_Login_ServiceError(t *testing.T) {
	mockService := new(MockService)
	router, _ := setupTestRouter(mockService)

	reqBody := LoginRequest{
		Email: "test@test.com",
		Password: "wrongpassword",
	}

	mockService.On("Login", mock.Anything, reqBody).
		Return(nil, errors.New("Email or password is wrong"))

	w := performRequest(router, http.MethodPost, "/api/v1/auth/login", reqBody)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Email or password is wrong", res.Message)
	mockService.AssertExpectations(t)
}