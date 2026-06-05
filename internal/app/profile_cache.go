package app

import (
	"errors"
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

// NewProfileCache creates a new profile cache instance with automatic cleanup.
// The cache reduces database load during RADIUS authentication by storing frequently
// accessed profile configurations in memory.
//
// Parameters:
//   - db: GORM database instance for cache misses
//   - ttl: Time-to-live for cache entries (0 uses DefaultProfileCacheTTL = 5 minutes)
//
// Returns:
//   - *ProfileCache: Cache instance with background cleanup goroutine started
//
// Side effects:
//   - Starts a background goroutine for periodic cleanup (every 1 minute)
//   - Goroutine stops when Stop() is called
//
// Example:
//
//	cache := app.NewProfileCache(db, 10*time.Minute)
//	defer cache.Stop()  // Important: stop cleanup goroutine on shutdown
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

// Get retrieves a profile from cache or database with automatic cache-aside pattern.
// On cache miss, the profile is fetched from database and stored in cache.
//
// Parameters:
//   - profileID: Unique profile identifier
//
// Returns:
//   - *domain.RadiusProfile: Profile object (never nil if error is nil)
//   - error: Database error if profile not found or query fails
//
// Behavior:
//   - Cache hit: Returns cached profile if not expired
//   - Cache miss/expired: Fetches from DB, updates cache, returns profile
//   - Thread-safe: Uses RWMutex for concurrent access
//
// Example:
//
//	profile, err := cache.Get(user.ProfileID)
//	if err != nil {
//	    return fmt.Errorf("profile not found: %w", err)
//	}
func (pc *ProfileCache) Get(profileID int64) (*domain.RadiusProfile, error) {
	if pc.db == nil {
		return nil, errors.New("profile cache: database connection is nil")
	}

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

// Set manually stores a profile in cache with the configured TTL.
// This is useful for pre-warming the cache or updating after database writes.
//
// Parameters:
//   - profileID: Profile identifier (key)
//   - profile: Profile object to cache (must not be nil)
//
// Side effects:
//   - Overwrites existing cache entry if present
//   - Sets expiration time to now + TTL
//
// Example:
//
//	// Pre-warm cache after creating new profile
//	db.Create(&profile)
//	cache.Set(profile.ID, &profile)
func (pc *ProfileCache) Set(profileID int64, profile *domain.RadiusProfile) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.cache[profileID] = &profileCacheEntry{
		profile:   profile,
		expiresAt: time.Now().Add(pc.ttl),
	}
}

// Invalidate removes a specific profile from cache.
// Call this after updating or deleting a profile to ensure cache consistency.
//
// Parameters:
//   - profileID: Profile ID to remove from cache
//
// Side effects:
//   - Removes cache entry (no-op if not present)
//   - Next Get() will fetch fresh data from database
//
// Example:
//
//	// Update profile bandwidth limit
//	db.Model(&profile).Update("bandwidth", newValue)
//	cache.Invalidate(profile.ID)  // Force re-fetch on next access
func (pc *ProfileCache) Invalidate(profileID int64) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	delete(pc.cache, profileID)
}

// InvalidateAll clears the entire cache by discarding all entries.
// This is useful for bulk updates or when cache coherence is uncertain.
//
// Side effects:
//   - Replaces cache map with a new empty map
//   - All subsequent Get() calls will trigger database queries
//
// Example:
//
//	// After bulk profile update
//	db.Model(&domain.RadiusProfile{}).Where("node_id = ?", nodeID).Update("status", "disabled")
//	cache.InvalidateAll()  // Clear entire cache
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

// Stop halts the background cleanup goroutine gracefully.
// Call this during application shutdown to prevent goroutine leaks.
//
// Side effects:
//   - Closes stopClean channel, signaling cleanup goroutine to exit
//   - Cleanup goroutine terminates within CleanupInterval (1 minute max)
//   - No-op if autoClean was disabled at creation
//
// Example:
//
//	func (app *Application) Shutdown() {
//	    app.profileCache.Stop()  // Stop background cleanup
//	    app.db.Close()
//	}
func (pc *ProfileCache) Stop() {
	if pc.autoClean {
		close(pc.stopClean)
	}
}

// Stats returns current cache statistics for monitoring and debugging.
//
// Returns:
//   - map[string]interface{}: Statistics with keys:
//   - "size" (int): Number of cached profiles
//   - "ttl" (string): Cache TTL duration (e.g., "5m0s")
//
// Example:
//
//	stats := cache.Stats()
//	log.Printf("Cache size: %v, TTL: %v", stats["size"], stats["ttl"])
func (pc *ProfileCache) Stats() map[string]interface{} {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return map[string]interface{}{
		"size": len(pc.cache),
		"ttl":  pc.ttl.String(),
	}
}
