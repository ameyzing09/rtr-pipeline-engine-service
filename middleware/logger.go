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
		role := "unknown"
		if exists {
			tenantID = requestCtx.TenantID
			userID = requestCtx.UserID
			role = string(requestCtx.Role)
		}

		// Get request ID if available
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = "-"
		}

		// Get target tenant for cross-tenant operations
		targetTenant, hasTenantTarget := GetTargetTenantID(c)
		if !hasTenantTarget {
			targetTenant = tenantID
		}

		// Log request details with comprehensive audit information
		utils.Info(
			"[HTTP] request_id=%s method=%s path=%s status=%d duration=%s "+
				"tenant_id=%s user_id=%s role=%s target_tenant=%s ip=%s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			wrapped.statusCode,
			duration,
			tenantID,
			userID,
			role,
			targetTenant,
			c.ClientIP(),
		)

		// Log errors separately for easier monitoring
		if wrapped.statusCode >= 400 {
			utils.Warn(
				"[HTTP_ERROR] request_id=%s status=%d path=%s user=%s tenant=%s",
				requestID,
				wrapped.statusCode,
				c.Request.URL.Path,
				userID,
				tenantID,
			)
		}
	}
}
