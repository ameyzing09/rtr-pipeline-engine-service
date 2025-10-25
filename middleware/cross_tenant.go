package middleware

import (
	"net/http"

	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/gin-gonic/gin"
)

const (
	HeaderTenantID    = "X-Tenant-Id"
	CtxTargetTenantID = "target_tenant_id"
)

// CrossTenantMiddleware extracts the effective tenant ID for the operation
// For SUPERADMIN: can use any value in X-Tenant-Id (cross-tenant operation)
// For ADMIN/HR: must use their own tenant (validated by TenantMiddleware)
func CrossTenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request context
		requestCtx, exists := GetRequestContext(c)
		if !exists {
			utils.Debug("[CrossTenantMiddleware] Request context not found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get tenant ID from header (already validated by TenantMiddleware for format)
		targetTenantID := c.GetHeader(HeaderTenantID)
		if targetTenantID == "" {
			// Try alternate case
			targetTenantID = c.GetHeader("X-Tenant-ID")
		}

		if targetTenantID == "" {
			utils.Debug("[CrossTenantMiddleware] Missing X-Tenant-Id header")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Tenant-Id header is required",
				"code":  "MISSING_TENANT_ID",
			})
			c.Abort()
			return
		}

		// Set the target tenant ID in context for handlers to use
		c.Set(CtxTargetTenantID, targetTenantID)

		// Log cross-tenant operations
		if requestCtx.IsSuperAdmin() && targetTenantID != requestCtx.TenantID {
			utils.Debug("[CrossTenantMiddleware] SUPERADMIN cross-tenant operation: user=%s, jwt_tenant=%s, target_tenant=%s",
				requestCtx.UserID, requestCtx.TenantID, targetTenantID)
		} else {
			utils.Debug("[CrossTenantMiddleware] Self-tenant operation: user=%s, tenant=%s",
				requestCtx.UserID, targetTenantID)
		}

		c.Next()
	}
}

// GetTargetTenantID retrieves the effective tenant ID from context
// This is the value from X-Tenant-Id header
func GetTargetTenantID(c *gin.Context) (string, bool) {
	tenantID, exists := c.Get(CtxTargetTenantID)
	if !exists {
		return "", false
	}
	targetID, ok := tenantID.(string)
	return targetID, ok
}
