package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ameyzing09/rtr-pipeline-engine-service/config"
	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID   string `json:"uid"`
	TenantID string `json:"tid"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// JWTMiddleware parses JWT token from Authorization header and injects
// tenantId, userId, and role into request context
func JWTMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		tokenString := extractToken(c)
		if tokenString == "" {
			utils.Debug("[JWTMiddleware] Missing authorization token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization token",
			})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := parseToken(tokenString, cfg.JWT.Secret)
		if err != nil {
			utils.Debug("[JWTMiddleware] Failed to parse token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			utils.Debug("[JWTMiddleware] Invalid token claims")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			c.Abort()
			return
		}
		//log the claims for debugging
		utils.Debug("[JWTMiddleware] Token claims: %+v", claims)
		// Validate required fields
		if claims.UserID == "" || claims.TenantID == "" || claims.Role == "" {
			utils.Debug("[JWTMiddleware] Token missing required fields")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token missing required fields",
			})
			c.Abort()
			return
		}

		// Create request context
		requestCtx := &RequestContext{
			UserID:   claims.UserID,
			TenantID: claims.TenantID,
			Role:     models.Role(claims.Role),
			Email:    claims.Email,
		}

		// Set in gin context
		SetRequestContext(c, requestCtx)

		utils.Debug("[JWTMiddleware] Authenticated user: %s (role=%s, tenant=%s)",
			claims.UserID, claims.Role, claims.TenantID)

		// Continue to next handler
		c.Next()
	}
}

// extractToken extracts the Bearer token from Authorization header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// parseToken parses and validates JWT token
func parseToken(tokenString, secret string) (*jwt.Token, error) {
	if secret == "" {
		return nil, fmt.Errorf("JWT secret not configured")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		},
	)

	return token, err
}
