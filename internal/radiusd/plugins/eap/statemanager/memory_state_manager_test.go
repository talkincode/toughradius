package statemanager

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
)

func TestNewMemoryStateManager(t *testing.T) {
	mgr := NewMemoryStateManager()
	defer mgr.Close()
	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.states)
}

func TestMemoryStateManager_SetState(t *testing.T) {
	mgr := NewMemoryStateManager()
	defer mgr.Close()

	state := &eap.EAPState{
		StateID:  "state-123",
		Username: "testuser",
		Method:   "eap-md5",
		Data: map[string]interface{}{
			"challenge": []byte("random-challenge"),
		},
	}

	err := mgr.SetState("state-id-1", state)
	assert.NoError(t, err)

	stored, err := mgr.GetState("state-id-1")
	assert.NoError(t, err)
	assert.Equal(t, state.StateID, stored.StateID)
	assert.Equal(t, state.Username, stored.Username)
	assert.Equal(t, state.Method, stored.Method)
}

func TestMemoryStateManager_SetState_WithNilData(t *testing.T) {
	mgr := NewMemoryStateManager()
	defer mgr.Close()

	state := &eap.EAPState{
		StateID:  "state-123",
		Username: "testuser",
		Method:   "eap-md5",
		Data:     nil,
	}

	err := mgr.SetState("state-id-1", state)
	assert.NoError(t, err)

	stored, err := mgr.GetState("state-id-1")
	assert.NoError(t, err)
	assert.Nil(t, stored.Data)
}

func TestMemoryStateManager_GetState_NotFound(t *testing.T) {
	mgr := NewMemoryStateManager()
	defer mgr.Close()

	state, err := mgr.GetState("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryStateManager_GetState_ReturnsCopy(t *testing.T) {
	mgr := NewMemoryStateManager()
	defer mgr.Close()

	originalData := map[string]interface{}{
		"key1": "value1",
	}
	state := &eap.EAPState{
		StateID: "state-123",
		Data:    originalData,
	}

	err := mgr.SetState("state-id-1", state)
	require.NoError(t, err)

	retrieved, err := mgr.GetState("state-id-1")
	require.NoError(t, err)
	retrieved.Data["key2"] = "value2"
	retrieved.StateID = "modified"

	original, err := mgr.GetState("state-id-1")
	require.NoError(t, err)
	assert.Equal(t, "state-123", original.StateID)
	_, hasKey2 := original.Data["key2"]
	assert.False(t, hasKey2, "original Data should not have key2")
}

func TestMemoryStateManager_SetState_StoresCopy(t *testing.T) {
	mgr := NewMemoryStateManager()
	defer mgr.Close()

	data := map[string]interface{}{
		"key1": "value1",
	}
	state := &eap.EAPState{
		StateID: "state-123",
		Data:    data,
	}

	err := mgr.SetState("state-id-1", state)
	require.NoError(t, err)

	state.StateID = "modified-external"
	data["key2"] = "value2"

	stored, err := mgr.GetState("state-id-1")
	require.NoError(t, err)
	assert.Equal(t, "state-123", stored.StateID)
	_, hasKey2 := stored.Data["key2"]
	assert.False(t, hasKey2, "stored Data should not have key2")
}

func TestMemoryStateManager_DeleteState(t *testing.T) {
	mgr := NewMemoryStateManager()
	defer mgr.Close()

	state := &eap.EAPState{
		StateID: "state-123",
	}
	err := mgr.SetState("state-id-1", state)
	require.NoError(t, err)

	err = mgr.DeleteState("state-id-1")
	assert.NoError(t, err)

	_, err = mgr.GetState("state-id-1")
	assert.Error(t, err)
}

func TestMemoryStateManager_DeleteState_NonExistent(t *testing.T) {
	mgr := NewMemoryStateManager()
	defer mgr.Close()

	err := mgr.DeleteState("nonexistent")
	assert.NoError(t, err)
}

func TestMemoryStateManager_OverwriteState(t *testing.T) {
	mgr := NewMemoryStateManager()
	defer mgr.Close()

	state1 := &eap.EAPState{
		StateID:  "state-1",
		Username: "user1",
	}
	state2 := &eap.EAPState{
		StateID:  "state-2",
		Username: "user2",
	}

	err := mgr.SetState("state-id", state1)
	require.NoError(t, err)

	err = mgr.SetState("state-id", state2)
	require.NoError(t, err)

	stored, err := mgr.GetState("state-id")
	require.NoError(t, err)
	assert.Equal(t, "state-2", stored.StateID)
	assert.Equal(t, "user2", stored.Username)
}

func TestMemoryStateManager_ConcurrentAccess(t *testing.T) {
	mgr := NewMemoryStateManager()
	defer mgr.Close()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			state := &eap.EAPState{
				StateID: "state",
				Data: map[string]interface{}{
					"index": n,
				},
			}
			_ = mgr.SetState("concurrent-state", state)
		}(i)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = mgr.GetState("concurrent-state")
		}()
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = mgr.DeleteState("concurrent-state")
		}(i)
	}

	wg.Wait()
}

func TestMemoryStateManager_MultipleStates(t *testing.T) {
	mgr := NewMemoryStateManager()
	defer mgr.Close()

	states := map[string]*eap.EAPState{
		"state-1": {StateID: "id-1", Username: "user1"},
		"state-2": {StateID: "id-2", Username: "user2"},
		"state-3": {StateID: "id-3", Username: "user3"},
	}

	for id, state := range states {
		err := mgr.SetState(id, state)
		require.NoError(t, err)
	}

	for id, expected := range states {
		stored, err := mgr.GetState(id)
		require.NoError(t, err)
		assert.Equal(t, expected.StateID, stored.StateID)
		assert.Equal(t, expected.Username, stored.Username)
	}

	err := mgr.DeleteState("state-2")
	require.NoError(t, err)

	_, err = mgr.GetState("state-1")
	assert.NoError(t, err)
	_, err = mgr.GetState("state-2")
	assert.Error(t, err)
	_, err = mgr.GetState("state-3")
	assert.NoError(t, err)
}

func TestMemoryStateManager_ImplementsInterface(t *testing.T) {
	// Verify MemoryStateManager implements EAPStateManager
	var _ eap.EAPStateManager = (*MemoryStateManager)(nil)
}

func TestMemoryStateManager_ExpiresAfterTTL(t *testing.T) {
mgr := NewMemoryStateManagerWithTTL(20*time.Millisecond, 0)
defer mgr.Close()

require.NoError(t, mgr.SetState("s1", &eap.EAPState{StateID: "s1", Username: "u"}))

// Before expiry the state is retrievable.
got, err := mgr.GetState("s1")
require.NoError(t, err)
assert.Equal(t, "u", got.Username)

// After the TTL elapses the state is treated as absent and removed lazily.
time.Sleep(40 * time.Millisecond)
_, err = mgr.GetState("s1")
require.Error(t, err)
assert.Contains(t, err.Error(), "not found")

mgr.mu.RLock()
_, exists := mgr.states["s1"]
mgr.mu.RUnlock()
assert.False(t, exists, "expired state should be removed on read")
}

func TestMemoryStateManager_JanitorSweepsExpired(t *testing.T) {
mgr := NewMemoryStateManagerWithTTL(15*time.Millisecond, 10*time.Millisecond)
defer mgr.Close()

require.NoError(t, mgr.SetState("s1", &eap.EAPState{StateID: "s1"}))
require.NoError(t, mgr.SetState("s2", &eap.EAPState{StateID: "s2"}))

// Wait long enough for the janitor to run at least once after expiry.
assert.Eventually(t, func() bool {
mgr.mu.RLock()
n := len(mgr.states)
mgr.mu.RUnlock()
return n == 0
}, time.Second, 5*time.Millisecond, "janitor should reclaim expired states")
}

// TestMemoryStateManager_DeleteIfExpired_ReCheck verifies the re-check inside
// deleteIfExpired: a live (non-expired) entry must never be removed, while a
// genuinely expired entry is removed. This guards the read path that releases
// the read lock before taking the write lock to evict an expired entry.
func TestMemoryStateManager_DeleteIfExpired_ReCheck(t *testing.T) {
	mgr := NewMemoryStateManagerWithTTL(time.Minute, 0)
	defer mgr.Close()

	require.NoError(t, mgr.SetState("live", &eap.EAPState{StateID: "live"}))

	// A live entry must survive a deleteIfExpired call (simulating a refresh
	// that won the race after GetState observed an expired snapshot).
	mgr.deleteIfExpired("live")
	if _, err := mgr.GetState("live"); err != nil {
		t.Fatalf("live entry must not be deleted by deleteIfExpired: %v", err)
	}

	// Force an already-expired entry and confirm it is evicted.
	mgr.mu.Lock()
	mgr.states["stale"] = &stateEntry{
		state:     &eap.EAPState{StateID: "stale"},
		expiresAt: time.Now().Add(-time.Minute),
	}
	mgr.mu.Unlock()

	mgr.deleteIfExpired("stale")

	mgr.mu.RLock()
	_, exists := mgr.states["stale"]
	mgr.mu.RUnlock()
	assert.False(t, exists, "expired entry should be evicted by deleteIfExpired")
}

// BenchmarkMemoryStateManager_GetStateParallel measures concurrent reads of
// pre-populated live states, the scenario where the read-locked GetState path
// avoids serializing parallel EAP handshakes.
func BenchmarkMemoryStateManager_GetStateParallel(b *testing.B) {
	mgr := NewMemoryStateManagerWithTTL(time.Hour, 0)
	defer mgr.Close()

	const keys = 64
	ids := make([]string, keys)
	for i := 0; i < keys; i++ {
		ids[i] = "state-" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		_ = mgr.SetState(ids[i], &eap.EAPState{StateID: ids[i], Username: "u"})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = mgr.GetState(ids[i%keys])
			i++
		}
	})
}
