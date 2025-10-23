package httpx

import (
	"net/http"

	"github.com/ameyzing09/rtr-pipeline-engine-service/domain"
	"github.com/gin-gonic/gin"
)

// HandleError maps domain errors to appropriate HTTP responses with standardized format
func HandleError(c *gin.Context, err error) {
	switch err {
	case domain.ErrUnauthorized:
		RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized access", err.Error())
	case domain.ErrInvalidToken:
		RespondWithError(c, http.StatusUnauthorized, "INVALID_TOKEN", "Failed to parse or validate token", err.Error())
	case domain.ErrMissingToken:
		RespondWithError(c, http.StatusUnauthorized, "MISSING_TOKEN", "Missing authorization token", "Authorization header not found or malformed")
	case domain.ErrForbidden:
		RespondWithError(c, http.StatusForbidden, "FORBIDDEN", "Access denied", err.Error())
	case domain.ErrInvalidTenantID:
		RespondWithError(c, http.StatusForbidden, "INVALID_TENANT_ID", "Access denied: Tenant ID in JWT does not match X-Tenant-Id header", "")
	case domain.ErrMissingTenantID:
		RespondWithError(c, http.StatusBadRequest, "MISSING_TENANT_ID", "Missing tenant ID header", "x-tenant-id header is required")
	case domain.ErrPipelineNotFound:
		RespondWithError(c, http.StatusNotFound, "PIPELINE_NOT_FOUND", "Pipeline not found", "The requested pipeline does not exist or you don't have access to it")
	case domain.ErrDuplicatePipeline:
		RespondWithError(c, http.StatusConflict, "DUPLICATE_PIPELINE", "Pipeline with this name already exists", "A pipeline with this name already exists for your tenant. Please use a different name.")
	case domain.ErrJobNotFound:
		RespondWithError(c, http.StatusNotFound, "JOB_NOT_FOUND", "Job not found", "The requested job does not exist or you don't have access to it")
	case domain.ErrValidationFailed:
		RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err.Error())
	case domain.ErrInvalidRequest:
		RespondWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", err.Error())
	case domain.ErrRateLimitExceeded:
		RespondWithError(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded", "Too many requests, please try again later")
	case domain.ErrDatabaseOperation:
		RespondWithError(c, http.StatusInternalServerError, "DATABASE_ERROR", "An internal error occurred", "")
	default:
		RespondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred", err.Error())
	}
}
