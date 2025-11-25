package statemanager

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
)

func TestNewMemoryStateManager(t *testing.T) {
	mgr := NewMemoryStateManager()
	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.states)
}

func TestMemoryStateManager_SetState(t *testing.T) {
	mgr := NewMemoryStateManager()

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

	state, err := mgr.GetState("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryStateManager_GetState_ReturnsCopy(t *testing.T) {
	mgr := NewMemoryStateManager()

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

	err := mgr.DeleteState("nonexistent")
	assert.NoError(t, err)
}

func TestMemoryStateManager_OverwriteState(t *testing.T) {
	mgr := NewMemoryStateManager()

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
