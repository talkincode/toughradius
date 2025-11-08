package statemanager

import (
	"errors"
	"sync"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
)

// MemoryStateManager 基于内存的 EAP 状态管理器
type MemoryStateManager struct {
	states map[string]*eap.EAPState
	mu     sync.RWMutex
}

// NewMemoryStateManager 创建新的内存状态管理器
func NewMemoryStateManager() *MemoryStateManager {
	return &MemoryStateManager{
		states: make(map[string]*eap.EAPState),
	}
}

// GetState 获取 EAP 状态
func (m *MemoryStateManager) GetState(stateID string) (*eap.EAPState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, ok := m.states[stateID]
	if !ok {
		return nil, errors.New("state not found")
	}

	// 返回副本以避免并发修改
	stateCopy := *state
	if state.Data != nil {
		stateCopy.Data = make(map[string]interface{})
		for k, v := range state.Data {
			stateCopy.Data[k] = v
		}
	}

	return &stateCopy, nil
}

// SetState 设置 EAP 状态
func (m *MemoryStateManager) SetState(stateID string, state *eap.EAPState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 存储副本以避免外部修改
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

// DeleteState 删除 EAP 状态
func (m *MemoryStateManager) DeleteState(stateID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.states, stateID)
	return nil
}
