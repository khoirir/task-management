package response

import (
	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data any `json:"data,omitempty"`
	Error any `json:"error,omitempty"`
}

func Success(c *gin.Context, statusCode int, message string, data any){
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data: data,
	})	
}

func Error(c *gin.Context, statusCode int, message string, err any){
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Error: err,
	})
}