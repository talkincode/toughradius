package statemanager

import (
	"errors"
	"sync"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
)

// MemoryStateManager is an in-memory EAP state manager
type MemoryStateManager struct {
	states map[string]*eap.EAPState
	mu     sync.RWMutex
}

// NewMemoryStateManager creates a new in-memory state manager
func NewMemoryStateManager() *MemoryStateManager {
	return &MemoryStateManager{
		states: make(map[string]*eap.EAPState),
	}
}

// GetState get EAP Status
func (m *MemoryStateManager) GetState(stateID string) (*eap.EAPState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, ok := m.states[stateID]
	if !ok {
		return nil, errors.New("state not found")
	}

	// Returns a copy to avoid concurrent modification
	stateCopy := *state
	if state.Data != nil {
		stateCopy.Data = make(map[string]interface{})
		for k, v := range state.Data {
			stateCopy.Data[k] = v
		}
	}

	return &stateCopy, nil
}

// SetState stores the EAP status
func (m *MemoryStateManager) SetState(stateID string, state *eap.EAPState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store a copy to avoid external modification
	stateCopy := *state
	if state.Data != nil {
		stateCopy.Data = make(map[string]interface{})
		for k, v := range state.Data {
			stateCopy.Data[k] = v
		}
	}

	m.states[stateID] = &stateCopy
	return nil
}

// DeleteState Delete EAP Status
func (m *MemoryStateManager) DeleteState(stateID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.states, stateID)
	return nil
}
