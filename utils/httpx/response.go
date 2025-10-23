package httpx

import (
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Details    string `json:"details,omitempty"`
}

// SuccessResponse represents a standardized success response with data
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, statusCode int, code, message, details string) {
	c.JSON(statusCode, ErrorResponse{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
	})
}

// RespondWithSuccess sends a standardized success response
func RespondWithSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Data: data,
	})
}

// RespondWithMessage sends a simple success message
func RespondWithMessage(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"message": message})
}
