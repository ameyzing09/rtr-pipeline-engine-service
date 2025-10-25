package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics holds HTTP request metrics
type Metrics struct {
	mu sync.RWMutex

	// Request counters by status code
	requestsByStatus map[int]int64

	// Request counters by endpoint
	requestsByEndpoint map[string]int64

	// Error counters by type
	authErrors     int64 // 401/403
	businessErrors int64 // 409/404
	serverErrors   int64 // 500+

	// Request duration tracking
	totalRequests    int64
	totalDurationMs  int64
	minDurationMs    int64
	maxDurationMs    int64
}

var globalMetrics *Metrics
var metricsOnce sync.Once

// GetMetrics returns the singleton metrics instance
func GetMetrics() *Metrics {
	metricsOnce.Do(func() {
		globalMetrics = &Metrics{
			requestsByStatus:   make(map[int]int64),
			requestsByEndpoint: make(map[string]int64),
			minDurationMs:      999999,
		}
	})
	return globalMetrics
}

// MetricsMiddleware tracks HTTP request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		durationMs := duration.Milliseconds()

		// Get status code
		status := c.Writer.Status()

		// Update metrics
		metrics := GetMetrics()
		metrics.mu.Lock()
		defer metrics.mu.Unlock()

		// Increment status counter
		metrics.requestsByStatus[status]++

		// Increment endpoint counter
		endpoint := c.Request.Method + " " + c.FullPath()
		metrics.requestsByEndpoint[endpoint]++

		// Track errors by type
		if status == 401 || status == 403 {
			metrics.authErrors++
		} else if status == 404 || status == 409 {
			metrics.businessErrors++
		} else if status >= 500 {
			metrics.serverErrors++
		}

		// Update duration stats
		metrics.totalRequests++
		metrics.totalDurationMs += durationMs
		if durationMs < metrics.minDurationMs {
			metrics.minDurationMs = durationMs
		}
		if durationMs > metrics.maxDurationMs {
			metrics.maxDurationMs = durationMs
		}
	}
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Copy status counts
	statusCounts := make(map[string]int64)
	for status, count := range m.requestsByStatus {
		statusCounts[string(rune(status+'0'))] = count
	}

	// Copy endpoint counts
	endpointCounts := make(map[string]int64)
	for endpoint, count := range m.requestsByEndpoint {
		endpointCounts[endpoint] = count
	}

	avgDuration := int64(0)
	if m.totalRequests > 0 {
		avgDuration = m.totalDurationMs / m.totalRequests
	}

	return map[string]interface{}{
		"total_requests":  m.totalRequests,
		"auth_errors":     m.authErrors,
		"business_errors": m.businessErrors,
		"server_errors":   m.serverErrors,
		"requests_by_status": map[string]int64{
			"2xx": m.get2xxCount(),
			"4xx": m.get4xxCount(),
			"5xx": m.get5xxCount(),
		},
		"requests_by_endpoint": endpointCounts,
		"latency_ms": map[string]int64{
			"min": m.minDurationMs,
			"max": m.maxDurationMs,
			"avg": avgDuration,
		},
		"error_breakdown": map[string]int64{
			"401_403": m.authErrors,
			"404_409": m.businessErrors,
			"5xx":     m.serverErrors,
		},
	}
}

func (m *Metrics) get2xxCount() int64 {
	count := int64(0)
	for status, cnt := range m.requestsByStatus {
		if status >= 200 && status < 300 {
			count += cnt
		}
	}
	return count
}

func (m *Metrics) get4xxCount() int64 {
	count := int64(0)
	for status, cnt := range m.requestsByStatus {
		if status >= 400 && status < 500 {
			count += cnt
		}
	}
	return count
}

func (m *Metrics) get5xxCount() int64 {
	count := int64(0)
	for status, cnt := range m.requestsByStatus {
		if status >= 500 {
			count += cnt
		}
	}
	return count
}

// Reset clears all metrics (useful for testing)
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestsByStatus = make(map[int]int64)
	m.requestsByEndpoint = make(map[string]int64)
	m.authErrors = 0
	m.businessErrors = 0
	m.serverErrors = 0
	m.totalRequests = 0
	m.totalDurationMs = 0
	m.minDurationMs = 999999
	m.maxDurationMs = 0
}
