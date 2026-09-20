package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/khoirir/task-management/internal/auth"
	"github.com/khoirir/task-management/internal/response"
	"github.com/stretchr/testify/assert"
)

const testSecret = "test_jwt_secret_key_portofolio_12345"

func generateTestToken(secret string, userID uuid.UUID, email string, duration time.Duration) string {
	claims := auth.JWTClaims{
		UserID: userID,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func setupTestRouter(jwtSecret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/protected", AuthMiddleware(jwtSecret), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"email": email,
		})
	})

	return router
}

func TestAuthMiddleware_Success(t *testing.T) {
	router := setupTestRouter(testSecret)

	expectedUserID := uuid.New()
	expectedEmail := "test@test.com"
	token := generateTestToken(testSecret, expectedUserID, expectedEmail, 1*time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer " +token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, expectedUserID.String(), res["user_id"])
	assert.Equal(t, expectedEmail, res["email"])
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router := setupTestRouter(testSecret)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Authorization header is required", res.Message)
}

func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	router := setupTestRouter(testSecret)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic some_token_here")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Invalid authorization header format", res.Message)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	router := setupTestRouter(testSecret)

	expiredToken := generateTestToken(testSecret, uuid.New(), "expired@test.com", -1*time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Invalid or expired token", res.Message)
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	router := setupTestRouter(testSecret)

	fakeToken := generateTestToken("another_wrong_secret_12345", uuid.New(), "fake@test.com", 1*time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+fakeToken)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Invalid or expired token", res.Message)
}