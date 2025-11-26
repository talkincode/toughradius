// Package metrics provides simple in-memory metrics collection.
// This is a placeholder implementation; consider using Prometheus for production metrics.
package metrics

import (
	"sync"
	"sync/atomic"
)

const (
	// MaxMetrics is the maximum number of unique metric names allowed
	// to prevent unbounded memory growth from dynamic metric names
	MaxMetrics = 1000
)

// Counter represents a simple atomic counter for metrics
type Counter struct {
	value int64
}

// Inc increments the counter by 1
func (c *Counter) Inc() {
	atomic.AddInt64(&c.value, 1)
}

// Add adds the given value to the counter
func (c *Counter) Add(delta int64) {
	atomic.AddInt64(&c.value, delta)
}

// Value returns the current counter value
func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

// Reset resets the counter to 0
func (c *Counter) Reset() {
	atomic.StoreInt64(&c.value, 0)
}

// Gauge represents a value that can go up and down
type Gauge struct {
	value int64
}

// Set sets the gauge to the given value
func (g *Gauge) Set(value int64) {
	atomic.StoreInt64(&g.value, value)
}

// Value returns the current gauge value
func (g *Gauge) Value() int64 {
	return atomic.LoadInt64(&g.value)
}

// MetricsStore holds all application metrics
type MetricsStore struct {
	mu       sync.RWMutex
	counters map[string]*Counter
	gauges   map[string]*Gauge
}

var globalStore *MetricsStore

// InitMetrics initializes the metrics store
// The workdir parameter is kept for API compatibility but not used
func InitMetrics(_ string) error {
	globalStore = &MetricsStore{
		counters: make(map[string]*Counter),
		gauges:   make(map[string]*Gauge),
	}
	return nil
}

// GetStore returns the global metrics store
func GetStore() *MetricsStore {
	return globalStore
}

// Counter returns or creates a counter with the given name
// Returns nil if MaxMetrics limit is reached
func (s *MetricsStore) Counter(name string) *Counter {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.counters[name]; ok {
		return c
	}
	// Check limit before creating new counter
	if len(s.counters) >= MaxMetrics {
		return &Counter{} // Return a dummy counter that won't be stored
	}
	c := &Counter{}
	s.counters[name] = c
	return c
}

// Gauge returns or creates a gauge with the given name
// Returns nil if MaxMetrics limit is reached
func (s *MetricsStore) Gauge(name string) *Gauge {
	s.mu.Lock()
	defer s.mu.Unlock()

	if g, ok := s.gauges[name]; ok {
		return g
	}
	// Check limit before creating new gauge
	if len(s.gauges) >= MaxMetrics {
		return &Gauge{} // Return a dummy gauge that won't be stored
	}
	g := &Gauge{}
	s.gauges[name] = g
	return g
}

// GetCounterValue returns the value of a counter, or 0 if not found
func (s *MetricsStore) GetCounterValue(name string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if c, ok := s.counters[name]; ok {
		return c.Value()
	}
	return 0
}

// GetGaugeValue returns the value of a gauge, or 0 if not found
func (s *MetricsStore) GetGaugeValue(name string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if g, ok := s.gauges[name]; ok {
		return g.Value()
	}
	return 0
}

// GetAllCounters returns a map of all counter names to their values
func (s *MetricsStore) GetAllCounters() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]int64, len(s.counters))
	for name, counter := range s.counters {
		result[name] = counter.Value()
	}
	return result
}

// GetAllGauges returns a map of all gauge names to their values
func (s *MetricsStore) GetAllGauges() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]int64, len(s.gauges))
	for name, gauge := range s.gauges {
		result[name] = gauge.Value()
	}
	return result
}

// Close closes the metrics store (no-op for in-memory store)
func Close() error {
	return nil
}

// Inc increments a counter by name (convenience function)
func Inc(name string) {
	if globalStore != nil {
		globalStore.Counter(name).Inc()
	}
}

// Add adds to a counter by name (convenience function)
func Add(name string, delta int64) {
	if globalStore != nil {
		globalStore.Counter(name).Add(delta)
	}
}

// SetGauge sets a gauge value by name (convenience function)
func SetGauge(name string, value int64) {
	if globalStore != nil {
		globalStore.Gauge(name).Set(value)
	}
}

// GetCounter returns a counter value by name (convenience function)
func GetCounter(name string) int64 {
	if globalStore != nil {
		return globalStore.GetCounterValue(name)
	}
	return 0
}
