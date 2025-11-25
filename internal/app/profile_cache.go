package app

import (
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// ProfileCache provides in-memory caching for RadiusProfile to optimize authentication performance
// when using dynamic profile linking mode
type ProfileCache struct {
	cache     map[int64]*profileCacheEntry
	mu        sync.RWMutex
	ttl       time.Duration
	db        *gorm.DB
	autoClean bool
	stopClean chan struct{}
}

type profileCacheEntry struct {
	profile   *domain.RadiusProfile
	expiresAt time.Time
}

const (
	// DefaultProfileCacheTTL is the default cache TTL (5 minutes)
	DefaultProfileCacheTTL = 5 * time.Minute

	// CleanupInterval for expired cache entries
	CleanupInterval = 1 * time.Minute
)

// NewProfileCache creates a new profile cache instance
func NewProfileCache(db *gorm.DB, ttl time.Duration) *ProfileCache {
	if ttl == 0 {
		ttl = DefaultProfileCacheTTL
	}

	pc := &ProfileCache{
		cache:     make(map[int64]*profileCacheEntry),
		ttl:       ttl,
		db:        db,
		autoClean: true,
		stopClean: make(chan struct{}),
	}

	// Start background cleanup goroutine
	if pc.autoClean {
		go pc.cleanupLoop()
	}

	return pc
}

// Get retrieves a profile from cache or database
func (pc *ProfileCache) Get(profileID int64) (*domain.RadiusProfile, error) {
	// Try cache first
	pc.mu.RLock()
	entry, found := pc.cache[profileID]
	pc.mu.RUnlock()

	if found && time.Now().Before(entry.expiresAt) {
		// Cache hit and not expired
		return entry.profile, nil
	}

	// Cache miss or expired, fetch from database
	var profile domain.RadiusProfile
	if err := pc.db.Where("id = ?", profileID).First(&profile).Error; err != nil {
		return nil, err
	}

	// Update cache
	pc.Set(profileID, &profile)

	return &profile, nil
}

// Set stores a profile in cache
func (pc *ProfileCache) Set(profileID int64, profile *domain.RadiusProfile) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.cache[profileID] = &profileCacheEntry{
		profile:   profile,
		expiresAt: time.Now().Add(pc.ttl),
	}
}

// Invalidate removes a profile from cache (call when profile is updated/deleted)
func (pc *ProfileCache) Invalidate(profileID int64) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	delete(pc.cache, profileID)
}

// InvalidateAll clears the entire cache
func (pc *ProfileCache) InvalidateAll() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.cache = make(map[int64]*profileCacheEntry)
}

// cleanupLoop periodically removes expired entries
func (pc *ProfileCache) cleanupLoop() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pc.cleanup()
		case <-pc.stopClean:
			return
		}
	}
}

// cleanup removes expired cache entries
func (pc *ProfileCache) cleanup() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	now := time.Now()
	for id, entry := range pc.cache {
		if now.After(entry.expiresAt) {
			delete(pc.cache, id)
		}
	}
}

// Stop halts the background cleanup goroutine
func (pc *ProfileCache) Stop() {
	if pc.autoClean {
		close(pc.stopClean)
	}
}

// Stats returns cache statistics
func (pc *ProfileCache) Stats() map[string]interface{} {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return map[string]interface{}{
		"size": len(pc.cache),
		"ttl":  pc.ttl.String(),
	}
}
