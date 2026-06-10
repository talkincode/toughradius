// Package cache provides lightweight, concurrency-safe in-memory caches used by
// RADIUS protocol hot paths.
//
// The package currently exposes a generic TTL cache that stores entries by
// string key, evicts expired entries lazily, and bounds memory growth with a
// configurable maximum entry count.
package cache
