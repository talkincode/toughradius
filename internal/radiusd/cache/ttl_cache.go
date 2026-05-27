package cache

import (
	"sync"
	"time"
)

type entry[T any] struct {
	value   T
	expires time.Time
}

// TTLCache provides a minimal, concurrency-safe cache with a fixed TTL per entry.
type TTLCache[T any] struct {
	ttl        time.Duration
	maxEntries int
	mu         sync.RWMutex
	data       map[string]entry[T]
}

// NewTTLCache creates a TTL-bound cache. maxEntries <= 0 falls back to 1.
func NewTTLCache[T any](ttl time.Duration, maxEntries int) *TTLCache[T] {
	if maxEntries <= 0 {
		maxEntries = 1
	}
	return &TTLCache[T]{
		ttl:        ttl,
		maxEntries: maxEntries,
		data:       make(map[string]entry[T]),
	}
}

// Get retrieves a value if present and not expired.
func (c *TTLCache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	e, ok := c.data[key]
	c.mu.RUnlock()
	if !ok {
		var zero T
		return zero, false
	}
	if time.Now().After(e.expires) {
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		var zero T
		return zero, false
	}
	return e.value, true
}

// Set stores a value and evicts stale entries or random survivors when capacity is exceeded.
func (c *TTLCache[T]) Set(key string, value T) {
	c.mu.Lock()
	c.data[key] = entry[T]{
		value:   value,
		expires: time.Now().Add(c.ttl),
	}
	if len(c.data) > c.maxEntries {
		c.evictExpiredLocked()
		if len(c.data) > c.maxEntries {
			for k := range c.data {
				delete(c.data, k)
				break
			}
		}
	}
	c.mu.Unlock()
}

func (c *TTLCache[T]) evictExpiredLocked() {
	now := time.Now()
	for k, v := range c.data {
		if now.After(v.expires) {
			delete(c.data, k)
		}
	}
}

// Delete removes a specific key from the cache.
func (c *TTLCache[T]) Delete(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

// Clear drops all cached entries.
func (c *TTLCache[T]) Clear() {
	c.mu.Lock()
	c.data = make(map[string]entry[T])
	c.mu.Unlock()
}
