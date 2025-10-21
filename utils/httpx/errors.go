package httpx

import (
	"net/http"

	"github.com/ameyzing09/rtr-pipeline-engine-service/domain"
	"github.com/gin-gonic/gin"
)

// HandleError maps domain errors to appropriate HTTP responses
func HandleError(c *gin.Context, err error) {
	switch err {
	case domain.ErrUnauthorized, domain.ErrInvalidToken, domain.ErrMissingToken:
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case domain.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case domain.ErrPipelineNotFound, domain.ErrJobNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case domain.ErrValidationFailed, domain.ErrInvalidRequest, domain.ErrInvalidTenantID, domain.ErrMissingTenantID:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case domain.ErrRateLimitExceeded:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
	case domain.ErrDatabaseOperation:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal error occurred"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
