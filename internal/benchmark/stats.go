// Package benchmark provides components for RADIUS server performance testing and benchmarking.
//
// This package implements a comprehensive benchmarking framework for testing RADIUS authentication
// and accounting services under various load conditions. It supports concurrent testing, detailed
// metrics collection, and flexible reporting.
package benchmark

import (
	"sync/atomic"
	"time"
)

// AuthStats tracks authentication-related metrics for RADIUS testing.
// All counters are thread-safe using atomic operations.
type AuthStats struct {
	// Request counters
	Requests int64 // Total authentication requests sent
	Accepts  int64 // Access-Accept responses received
	Rejects  int64 // Access-Reject responses received
	Drops    int64 // Requests that received no response
	Timeouts int64 // Requests that exceeded timeout threshold

	// Latency distribution (milliseconds)
	LatencyUnder10ms   int64 // Responses received within 0-10ms
	Latency10to100ms   int64 // Responses received within 10-100ms
	Latency100to1000ms int64 // Responses received within 100-1000ms
	LatencyOver1000ms  int64 // Responses received over 1000ms
}

// AcctStats tracks accounting-related metrics for RADIUS testing.
// All counters are thread-safe using atomic operations.
type AcctStats struct {
	// Request type counters
	StartRequests  int64 // Accounting-Start requests sent
	StopRequests   int64 // Accounting-Stop requests sent
	UpdateRequests int64 // Interim-Update requests sent
	Responses      int64 // Accounting-Response messages received
	Drops          int64 // Requests that received no response
	Timeouts       int64 // Requests that exceeded timeout threshold

	// Session tracking
	OnlineCount  int64 // Number of sessions currently online
	OfflineCount int64 // Number of sessions that went offline

	// Latency distribution (milliseconds)
	LatencyUnder10ms   int64 // Responses received within 0-10ms
	Latency10to100ms   int64 // Responses received within 10-100ms
	Latency100to1000ms int64 // Responses received within 100-1000ms
	LatencyOver1000ms  int64 // Responses received over 1000ms
}

// NetworkStats tracks network-level metrics for RADIUS testing.
// All counters are thread-safe using atomic operations.
type NetworkStats struct {
	RequestBytes  int64 // Total bytes sent in requests
	ResponseBytes int64 // Total bytes received in responses
}

// PerformanceMetrics contains calculated performance indicators.
// These are derived from the raw statistics and updated periodically.
type PerformanceMetrics struct {
	// Current rates (per second)
	AuthQPS      int64 // Authentication queries per second
	AcctQPS      int64 // Accounting queries per second
	TotalQPS     int64 // Combined QPS
	OnlineRate   int64 // Online sessions per second
	OfflineRate  int64 // Offline sessions per second
	UploadRate   int64 // Upload bandwidth (bytes/sec)
	DownloadRate int64 // Download bandwidth (bytes/sec)

	// Peak values
	MaxAuthQPS      int64
	MaxAcctQPS      int64
	MaxTotalQPS     int64
	MaxOnlineRate   int64
	MaxOfflineRate  int64
	MaxUploadRate   int64
	MaxDownloadRate int64
}

// Statistics aggregates all benchmark statistics in a structured format.
// This replaces the original BenchmarkStat with clearer organization.
type Statistics struct {
	Auth        AuthStats
	Acct        AcctStats
	Network     NetworkStats
	Performance PerformanceMetrics
	StartTime   time.Time
	LastUpdate  time.Time
}

// NewStatistics creates a new Statistics instance with timestamps initialized.
func NewStatistics() *Statistics {
	now := time.Now()
	return &Statistics{
		StartTime:  now,
		LastUpdate: now,
	}
}

// IncrAuthRequest atomically increments the authentication request counter.
func (s *Statistics) IncrAuthRequest() {
	atomic.AddInt64(&s.Auth.Requests, 1)
}

// IncrAuthAccept atomically increments the authentication accept counter.
func (s *Statistics) IncrAuthAccept() {
	atomic.AddInt64(&s.Auth.Accepts, 1)
}

// IncrAuthReject atomically increments the authentication reject counter.
func (s *Statistics) IncrAuthReject() {
	atomic.AddInt64(&s.Auth.Rejects, 1)
}

// IncrAuthDrop atomically increments the authentication drop counter.
func (s *Statistics) IncrAuthDrop() {
	atomic.AddInt64(&s.Auth.Drops, 1)
}

// IncrAuthTimeout atomically increments the authentication timeout counter.
func (s *Statistics) IncrAuthTimeout() {
	atomic.AddInt64(&s.Auth.Timeouts, 1)
}

// RecordAuthLatency records authentication latency in the appropriate bucket.
// latencyMs is the response time in milliseconds.
func (s *Statistics) RecordAuthLatency(latencyMs int64) {
	switch {
	case latencyMs <= 10:
		atomic.AddInt64(&s.Auth.LatencyUnder10ms, 1)
	case latencyMs <= 100:
		atomic.AddInt64(&s.Auth.Latency10to100ms, 1)
	case latencyMs <= 1000:
		atomic.AddInt64(&s.Auth.Latency100to1000ms, 1)
	default:
		atomic.AddInt64(&s.Auth.LatencyOver1000ms, 1)
	}
}

// IncrAcctStart atomically increments the accounting start counter.
func (s *Statistics) IncrAcctStart() {
	atomic.AddInt64(&s.Acct.StartRequests, 1)
}

// IncrAcctStop atomically increments the accounting stop counter.
func (s *Statistics) IncrAcctStop() {
	atomic.AddInt64(&s.Acct.StopRequests, 1)
}

// IncrAcctUpdate atomically increments the accounting update counter.
func (s *Statistics) IncrAcctUpdate() {
	atomic.AddInt64(&s.Acct.UpdateRequests, 1)
}

// IncrAcctResponse atomically increments the accounting response counter.
func (s *Statistics) IncrAcctResponse() {
	atomic.AddInt64(&s.Acct.Responses, 1)
}

// IncrAcctDrop atomically increments the accounting drop counter.
func (s *Statistics) IncrAcctDrop() {
	atomic.AddInt64(&s.Acct.Drops, 1)
}

// IncrAcctTimeout atomically increments the accounting timeout counter.
func (s *Statistics) IncrAcctTimeout() {
	atomic.AddInt64(&s.Acct.Timeouts, 1)
}

// IncrOnline atomically increments the online session counter.
func (s *Statistics) IncrOnline() {
	atomic.AddInt64(&s.Acct.OnlineCount, 1)
}

// IncrOffline atomically increments the offline session counter.
func (s *Statistics) IncrOffline() {
	atomic.AddInt64(&s.Acct.OfflineCount, 1)
}

// RecordAcctLatency records accounting latency in the appropriate bucket.
// latencyMs is the response time in milliseconds.
func (s *Statistics) RecordAcctLatency(latencyMs int64) {
	switch {
	case latencyMs <= 10:
		atomic.AddInt64(&s.Acct.LatencyUnder10ms, 1)
	case latencyMs <= 100:
		atomic.AddInt64(&s.Acct.Latency10to100ms, 1)
	case latencyMs <= 1000:
		atomic.AddInt64(&s.Acct.Latency100to1000ms, 1)
	default:
		atomic.AddInt64(&s.Acct.LatencyOver1000ms, 1)
	}
}

// AddRequestBytes atomically adds bytes to the request counter.
func (s *Statistics) AddRequestBytes(bytes int64) {
	atomic.AddInt64(&s.Network.RequestBytes, bytes)
}

// AddResponseBytes atomically adds bytes to the response counter.
func (s *Statistics) AddResponseBytes(bytes int64) {
	atomic.AddInt64(&s.Network.ResponseBytes, bytes)
}

// UpdatePerformanceMetrics calculates and updates performance metrics based on time delta.
// This should be called periodically (e.g., every 5 seconds) to update QPS and rate metrics.
//
// The method uses atomic snapshots of counters to calculate rates, then updates peak values.
// It's safe to call concurrently with other metric updates.
func (s *Statistics) UpdatePerformanceMetrics(prevStats *Statistics, duration time.Duration) {
	if duration <= 0 {
		return
	}

	seconds := duration.Seconds()

	// Calculate authentication QPS
	authDelta := (s.Auth.Accepts + s.Auth.Rejects) - (prevStats.Auth.Accepts + prevStats.Auth.Rejects)
	s.Performance.AuthQPS = int64(float64(authDelta) / seconds)

	// Calculate accounting QPS
	acctDelta := s.Acct.Responses - prevStats.Acct.Responses
	s.Performance.AcctQPS = int64(float64(acctDelta) / seconds)

	// Calculate total QPS
	s.Performance.TotalQPS = s.Performance.AuthQPS + s.Performance.AcctQPS

	// Calculate online/offline rates
	onlineDelta := s.Acct.OnlineCount - prevStats.Acct.OnlineCount
	s.Performance.OnlineRate = int64(float64(onlineDelta) / seconds)

	offlineDelta := s.Acct.OfflineCount - prevStats.Acct.OfflineCount
	s.Performance.OfflineRate = int64(float64(offlineDelta) / seconds)

	// Calculate bandwidth rates
	reqBytesDelta := s.Network.RequestBytes - prevStats.Network.RequestBytes
	s.Performance.UploadRate = int64(float64(reqBytesDelta) / seconds)

	respBytesDelta := s.Network.ResponseBytes - prevStats.Network.ResponseBytes
	s.Performance.DownloadRate = int64(float64(respBytesDelta) / seconds)

	// Update peak values
	s.updatePeakValues()

	s.LastUpdate = time.Now()
}

// updatePeakValues updates maximum observed values for performance metrics.
func (s *Statistics) updatePeakValues() {
	if s.Performance.AuthQPS > s.Performance.MaxAuthQPS {
		s.Performance.MaxAuthQPS = s.Performance.AuthQPS
	}
	if s.Performance.AcctQPS > s.Performance.MaxAcctQPS {
		s.Performance.MaxAcctQPS = s.Performance.AcctQPS
	}
	if s.Performance.TotalQPS > s.Performance.MaxTotalQPS {
		s.Performance.MaxTotalQPS = s.Performance.TotalQPS
	}
	if s.Performance.OnlineRate > s.Performance.MaxOnlineRate {
		s.Performance.MaxOnlineRate = s.Performance.OnlineRate
	}
	if s.Performance.OfflineRate > s.Performance.MaxOfflineRate {
		s.Performance.MaxOfflineRate = s.Performance.OfflineRate
	}
	if s.Performance.UploadRate > s.Performance.MaxUploadRate {
		s.Performance.MaxUploadRate = s.Performance.UploadRate
	}
	if s.Performance.DownloadRate > s.Performance.MaxDownloadRate {
		s.Performance.MaxDownloadRate = s.Performance.DownloadRate
	}
}

// Snapshot creates a deep copy of the current statistics.
// This is useful for calculating deltas between time periods.
func (s *Statistics) Snapshot() *Statistics {
	return &Statistics{
		Auth: AuthStats{
			Requests:           atomic.LoadInt64(&s.Auth.Requests),
			Accepts:            atomic.LoadInt64(&s.Auth.Accepts),
			Rejects:            atomic.LoadInt64(&s.Auth.Rejects),
			Drops:              atomic.LoadInt64(&s.Auth.Drops),
			Timeouts:           atomic.LoadInt64(&s.Auth.Timeouts),
			LatencyUnder10ms:   atomic.LoadInt64(&s.Auth.LatencyUnder10ms),
			Latency10to100ms:   atomic.LoadInt64(&s.Auth.Latency10to100ms),
			Latency100to1000ms: atomic.LoadInt64(&s.Auth.Latency100to1000ms),
			LatencyOver1000ms:  atomic.LoadInt64(&s.Auth.LatencyOver1000ms),
		},
		Acct: AcctStats{
			StartRequests:      atomic.LoadInt64(&s.Acct.StartRequests),
			StopRequests:       atomic.LoadInt64(&s.Acct.StopRequests),
			UpdateRequests:     atomic.LoadInt64(&s.Acct.UpdateRequests),
			Responses:          atomic.LoadInt64(&s.Acct.Responses),
			Drops:              atomic.LoadInt64(&s.Acct.Drops),
			Timeouts:           atomic.LoadInt64(&s.Acct.Timeouts),
			OnlineCount:        atomic.LoadInt64(&s.Acct.OnlineCount),
			OfflineCount:       atomic.LoadInt64(&s.Acct.OfflineCount),
			LatencyUnder10ms:   atomic.LoadInt64(&s.Acct.LatencyUnder10ms),
			Latency10to100ms:   atomic.LoadInt64(&s.Acct.Latency10to100ms),
			Latency100to1000ms: atomic.LoadInt64(&s.Acct.Latency100to1000ms),
			LatencyOver1000ms:  atomic.LoadInt64(&s.Acct.LatencyOver1000ms),
		},
		Network: NetworkStats{
			RequestBytes:  atomic.LoadInt64(&s.Network.RequestBytes),
			ResponseBytes: atomic.LoadInt64(&s.Network.ResponseBytes),
		},
		Performance: s.Performance,
		StartTime:   s.StartTime,
		LastUpdate:  s.LastUpdate,
	}
}
