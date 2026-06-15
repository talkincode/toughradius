package eap

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestKeyedMutex_SerializesSameKey verifies that contending goroutines using the
// same key never overlap inside the critical section.
func TestKeyedMutex_SerializesSameKey(t *testing.T) {
	km := newKeyedMutex()

	var active int32
	var maxActive int32
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unlock := km.lock("same")
			defer unlock()

			cur := atomic.AddInt32(&active, 1)
			for {
				prev := atomic.LoadInt32(&maxActive)
				if cur <= prev || atomic.CompareAndSwapInt32(&maxActive, prev, cur) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			atomic.AddInt32(&active, -1)
		}()
	}
	wg.Wait()

	if maxActive != 1 {
		t.Fatalf("expected at most 1 goroutine in the critical section, observed %d", maxActive)
	}
}

// TestKeyedMutex_DifferentKeysConcurrent verifies that distinct keys do not block
// one another.
func TestKeyedMutex_DifferentKeysConcurrent(t *testing.T) {
	km := newKeyedMutex()

	unlockA := km.lock("a")
	defer unlockA()

	done := make(chan struct{})
	go func() {
		unlockB := km.lock("b")
		unlockB()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("lock on a different key was blocked")
	}
}

// TestKeyedMutex_ReleasesMapEntry verifies that the underlying map entry is
// removed once no goroutine holds or waits on the key, so the lock map cannot
// grow without bound.
func TestKeyedMutex_ReleasesMapEntry(t *testing.T) {
	km := newKeyedMutex()

	unlock := km.lock("transient")
	unlock()

	km.mu.Lock()
	n := len(km.locks)
	km.mu.Unlock()

	if n != 0 {
		t.Fatalf("expected lock map to be empty after release, found %d entries", n)
	}
}
