package middleware

import (
	"time"

	"github.com/ameyzing09/rtr-pipeline-engine-service/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a configured CORS middleware
// Layer 2: CORS - Cross-Origin Resource Sharing configuration
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	config := cors.Config{
		// Allow specific origins from configuration
		AllowOrigins: cfg.CORS.AllowedOrigins,

		// Allow all common HTTP methods
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},

		// Allow required headers for authentication and API functionality
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",    // JWT token
			"X-Tenant-ID",      // Tenant identification
			"X-Requested-With", // AJAX requests
			"X-Request-ID",     // Request tracing
		},

		// Expose headers that frontend can read
		ExposeHeaders: []string{
			"Content-Length",
			"X-Request-ID",
		},

		// Allow credentials (cookies, authorization headers)
		// Required for JWT token authentication
		AllowCredentials: true,

		// Cache preflight requests for specified duration
		MaxAge: time.Duration(cfg.CORS.MaxAgeHours) * time.Hour,
	}

	return cors.New(config)
}
