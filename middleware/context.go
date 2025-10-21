package middleware

import (
	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"github.com/gin-gonic/gin"
)

// RequestContext holds the authenticated user and tenant information
// extracted from JWT and request headers
type RequestContext struct {
	UserID   string      `json:"user_id"`
	TenantID string      `json:"tenant_id"`
	Role     models.Role `json:"role"`
	Email    string      `json:"email"`
}

// SetRequestContext stores RequestContext in gin context
func SetRequestContext(c *gin.Context, ctx *RequestContext) {
	c.Set(CtxRequestContext, ctx)
}

// GetRequestContext retrieves RequestContext from gin context
func GetRequestContext(c *gin.Context) (*RequestContext, bool) {
	ctx, exists := c.Get(CtxRequestContext)
	if !exists {
		return nil, false
	}
	requestCtx, ok := ctx.(*RequestContext)
	return requestCtx, ok
}

// HasRole checks if the RequestContext has the specified role(s)
func (rc *RequestContext) HasRole(roles ...models.Role) bool {
	for _, role := range roles {
		if rc.Role == role {
			return true
		}
	}
	return false
}

// IsAdmin returns true if user is ADMIN
func (rc *RequestContext) IsAdmin() bool {
	return rc.Role == models.RoleAdmin
}

// IsHR returns true if user is HR
func (rc *RequestContext) IsHR() bool {
	return rc.Role == models.RoleHR
}

// IsInterviewer returns true if user is INTERVIEWER
func (rc *RequestContext) IsInterviewer() bool {
	return rc.Role == models.RoleInterviewer
}
