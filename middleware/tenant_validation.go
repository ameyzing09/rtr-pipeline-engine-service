package middleware

import (
	"net/http"

	"github.com/ameyzing09/rtr-pipeline-engine-service/cache"
	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/gin-gonic/gin"
)

// RequireTenantActive validates that the target tenant exists and is in ACTIVE state
// This middleware should be used for cross-tenant operations where we need to verify
// the tenant before performing operations on it.
//
// TODO: Implement actual tenant validation via one of:
//   1. HTTP call to auth-service GET /tenant/:id
//   2. Local cached/replicated tenants table
//   3. Shared tenant service/cache
func RequireTenantActive() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestCtx, exists := GetRequestContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// Get the target tenant ID
		targetTenantID, exists := GetTargetTenantID(c)
		if !exists || targetTenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Target tenant ID not found",
				"code":  "MISSING_TENANT_ID",
			})
			c.Abort()
			return
		}

		// Self-seed optimization: If user is operating on their own tenant,
		// we can skip validation because JWT issuance already proves tenant is active
		if targetTenantID == requestCtx.TenantID {
			utils.Debug("[TenantValidation] Self-tenant operation, skipping validation: tenant=%s", targetTenantID)
			c.Next()
			return
		}

		// Cross-tenant operation (SUPERADMIN): Must validate target tenant
		utils.Debug("[TenantValidation] Cross-tenant operation, validating tenant: %s", targetTenantID)

		// TODO: Implement actual validation
		// For now, we'll pass through and rely on FK constraints at DB level
		// When implemented, this should:
		//   1. Check if tenant exists
		//   2. Verify tenant status == ACTIVE
		//   3. Return 404 if not found
		//   4. Return 409 if not ACTIVE (PENDING/SUSPENDED/DELETED/FAILED)
		//   5. Return 503 if auth-service unavailable

		status, err := validateTenantStatusWithCache(targetTenantID)
		if err != nil {
			utils.Error("[TenantValidation] Error validating tenant %s: %v", targetTenantID, err)
			// For now, log and continue (fail-open during transition period)
			// TODO: Change to fail-closed once validation is implemented
			c.Next()
			return
		}

		if status != cache.TenantStatusActive {
			utils.Warn("[TenantValidation] Tenant %s is not active: status=%s", targetTenantID, status)
			c.JSON(http.StatusConflict, gin.H{
				"error":  "Tenant is not in ACTIVE state",
				"code":   "TENANT_NOT_ACTIVE",
				"status": status,
			})
			c.Abort()
			return
		}

		utils.Debug("[TenantValidation] Tenant %s is active", targetTenantID)
		c.Next()
	}
}

// validateTenantStatusWithCache validates the tenant status with caching (60s TTL)
func validateTenantStatusWithCache(tenantID string) (cache.TenantStatus, error) {
	// Try cache first
	tenantCache := cache.GetGlobalCache()
	if status, found := tenantCache.Get(tenantID); found {
		utils.Debug("[TenantValidation] Cache hit for tenant %s: status=%s", tenantID, status)
		return status, nil
	}

	// Cache miss - fetch from source
	utils.Debug("[TenantValidation] Cache miss for tenant %s, fetching status", tenantID)
	status, err := fetchTenantStatusFromSource(tenantID)
	if err != nil {
		return "", err
	}

	// Cache the result
	tenantCache.Set(tenantID, status)
	utils.Debug("[TenantValidation] Cached tenant %s status: %s", tenantID, status)

	return status, nil
}

// fetchTenantStatusFromSource fetches tenant status from the authoritative source
// TODO: Replace with actual implementation
func fetchTenantStatusFromSource(tenantID string) (cache.TenantStatus, error) {
	// Placeholder implementation
	// In production, this should either:
	//   1. Call auth-service: GET /internal/tenants/:id
	//   2. Query local replicated tenants table
	//   3. Query database with tenant info

	// For now, assume all tenants are active (rely on DB FK constraints)
	return cache.TenantStatusActive, nil

	// Production implementation example (HTTP call to auth-service):
	// resp, err := http.Get(fmt.Sprintf("%s/internal/tenants/%s", authServiceURL, tenantID))
	// if err != nil {
	//     return "", fmt.Errorf("failed to reach auth-service: %w", err)
	// }
	// defer resp.Body.Close()
	//
	// if resp.StatusCode == 404 {
	//     return cache.TenantStatusDeleted, nil
	// }
	//
	// var tenant struct {
	//     Status string `json:"status"`
	// }
	// if err := json.NewDecoder(resp.Body).Decode(&tenant); err != nil {
	//     return "", err
	// }
	//
	// return cache.TenantStatus(tenant.Status), nil
}
