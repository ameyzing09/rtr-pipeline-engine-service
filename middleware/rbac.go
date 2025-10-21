package middleware

import (
	"net/http"

	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/gin-gonic/gin"
)

// RequireRoles creates middleware that enforces role-based access control
// Only users with specified roles are allowed to proceed
func RequireRoles(allowedRoles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request context
		requestCtx, exists := GetRequestContext(c)
		if !exists {
			utils.Debug("[RequireRoles] Request context not found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Check if user has one of the allowed roles
		if !requestCtx.HasRole(allowedRoles...) {
			utils.Warn("[RequireRoles] Insufficient permissions: user=%s, role=%s, required=%v",
				requestCtx.UserID, requestCtx.Role, allowedRoles)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin creates middleware that restricts access to ADMIN role only
func RequireAdmin() gin.HandlerFunc {
	return RequireRoles(models.RoleAdmin)
}

// RequireAdminOrHR creates middleware that restricts access to ADMIN or HR roles
func RequireAdminOrHR() gin.HandlerFunc {
	return RequireRoles(models.RoleAdmin, models.RoleHR)
}

// RequireAdminOrHROrInterviewer creates middleware that allows access to all roles
func RequireAdminOrHROrInterviewer() gin.HandlerFunc {
	return RequireRoles(models.RoleAdmin, models.RoleHR, models.RoleInterviewer)
}

// AllowReadOnly creates middleware that allows read-only operations for INTERVIEWER role
// and full access for ADMIN/HR roles
func AllowReadOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request context
		requestCtx, exists := GetRequestContext(c)
		if !exists {
			utils.Debug("[AllowReadOnly] Request context not found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Check method type for INTERVIEWER
		if requestCtx.IsInterviewer() {
			// Only allow GET requests for INTERVIEWER
			if c.Request.Method != "GET" {
				utils.Warn("[AllowReadOnly] INTERVIEWER attempted non-GET: %s %s",
					c.Request.Method, c.Request.URL.Path)
				c.JSON(http.StatusForbidden, gin.H{
					"error": "INTERVIEWER role has read-only access",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
