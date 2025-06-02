package tenant_middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//Later retrieve from token JWT
		tenantId := c.GetHeader("x-tenant-id")
		if tenantId == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenant ID"})
			c.Abort()
			return
		}
		c.Set("tenantId", tenantId)
		c.Next()
	}
}
