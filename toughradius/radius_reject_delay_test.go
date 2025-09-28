package toughradius

import (
    "sync"
    "testing"
    "time"
)

func TestRejectItemConcurrentAccess(t *testing.T) {
    ri := &RejectItem{
        Rejects:    1,
        LastReject: time.Now().Add(-15 * time.Second), // Ensure it's over 10 seconds ago
        Lock:       sync.RWMutex{},
    }

    // Test concurrent access that would trigger the lock upgrade path
    var wg sync.WaitGroup
    const numGoroutines = 100
    const iterations = 10

    // Channel to collect any panics
    panicChannel := make(chan interface{}, numGoroutines*iterations)

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            defer func() {
                if r := recover(); r != nil {
                    panicChannel <- r
                }
            }()

            for j := 0; j < iterations; j++ {
                // This should trigger the lock upgrade path due to LastReject being > 10 seconds ago
                result := ri.IsOver(5)
                _ = result // Consume the result to avoid compiler optimization
            }
        }()
    }

    wg.Wait()
    close(panicChannel)

    // Check if any goroutine panicked
    for panic := range panicChannel {
        t.Fatalf("Concurrent access caused panic: %v", panic)
    }
}

func TestRejectCacheConcurrentAccess(t *testing.T) {
    rc := &RejectCache{
        Items: make(map[string]*RejectItem),
        Lock:  sync.RWMutex{},
    }

    // Pre-populate to trigger the cleanup path
    for i := 0; i < 65536; i++ {
        rc.Items["user"+string(rune(i))] = &RejectItem{
            Rejects:    1,
            LastReject: time.Now(),
        }
    }

    var wg sync.WaitGroup
    const numGoroutines = 50
    const iterations = 10

    // Channel to collect any panics
    panicChannel := make(chan interface{}, numGoroutines*iterations)

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            defer func() {
                if r := recover(); r != nil {
                    panicChannel <- r
                }
            }()

            for j := 0; j < iterations; j++ {
                // This should trigger the lock upgrade path due to map size >= 65535
                item := rc.GetItem("testuser" + string(rune(id)))
                _ = item // Consume the result to avoid compiler optimization
            }
        }(i)
    }

    wg.Wait()
    close(panicChannel)

    // Check if any goroutine panicked
    for panic := range panicChannel {
        t.Fatalf("Concurrent access caused panic: %v", panic)
    }
}

func TestRejectItemNormalOperation(t *testing.T) {
    ri := &RejectItem{
        Rejects:    5,
        LastReject: time.Now(),
        Lock:       sync.RWMutex{},
    }

    // Test normal operation (should not trigger lock upgrade)
    result := ri.IsOver(3)
    if !result {
        t.Error("Expected IsOver to return true when rejects > max")
    }

    result = ri.IsOver(10)
    if result {
        t.Error("Expected IsOver to return false when rejects <= max")
    }
}

func TestRejectItemReset(t *testing.T) {
    ri := &RejectItem{
        Rejects:    10,
        LastReject: time.Now().Add(-15 * time.Second), // More than 10 seconds ago
        Lock:       sync.RWMutex{},
    }

    // This should trigger reset of rejects to 0
    result := ri.IsOver(5)
    if result {
        t.Error("Expected IsOver to return false after reset")
    }

    // Verify rejects was reset
    if ri.Rejects != 0 {
        t.Errorf("Expected rejects to be 0 after reset, got %d", ri.Rejects)
    }
}