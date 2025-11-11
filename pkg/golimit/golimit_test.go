package golimit

import (
	"sync"
	"testing"
	"time"
)

// TestNewGoLimit tests creating a new GoLimit instance
func TestNewGoLimit(t *testing.T) {
	tests := []struct {
		name string
		max  int
	}{
		{"Small limit", 5},
		{"Medium limit", 100},
		{"Large limit", 1000},
		{"Single goroutine", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLimit(tt.max)
			if gl == nil {
				t.Fatal("NewGoLimit returned nil")
			}
			if gl.ch == nil {
				t.Error("channel should be initialized")
			}
			if cap(gl.ch) != tt.max {
				t.Errorf("expected channel capacity %d, got %d", tt.max, cap(gl.ch))
			}
		})
	}
}

// TestGoLimit_Add tests adding to the limiter
func TestGoLimit_Add(t *testing.T) {
	gl := NewGoLimit(3)

	// Should be able to add up to max without blocking
	gl.Add()
	gl.Add()
	gl.Add()

	// Channel should be full now
	if len(gl.ch) != 3 {
		t.Errorf("expected 3 items in channel, got %d", len(gl.ch))
	}
}

// TestGoLimit_Done tests releasing from the limiter
func TestGoLimit_Done(t *testing.T) {
	gl := NewGoLimit(3)

	gl.Add()
	gl.Add()

	if len(gl.ch) != 2 {
		t.Errorf("expected 2 items in channel, got %d", len(gl.ch))
	}

	gl.Done()

	if len(gl.ch) != 1 {
		t.Errorf("expected 1 item in channel after Done, got %d", len(gl.ch))
	}

	gl.Done()

	if len(gl.ch) != 0 {
		t.Errorf("expected 0 items in channel after second Done, got %d", len(gl.ch))
	}
}

// TestGoLimit_AddDone tests the Add/Done cycle
func TestGoLimit_AddDone(t *testing.T) {
	gl := NewGoLimit(5)

	for i := 0; i < 10; i++ {
		gl.Add()
		gl.Done()
	}

	// Should be empty after all Done calls
	if len(gl.ch) != 0 {
		t.Errorf("expected empty channel, got %d items", len(gl.ch))
	}
}

// TestGoLimit_ConcurrentAccess tests concurrent goroutines with limiter
func TestGoLimit_ConcurrentAccess(t *testing.T) {
	maxGoroutines := 10
	totalTasks := 100
	gl := NewGoLimit(maxGoroutines)

	var wg sync.WaitGroup
	var counter int
	var mu sync.Mutex
	var maxConcurrent int
	var currentConcurrent int

	wg.Add(totalTasks)

	for i := 0; i < totalTasks; i++ {
		go func(taskID int) {
			defer wg.Done()

			gl.Add()
			defer gl.Done()

			// Track concurrent goroutines
			mu.Lock()
			currentConcurrent++
			if currentConcurrent > maxConcurrent {
				maxConcurrent = currentConcurrent
			}
			mu.Unlock()

			// Simulate work
			time.Sleep(10 * time.Millisecond)

			mu.Lock()
			counter++
			currentConcurrent--
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify all tasks completed
	if counter != totalTasks {
		t.Errorf("expected %d tasks to complete, got %d", totalTasks, counter)
	}

	// Verify concurrency limit was respected
	if maxConcurrent > maxGoroutines {
		t.Errorf("max concurrent goroutines %d exceeded limit %d", maxConcurrent, maxGoroutines)
	}

	// Channel should be empty
	if len(gl.ch) != 0 {
		t.Errorf("expected empty channel after all work, got %d items", len(gl.ch))
	}
}

// TestGoLimit_BlockingBehavior tests that Add blocks when limit is reached
func TestGoLimit_BlockingBehavior(t *testing.T) {
	gl := NewGoLimit(2)

	// Fill the limiter
	gl.Add()
	gl.Add()

	blocked := make(chan bool, 1)

	// Try to add when full (should block)
	go func() {
		blocked <- true
		gl.Add() // This should block
		blocked <- false
	}()

	// Wait a bit to ensure goroutine has attempted to add
	time.Sleep(50 * time.Millisecond)

	select {
	case <-blocked:
		// Goroutine started
	default:
		t.Fatal("goroutine didn't start")
	}

	// Release one slot
	gl.Done()

	// Now the blocked Add should complete
	select {
	case isBlocked := <-blocked:
		if isBlocked {
			t.Error("Add should have completed after Done")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Add didn't complete after Done")
	}
}

// TestGoLimit_StressTest tests limiter under stress
func TestGoLimit_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	maxGoroutines := 50
	totalTasks := 1000
	gl := NewGoLimit(maxGoroutines)

	var wg sync.WaitGroup
	completed := make([]bool, totalTasks)
	var mu sync.Mutex

	wg.Add(totalTasks)

	for i := 0; i < totalTasks; i++ {
		go func(taskID int) {
			defer wg.Done()

			gl.Add()
			defer gl.Done()

			// Simulate varying work durations
			time.Sleep(time.Duration(taskID%10) * time.Millisecond)

			mu.Lock()
			completed[taskID] = true
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify all tasks completed
	for i, done := range completed {
		if !done {
			t.Errorf("task %d did not complete", i)
		}
	}

	// Channel should be empty
	if len(gl.ch) != 0 {
		t.Errorf("expected empty channel, got %d items", len(gl.ch))
	}
}

// TestGoLimit_ZeroLimit tests edge case with zero limit
func TestGoLimit_ZeroLimit(t *testing.T) {
	gl := NewGoLimit(0)

	// This should block immediately when trying to Add
	// We'll use a timeout to verify
	done := make(chan bool, 1)

	go func() {
		gl.Add()
		done <- true
	}()

	select {
	case <-done:
		t.Error("Add with zero limit should block indefinitely")
	case <-time.After(50 * time.Millisecond):
		// Expected behavior - goroutine is blocked
	}
}

// BenchmarkGoLimit_AddDone benchmarks the Add/Done cycle
func BenchmarkGoLimit_AddDone(b *testing.B) {
	gl := NewGoLimit(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gl.Add()
		gl.Done()
	}
}

// BenchmarkGoLimit_Concurrent benchmarks concurrent usage
func BenchmarkGoLimit_Concurrent(b *testing.B) {
	gl := NewGoLimit(10)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gl.Add()
			// Simulate minimal work
			gl.Done()
		}
	})
}

// BenchmarkGoLimit_HighConcurrency benchmarks high concurrency scenario
func BenchmarkGoLimit_HighConcurrency(b *testing.B) {
	gl := NewGoLimit(1000)
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gl.Add()
			gl.Done()
		}()
	}
	wg.Wait()
}
