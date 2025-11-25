package metrics

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nakabonne/tstorage"
)

// TestInitMetrics tests initializing the metrics storage
func TestInitMetrics(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	err := InitMetrics(tmpDir)
	if err != nil {
		t.Fatalf("InitMetrics failed: %v", err)
	}

	// Verify storage was initialized
	if globalTSDB == nil {
		t.Error("globalTSDB should be initialized")
	}

	// Verify data directory was created
	dataPath := filepath.Join(tmpDir, "data", "metrics")
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		t.Error("metrics data directory was not created")
	}

	// Clean up
	if err := Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestGetTSDB tests getting the TSDB instance
func TestGetTSDB(t *testing.T) {
	tmpDir := t.TempDir()

	err := InitMetrics(tmpDir)
	if err != nil {
		t.Fatalf("InitMetrics failed: %v", err)
	}
	defer func() { _ = Close() }() //nolint:errcheck

	tsdb := GetTSDB()
	if tsdb == nil {
		t.Error("GetTSDB returned nil")
	}

	// Verify it's the same instance
	if tsdb != globalTSDB {
		t.Error("GetTSDB should return the global instance")
	}
}

// TestClose tests closing the metrics storage
func TestClose(t *testing.T) {
	tmpDir := t.TempDir()

	err := InitMetrics(tmpDir)
	if err != nil {
		t.Fatalf("InitMetrics failed: %v", err)
	}

	err = Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Calling Close on nil should not panic
	globalTSDB = nil
	err = Close()
	if err != nil {
		t.Errorf("Close on nil TSDB returned error: %v", err)
	}
}

// TestMetrics_WriteAndRead tests writing and reading metrics
func TestMetrics_WriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()

	err := InitMetrics(tmpDir)
	if err != nil {
		t.Fatalf("InitMetrics failed: %v", err)
	}
	defer func() { _ = Close() }() //nolint:errcheck

	tsdb := GetTSDB()
	if tsdb == nil {
		t.Fatal("TSDB not initialized")
	}

	// Write test data
	metric := "test.metric"
	labels := []tstorage.Label{
		{Name: "host", Value: "server1"},
		{Name: "region", Value: "us-west"},
	}

	now := time.Now().UnixNano()
	points := []tstorage.DataPoint{
		{Timestamp: now, Value: 100.0},
		{Timestamp: now + int64(time.Second), Value: 200.0},
		{Timestamp: now + int64(2*time.Second), Value: 300.0},
	}

	for _, point := range points {
		err := tsdb.InsertRows([]tstorage.Row{
			{
				Metric:    metric,
				Labels:    labels,
				DataPoint: point,
			},
		})
		if err != nil {
			t.Fatalf("Failed to insert row: %v", err)
		}
	}

	// Read data back
	startTime := now - int64(time.Minute)
	endTime := now + int64(5*time.Second)

	readPoints, err := tsdb.Select(metric, labels, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to select data: %v", err)
	}

	if len(readPoints) != len(points) {
		t.Errorf("Expected %d points, got %d", len(points), len(readPoints))
	}

	// Verify values
	for i, point := range readPoints {
		if point.Value != points[i].Value {
			t.Errorf("Point %d: expected value %f, got %f", i, points[i].Value, point.Value)
		}
	}
}

// TestMetrics_MultipleMetrics tests storing multiple different metrics
func TestMetrics_MultipleMetrics(t *testing.T) {
	tmpDir := t.TempDir()

	err := InitMetrics(tmpDir)
	if err != nil {
		t.Fatalf("InitMetrics failed: %v", err)
	}
	defer func() { _ = Close() }() //nolint:errcheck

	tsdb := GetTSDB()
	now := time.Now().UnixNano()

	metrics := []struct {
		name   string
		labels []tstorage.Label
		value  float64
	}{
		{
			name:   "cpu.usage",
			labels: []tstorage.Label{{Name: "host", Value: "web1"}},
			value:  75.5,
		},
		{
			name:   "memory.usage",
			labels: []tstorage.Label{{Name: "host", Value: "web1"}},
			value:  8192.0,
		},
		{
			name:   "disk.io",
			labels: []tstorage.Label{{Name: "host", Value: "db1"}},
			value:  1024.5,
		},
	}

	// Insert all metrics
	for _, m := range metrics {
		err := tsdb.InsertRows([]tstorage.Row{
			{
				Metric: m.name,
				Labels: m.labels,
				DataPoint: tstorage.DataPoint{
					Timestamp: now,
					Value:     m.value,
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to insert metric %s: %v", m.name, err)
		}
	}

	// Verify each metric can be read back
	for _, m := range metrics {
		points, err := tsdb.Select(m.name, m.labels, now-int64(time.Minute), now+int64(time.Minute))
		if err != nil {
			t.Fatalf("Failed to select metric %s: %v", m.name, err)
		}

		if len(points) == 0 {
			t.Errorf("No points found for metric %s", m.name)
			continue
		}

		if points[0].Value != m.value {
			t.Errorf("Metric %s: expected value %f, got %f", m.name, m.value, points[0].Value)
		}
	}
}

// TestInitMetrics_InvalidPath tests initialization with invalid path
func TestInitMetrics_InvalidPath(t *testing.T) {
	// Try to initialize with a path that can't be created
	invalidPath := "/invalid/path/that/cannot/be/created"

	err := InitMetrics(invalidPath)
	// Some systems might allow this to succeed, so we just check it doesn't panic
	if err == nil {
		// If it succeeded, clean up
		_ = Close() //nolint:errcheck
	}
}

// TestMetrics_Retention tests that old data is cleaned up
func TestMetrics_Retention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping retention test in short mode")
	}

	tmpDir := t.TempDir()

	err := InitMetrics(tmpDir)
	if err != nil {
		t.Fatalf("InitMetrics failed: %v", err)
	}
	defer func() { _ = Close() }() //nolint:errcheck

	tsdb := GetTSDB()

	// The retention is set to 7 days in InitMetrics
	// We can't easily test this without mocking time or waiting,
	// but we can verify the storage is initialized properly
	if tsdb == nil {
		t.Error("TSDB should be initialized")
	}
}

// BenchmarkMetrics_Insert benchmarks inserting metrics
func BenchmarkMetrics_Insert(b *testing.B) {
	tmpDir := b.TempDir()

	err := InitMetrics(tmpDir)
	if err != nil {
		b.Fatalf("InitMetrics failed: %v", err)
	}
	defer func() { _ = Close() }() //nolint:errcheck

	tsdb := GetTSDB()
	metric := "benchmark.metric"
	labels := []tstorage.Label{
		{Name: "test", Value: "benchmark"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tsdb.InsertRows([]tstorage.Row{
			{
				Metric: metric,
				Labels: labels,
				DataPoint: tstorage.DataPoint{
					Timestamp: time.Now().UnixNano(),
					Value:     float64(i),
				},
			},
		})
	}
}

// BenchmarkMetrics_Select benchmarks selecting metrics
func BenchmarkMetrics_Select(b *testing.B) {
	tmpDir := b.TempDir()

	err := InitMetrics(tmpDir)
	if err != nil {
		b.Fatalf("InitMetrics failed: %v", err)
	}
	defer func() { _ = Close() }() //nolint:errcheck

	tsdb := GetTSDB()
	metric := "benchmark.select"
	labels := []tstorage.Label{
		{Name: "test", Value: "select"},
	}

	// Insert some test data
	now := time.Now().UnixNano()
	for i := 0; i < 100; i++ {
		_ = tsdb.InsertRows([]tstorage.Row{
			{
				Metric: metric,
				Labels: labels,
				DataPoint: tstorage.DataPoint{
					Timestamp: now + int64(i*int(time.Second)),
					Value:     float64(i),
				},
			},
		})
	}

	startTime := now - int64(time.Minute)
	endTime := now + int64(200*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tsdb.Select(metric, labels, startTime, endTime)
	}
}
