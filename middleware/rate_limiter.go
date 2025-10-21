package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/ameyzing09/rtr-pipeline-engine-service/config"
	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimiterMiddleware creates a rate limiting middleware
// Limits by IP address and Tenant ID combination
// Returns 429 Too Many Requests when limit exceeded
func RateLimiterMiddleware(cfg *config.Config) gin.HandlerFunc {
	// Create in-memory store
	store := memory.NewStore()

	// Create limiter with configured rate limit
	limiterInstance := limiter.New(
		store,
		limiter.Rate{
			Limit:  int64(cfg.RateLimit.RequestsPerMinute),
			Period: 60, // 1 minute in seconds
		},
	)

	return func(c *gin.Context) {
		// Get client IP
		clientIP := getClientIP(c)

		// Get tenant ID from context (may not be set yet)
		tenantID := c.GetString(CtxTenantID)

		// Create unique key: "ip:tenant" for rate limiting
		key := fmt.Sprintf("%s:%s", clientIP, tenantID)

		// Check rate limit
		context, err := limiterInstance.Get(c.Request.Context(), key)
		if err != nil {
			// On error, allow request to proceed (fail open)
			utils.Warn("[RateLimiter] Error checking rate limit: %v", err)
			c.Next()
			return
		}

		// Check if limit exceeded
		if context.Reached {
			utils.Warn("[RateLimiter] Rate limit exceeded for %s", key)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		c.Next()
	}
}

// getClientIP extracts the real client IP from request
// Checks X-Forwarded-For, X-Real-IP, and RemoteAddr
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (for proxied requests)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple are present
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}

	return ip
}
