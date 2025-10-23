package domain

import "errors"

var (
	// Authentication and authorization errors
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidToken    = errors.New("invalid token")
	ErrMissingToken    = errors.New("missing authentication token")
	ErrInvalidTenantID = errors.New("invalid tenant ID")
	ErrMissingTenantID = errors.New("missing tenant ID")

	// Validation errors
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidRequest   = errors.New("invalid request")

	// Resource errors
	ErrPipelineNotFound = errors.New("pipeline not found")
	ErrDuplicatePipeline = errors.New("pipeline with this name already exists")
	ErrJobNotFound       = errors.New("job not found")

	// Database errors
	ErrDatabaseOperation = errors.New("database operation failed")

	// Rate limiting
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
