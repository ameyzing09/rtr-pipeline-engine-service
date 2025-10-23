package middleware

import (
	"time"

	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/gin-gonic/gin"
)

// responseWriter is a wrapper around gin.ResponseWriter to capture status code
type responseWriter struct {
	gin.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggerMiddleware logs HTTP requests with tenant_id, user_id, route, and status
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: c.Writer,
			statusCode:     200, // Default status code
		}
		c.Writer = wrapped

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Extract request context (contains tenant_id and user_id from JWT)
		requestCtx, exists := GetRequestContext(c)

		tenantID := "unknown"
		userID := "unknown"
		if exists {
			tenantID = requestCtx.TenantID
			userID = requestCtx.UserID
		}

		// Log request details
		utils.Info(
			"[HTTP] method=%s path=%s status=%d duration=%s tenant_id=%s user_id=%s ip=%s",
			c.Request.Method,
			c.Request.URL.Path,
			wrapped.statusCode,
			duration,
			tenantID,
			userID,
			c.ClientIP(),
		)
	}
}
