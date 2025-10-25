package handlers

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/ameyzing09/rtr-pipeline-engine-service/middleware"
	"github.com/gin-gonic/gin"
)

var (
	// Build information (can be set via ldflags during build)
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"

	// Application start time
	startTime = time.Now()
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health returns basic health status
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// HealthDetailed returns detailed health information with build info and metrics
func (h *HealthHandler) HealthDetailed(c *gin.Context) {
	uptime := time.Since(startTime)

	// Get metrics snapshot
	metrics := middleware.GetMetrics().GetSnapshot()

	// Get environment
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"build": gin.H{
			"version":    Version,
			"build_time": BuildTime,
			"git_commit": GitCommit,
			"go_version": runtime.Version(),
		},
		"runtime": gin.H{
			"uptime_seconds": int64(uptime.Seconds()),
			"uptime_human":   uptime.String(),
			"num_goroutines": runtime.NumGoroutine(),
			"memory_mb":      getMemoryUsage(),
		},
		"environment": env,
		"metrics":     metrics,
	})
}

// getMemoryUsage returns memory usage in MB
func getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc / 1024 / 1024 // Convert bytes to MB
}
