package middleware

import (
	"net/http"
	"strings"

	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TenantMiddleware validates the X-Tenant-Id header
// and cross-checks it with the tenant ID from JWT claims
// Returns 403 if tenant IDs don't match (Layer 1: Security Validation)
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant ID from header (normalize case)
		headerTenantID := c.GetHeader("X-Tenant-Id")
		if headerTenantID == "" {
			// Try alternate case
			headerTenantID = c.GetHeader("X-Tenant-ID")
		}

		// Trim whitespace
		headerTenantID = strings.TrimSpace(headerTenantID)

		if headerTenantID == "" {
			utils.Debug("[TenantMiddleware] Missing X-Tenant-Id header")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Tenant-Id header is required",
				"code":  "MISSING_TENANT_ID",
			})
			c.Abort()
			return
		}

		// Validate UUID format
		if _, err := uuid.Parse(headerTenantID); err != nil {
			utils.Warn("[TenantMiddleware] Invalid tenant ID format: %s", headerTenantID)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Tenant-Id must be a valid UUID",
				"code":  "INVALID_TENANT_ID_FORMAT",
			})
			c.Abort()
			return
		}

		// Get request context (set by JWT middleware)
		requestCtx, exists := GetRequestContext(c)
		if exists {
			// SUPERADMIN can operate on any tenant - skip the cross-check
			if !requestCtx.IsSuperAdmin() {
				// Cross-check: JWT.tid === X-Tenant-Id header (for non-SUPERADMIN users)
				if headerTenantID != requestCtx.TenantID {
					utils.Warn("[TenantMiddleware] Tenant mismatch: header=%s, jwt=%s, role=%s",
						headerTenantID, requestCtx.TenantID, requestCtx.Role)
					c.JSON(http.StatusForbidden, gin.H{
						"error": "Tenant ID mismatch - non-SUPERADMIN users can only access their own tenant",
					})
					c.Abort()
					return
				}
			} else {
				// SUPERADMIN cross-tenant operation
				utils.Debug("[TenantMiddleware] SUPERADMIN cross-tenant operation: jwt_tenant=%s, target_tenant=%s",
					requestCtx.TenantID, headerTenantID)
			}
		}

		// Set tenant ID in context for backward compatibility
		c.Set(CtxTenantID, headerTenantID)

		utils.Debug("[TenantMiddleware] Tenant validated: %s", headerTenantID)
		c.Next()
	}
}
