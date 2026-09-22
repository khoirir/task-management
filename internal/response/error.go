package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Code int
	Message string
	Err error
}

func (e *AppError) Error() string {
	return e.Message
}

func ErrNotFound(message string) *AppError {
	return &AppError{
		Code: http.StatusNotFound,
		Message: message,
	}
}

func ErrForbidden(message string) *AppError {
	return &AppError{
		Code: http.StatusForbidden,
		Message: message,
	}
}

func ErrUnauthorized(message string) *AppError {
	return &AppError{
		Code: http.StatusUnauthorized,
		Message: message,
	}
}

func ErrBadRequest(message string) *AppError {
	return &AppError{
		Code: http.StatusBadRequest,
		Message: message,
	}
}

func ErrInternal(message string) *AppError {
	return &AppError{
		Code: http.StatusInternalServerError,
		Message: message,
	}
}

func HandleError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		Error(c, appErr.Code, appErr.Message, nil)
		return
	}

	Error(c, http.StatusInternalServerError, "Internal server error", nil)
}