package app

import (
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Create test tables
	err = db.AutoMigrate(&domain.RadiusProfile{})
	require.NoError(t, err)

	return db
}

// createTestProfile creates a test profile in the database
func createTestProfile(t *testing.T, db *gorm.DB, id int64, name string) *domain.RadiusProfile {
	profile := &domain.RadiusProfile{
		ID:        id,
		Name:      name,
		Status:    "enabled",
		AddrPool:  "pool1",
		ActiveNum: 5,
		UpRate:    10240,
		DownRate:  20480,
		Domain:    "test.com",
		BindMac:   1,
		BindVlan:  0,
	}
	err := db.Create(profile).Error
	require.NoError(t, err)
	return profile
}

func TestNewProfileCache(t *testing.T) {
	db := setupTestDB(t)

	t.Run("with default TTL", func(t *testing.T) {
		cache := NewProfileCache(db, 0)
		defer cache.Stop()

		assert.NotNil(t, cache)
		assert.Equal(t, DefaultProfileCacheTTL, cache.ttl)
		assert.NotNil(t, cache.cache)
		assert.True(t, cache.autoClean)
	})

	t.Run("with custom TTL", func(t *testing.T) {
		customTTL := 10 * time.Minute
		cache := NewProfileCache(db, customTTL)
		defer cache.Stop()

		assert.Equal(t, customTTL, cache.ttl)
	})
}

func TestProfileCache_Get(t *testing.T) {
	db := setupTestDB(t)
	cache := NewProfileCache(db, time.Minute)
	defer cache.Stop()

	t.Run("cache miss then hit", func(t *testing.T) {
		// Create test profile
		expectedProfile := createTestProfile(t, db, 1, "profile1")

		// First call - cache miss, should fetch from DB
		profile, err := cache.Get(1)
		require.NoError(t, err)
		assert.Equal(t, expectedProfile.Name, profile.Name)
		assert.Equal(t, expectedProfile.UpRate, profile.UpRate)
		assert.Equal(t, expectedProfile.DownRate, profile.DownRate)

		// Second call - should be from cache
		profile2, err := cache.Get(1)
		require.NoError(t, err)
		assert.Equal(t, expectedProfile.Name, profile2.Name)
	})

	t.Run("profile not found", func(t *testing.T) {
		_, err := cache.Get(9999)
		assert.Error(t, err)
	})

	t.Run("expired entry refetches from DB", func(t *testing.T) {
		// Create cache with very short TTL
		shortCache := &ProfileCache{
			cache:     make(map[int64]*profileCacheEntry),
			ttl:       1 * time.Millisecond,
			db:        db,
			autoClean: false,
			stopClean: make(chan struct{}),
		}

		createTestProfile(t, db, 2, "profile2")

		// First call
		_, err := shortCache.Get(2)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Update profile in DB
		db.Model(&domain.RadiusProfile{}).Where("id = ?", 2).Update("name", "profile2-updated")

		// Should refetch from DB
		profile, err := shortCache.Get(2)
		require.NoError(t, err)
		assert.Equal(t, "profile2-updated", profile.Name)
	})
}

func TestProfileCache_Set(t *testing.T) {
	db := setupTestDB(t)
	cache := NewProfileCache(db, time.Minute)
	defer cache.Stop()

	profile := &domain.RadiusProfile{
		ID:       100,
		Name:     "manual-set",
		UpRate:   1024,
		DownRate: 2048,
	}

	cache.Set(100, profile)

	// Verify it's in cache
	cache.mu.RLock()
	entry, found := cache.cache[100]
	cache.mu.RUnlock()

	assert.True(t, found)
	assert.Equal(t, "manual-set", entry.profile.Name)
	assert.True(t, time.Now().Before(entry.expiresAt))
}

func TestProfileCache_Invalidate(t *testing.T) {
	db := setupTestDB(t)
	cache := NewProfileCache(db, time.Minute)
	defer cache.Stop()

	// Add entry to cache
	profile := &domain.RadiusProfile{ID: 1, Name: "test"}
	cache.Set(1, profile)

	// Verify it's in cache
	cache.mu.RLock()
	_, found := cache.cache[1]
	cache.mu.RUnlock()
	assert.True(t, found)

	// Invalidate
	cache.Invalidate(1)

	// Verify it's removed
	cache.mu.RLock()
	_, found = cache.cache[1]
	cache.mu.RUnlock()
	assert.False(t, found)
}

func TestProfileCache_InvalidateAll(t *testing.T) {
	db := setupTestDB(t)
	cache := NewProfileCache(db, time.Minute)
	defer cache.Stop()

	// Add multiple entries
	cache.Set(1, &domain.RadiusProfile{ID: 1, Name: "profile1"})
	cache.Set(2, &domain.RadiusProfile{ID: 2, Name: "profile2"})
	cache.Set(3, &domain.RadiusProfile{ID: 3, Name: "profile3"})

	// Verify cache size
	cache.mu.RLock()
	size := len(cache.cache)
	cache.mu.RUnlock()
	assert.Equal(t, 3, size)

	// Invalidate all
	cache.InvalidateAll()

	// Verify cache is empty
	cache.mu.RLock()
	size = len(cache.cache)
	cache.mu.RUnlock()
	assert.Equal(t, 0, size)
}

func TestProfileCache_Cleanup(t *testing.T) {
	db := setupTestDB(t)

	cache := &ProfileCache{
		cache:     make(map[int64]*profileCacheEntry),
		ttl:       time.Minute,
		db:        db,
		autoClean: false,
		stopClean: make(chan struct{}),
	}

	// Add expired and non-expired entries
	now := time.Now()
	cache.cache[1] = &profileCacheEntry{
		profile:   &domain.RadiusProfile{ID: 1, Name: "expired"},
		expiresAt: now.Add(-time.Hour), // Already expired
	}
	cache.cache[2] = &profileCacheEntry{
		profile:   &domain.RadiusProfile{ID: 2, Name: "valid"},
		expiresAt: now.Add(time.Hour), // Still valid
	}
	cache.cache[3] = &profileCacheEntry{
		profile:   &domain.RadiusProfile{ID: 3, Name: "expired2"},
		expiresAt: now.Add(-time.Minute), // Already expired
	}

	// Run cleanup
	cache.cleanup()

	// Verify only valid entry remains
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	assert.Len(t, cache.cache, 1)
	_, found := cache.cache[2]
	assert.True(t, found)
}

func TestProfileCache_Stats(t *testing.T) {
	db := setupTestDB(t)
	cache := NewProfileCache(db, 3*time.Minute)
	defer cache.Stop()

	// Empty cache stats
	stats := cache.Stats()
	assert.Equal(t, 0, stats["size"])
	assert.Equal(t, "3m0s", stats["ttl"])

	// Add entries
	cache.Set(1, &domain.RadiusProfile{ID: 1})
	cache.Set(2, &domain.RadiusProfile{ID: 2})

	stats = cache.Stats()
	assert.Equal(t, 2, stats["size"])
}

func TestProfileCache_ConcurrentAccess(t *testing.T) {
	// Use shared memory database for concurrent access
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = db.AutoMigrate(&domain.RadiusProfile{})
	require.NoError(t, err)

	cache := NewProfileCache(db, time.Minute)
	defer cache.Stop()

	// Create test profiles
	for i := int64(1); i <= 10; i++ {
		db.Create(&domain.RadiusProfile{ID: i, Name: "profile", UpRate: 1024})
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			_, err := cache.Get((id % 10) + 1)
			if err != nil {
				errChan <- err
			}
		}(int64(i))
	}

	// Concurrent writes
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			cache.Set(id+100, &domain.RadiusProfile{ID: id + 100, Name: "concurrent"})
		}(int64(i))
	}

	// Concurrent invalidations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			cache.Invalidate(id + 100)
		}(int64(i))
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		t.Errorf("concurrent access error: %v", err)
	}
}

func TestProfileCache_Stop(t *testing.T) {
	db := setupTestDB(t)
	cache := NewProfileCache(db, time.Minute)

	// Stop should not panic
	assert.NotPanics(t, func() {
		cache.Stop()
	})
}

func TestProfileCache_GetWithNilDB(t *testing.T) {
	cache := &ProfileCache{
		cache:     make(map[int64]*profileCacheEntry),
		ttl:       time.Minute,
		db:        nil,
		autoClean: false,
		stopClean: make(chan struct{}),
	}

	// Should panic when trying to get from nil DB (cache miss)
	assert.Panics(t, func() {
		_, _ = cache.Get(1) //nolint:errcheck
	})
}

func TestProfileCache_TTLExpiration(t *testing.T) {
	db := setupTestDB(t)

	cache := &ProfileCache{
		cache:     make(map[int64]*profileCacheEntry),
		ttl:       50 * time.Millisecond,
		db:        db,
		autoClean: false,
		stopClean: make(chan struct{}),
	}

	createTestProfile(t, db, 1, "profile1")

	// First fetch
	_, err := cache.Get(1)
	require.NoError(t, err)

	// Verify entry exists and is valid
	cache.mu.RLock()
	entry, found := cache.cache[1]
	cache.mu.RUnlock()
	assert.True(t, found)
	assert.True(t, time.Now().Before(entry.expiresAt))

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Entry should be expired now
	cache.mu.RLock()
	entry, found = cache.cache[1]
	cache.mu.RUnlock()
	assert.True(t, found) // Still in cache but expired
	assert.True(t, time.Now().After(entry.expiresAt))

	// Get should refetch
	profile, err := cache.Get(1)
	require.NoError(t, err)
	assert.Equal(t, "profile1", profile.Name)

	// Entry should be refreshed
	cache.mu.RLock()
	entry, found = cache.cache[1]
	cache.mu.RUnlock()
	assert.True(t, found)
	assert.True(t, time.Now().Before(entry.expiresAt))
}

func TestProfileCache_AutoCleanupLoop(t *testing.T) {
	db := setupTestDB(t)

	// Create cache with very short cleanup interval for testing
	cache := &ProfileCache{
		cache:     make(map[int64]*profileCacheEntry),
		ttl:       10 * time.Millisecond,
		db:        db,
		autoClean: true,
		stopClean: make(chan struct{}),
	}

	// Start cleanup manually with short interval
	go func() {
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cache.cleanup()
			case <-cache.stopClean:
				return
			}
		}
	}()

	// Add entry that will expire
	cache.Set(1, &domain.RadiusProfile{ID: 1, Name: "will-expire"})

	// Verify entry exists
	cache.mu.RLock()
	_, found := cache.cache[1]
	cache.mu.RUnlock()
	assert.True(t, found)

	// Wait for expiration and cleanup
	time.Sleep(50 * time.Millisecond)

	// Entry should be cleaned up
	cache.mu.RLock()
	_, found = cache.cache[1]
	cache.mu.RUnlock()
	assert.False(t, found)

	cache.Stop()
}

func BenchmarkProfileCache_Get(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	_ = db.AutoMigrate(&domain.RadiusProfile{}) //nolint:errcheck
	db.Create(&domain.RadiusProfile{ID: 1, Name: "benchmark", UpRate: 1024})

	cache := NewProfileCache(db, time.Minute)
	defer cache.Stop()

	// Warm up cache
	_, _ = cache.Get(1) //nolint:errcheck

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(1) //nolint:errcheck
	}
}

func BenchmarkProfileCache_Set(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	cache := NewProfileCache(db, time.Minute)
	defer cache.Stop()

	profile := &domain.RadiusProfile{ID: 1, Name: "benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(1, profile)
	}
}

func BenchmarkProfileCache_ConcurrentGet(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	_ = db.AutoMigrate(&domain.RadiusProfile{}) //nolint:errcheck
	for i := int64(1); i <= 100; i++ {
		db.Create(&domain.RadiusProfile{ID: i, Name: "benchmark", UpRate: 1024})
	}

	cache := NewProfileCache(db, time.Minute)
	defer cache.Stop()

	// Warm up cache
	for i := int64(1); i <= 100; i++ {
		_, _ = cache.Get(i) //nolint:errcheck
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		id := int64(1)
		for pb.Next() {
			_, _ = cache.Get(id) //nolint:errcheck
			id = (id % 100) + 1
		}
	})
}
