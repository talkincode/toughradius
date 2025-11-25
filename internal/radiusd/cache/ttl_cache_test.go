package cache

import (
	"testing"
	"time"
)

func TestTTLCacheSetGetDelete(t *testing.T) {
	cache := NewTTLCache[string](time.Second, 5)

	cache.Set("foo", "bar")

	got, ok := cache.Get("foo")
	if !ok || got != "bar" {
		t.Fatalf("expected hit for foo, got=%q ok=%v", got, ok)
	}

	cache.Delete("foo")
	if _, ok := cache.Get("foo"); ok {
		t.Fatalf("expected foo to be deleted")
	}
}

func TestTTLCacheExpiration(t *testing.T) {
	cache := NewTTLCache[int](10*time.Millisecond, 5)
	cache.Set("num", 42)

	time.Sleep(20 * time.Millisecond)

	if _, ok := cache.Get("num"); ok {
		t.Fatalf("expected num to expire")
	}
}

func TestTTLCacheMaxEntriesEviction(t *testing.T) {
	cache := NewTTLCache[int](time.Second, 2)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	if len(cache.data) != cache.maxEntries {
		t.Fatalf("expected cache to enforce max entries, got %d", len(cache.data))
	}
}

func TestTTLCacheClear(t *testing.T) {
	cache := NewTTLCache[string](time.Second, 2)
	cache.Set("a", "1")
	cache.Set("b", "2")

	cache.Clear()

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	if len(cache.data) != 0 {
		t.Fatalf("expected cache to be empty after Clear, got %d entries", len(cache.data))
	}
}
