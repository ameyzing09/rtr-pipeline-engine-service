package cache

import (
	"sync"
	"time"
)

// TenantStatus represents the cached status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "ACTIVE"
	TenantStatusPending   TenantStatus = "PENDING"
	TenantStatusSuspended TenantStatus = "SUSPENDED"
	TenantStatusDeleted   TenantStatus = "DELETED"
	TenantStatusFailed    TenantStatus = "FAILED"
)

// CachedTenant holds tenant status with expiration
type CachedTenant struct {
	Status    TenantStatus
	ExpiresAt time.Time
}

// TenantCache provides thread-safe caching of tenant statuses
type TenantCache struct {
	mu    sync.RWMutex
	cache map[string]*CachedTenant
	ttl   time.Duration
}

// NewTenantCache creates a new tenant cache with the specified TTL
func NewTenantCache(ttl time.Duration) *TenantCache {
	tc := &TenantCache{
		cache: make(map[string]*CachedTenant),
		ttl:   ttl,
	}

	// Start background cleanup goroutine
	go tc.cleanupExpired()

	return tc
}

// Get retrieves a tenant status from cache
// Returns (status, true) if found and not expired
// Returns ("", false) if not found or expired
func (tc *TenantCache) Get(tenantID string) (TenantStatus, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	cached, exists := tc.cache[tenantID]
	if !exists {
		return "", false
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return "", false
	}

	return cached.Status, true
}

// Set stores a tenant status in cache with TTL
func (tc *TenantCache) Set(tenantID string, status TenantStatus) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.cache[tenantID] = &CachedTenant{
		Status:    status,
		ExpiresAt: time.Now().Add(tc.ttl),
	}
}

// Delete removes a tenant from cache
func (tc *TenantCache) Delete(tenantID string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	delete(tc.cache, tenantID)
}

// Clear removes all entries from cache
func (tc *TenantCache) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.cache = make(map[string]*CachedTenant)
}

// Size returns the current number of cached entries
func (tc *TenantCache) Size() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	return len(tc.cache)
}

// cleanupExpired periodically removes expired entries from cache
func (tc *TenantCache) cleanupExpired() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		tc.mu.Lock()
		now := time.Now()
		for tenantID, cached := range tc.cache {
			if now.After(cached.ExpiresAt) {
				delete(tc.cache, tenantID)
			}
		}
		tc.mu.Unlock()
	}
}

// Global tenant cache instance
var globalCache *TenantCache
var once sync.Once

// GetGlobalCache returns the singleton tenant cache instance
func GetGlobalCache() *TenantCache {
	once.Do(func() {
		globalCache = NewTenantCache(60 * time.Second) // 60s TTL as required
	})
	return globalCache
}
