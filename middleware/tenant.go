package middleware

import (
	"net/http"

	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/gin-gonic/gin"
)

// TenantMiddleware validates the X-Tenant-Id header
// and cross-checks it with the tenant ID from JWT claims
// Returns 403 if tenant IDs don't match (Layer 1: Security Validation)
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant ID from header
		headerTenantID := c.GetHeader("X-Tenant-ID")
		if headerTenantID == "" {
			utils.Debug("[TenantMiddleware] Missing X-Tenant-ID header")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing tenant ID header (X-Tenant-ID)",
			})
			c.Abort()
			return
		}

		// Get request context (set by JWT middleware)
		requestCtx, exists := GetRequestContext(c)
		if exists {
			// Cross-check: JWT.tid === X-Tenant-Id header
			if headerTenantID != requestCtx.TenantID {
				utils.Warn("[TenantMiddleware] Tenant mismatch: header=%s, jwt=%s",
					headerTenantID, requestCtx.TenantID)
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Tenant ID mismatch",
				})
				c.Abort()
				return
			}
		}

		// Set tenant ID in context for backward compatibility
		c.Set(CtxTenantID, headerTenantID)

		utils.Debug("[TenantMiddleware] Tenant validated: %s", headerTenantID)
		c.Next()
	}
}
