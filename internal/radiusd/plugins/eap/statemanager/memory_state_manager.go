package statemanager

import (
"errors"
"sync"
"time"

"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
)

const (
// DefaultStateTTL is how long an EAP state is retained before it is
// considered stale. EAP handshakes normally complete within seconds; this
// generous window tolerates slow clients while ensuring that abandoned
// (never-completed) handshakes cannot leak memory indefinitely.
DefaultStateTTL = 5 * time.Minute

// defaultCleanupInterval is how often the background janitor sweeps expired
// states out of the map.
defaultCleanupInterval = time.Minute
)

// stateEntry wraps a stored EAP state with its expiry deadline.
type stateEntry struct {
state     *eap.EAPState
expiresAt time.Time
}

// MemoryStateManager is an in-memory EAP state manager with TTL-based expiry.
//
// Each stored state has an absolute expiry deadline. Expired states are removed
// lazily on read and proactively by a background janitor, so an EAP handshake
// that is started but never finished cannot accumulate unbounded memory.
type MemoryStateManager struct {
states map[string]*stateEntry
mu     sync.RWMutex
ttl    time.Duration

stopOnce sync.Once
stopCh   chan struct{}
}

// NewMemoryStateManager creates a new in-memory state manager using the default
// TTL and cleanup interval, and starts its background janitor.
func NewMemoryStateManager() *MemoryStateManager {
return NewMemoryStateManagerWithTTL(DefaultStateTTL, defaultCleanupInterval)
}

// NewMemoryStateManagerWithTTL creates a state manager with a custom state TTL
// and janitor interval. A non-positive cleanupInterval disables the background
// janitor (expiry then relies solely on lazy removal during reads).
func NewMemoryStateManagerWithTTL(ttl, cleanupInterval time.Duration) *MemoryStateManager {
m := &MemoryStateManager{
states: make(map[string]*stateEntry),
ttl:    ttl,
stopCh: make(chan struct{}),
}
if cleanupInterval > 0 {
go m.janitor(cleanupInterval)
}
return m
}

// janitor periodically removes expired states until the manager is closed.
func (m *MemoryStateManager) janitor(interval time.Duration) {
ticker := time.NewTicker(interval)
defer ticker.Stop()
for {
select {
case <-ticker.C:
m.sweepExpired()
case <-m.stopCh:
return
}
}
}

// sweepExpired removes all states whose deadline has passed.
func (m *MemoryStateManager) sweepExpired() {
now := time.Now()
m.mu.Lock()
defer m.mu.Unlock()
for id, entry := range m.states {
if now.After(entry.expiresAt) {
delete(m.states, id)
}
}
}

// Close stops the background janitor. It is safe to call multiple times.
func (m *MemoryStateManager) Close() {
m.stopOnce.Do(func() { close(m.stopCh) })
}

// GetState returns the EAP state for the given ID. Expired states are treated as
// absent and removed.
func (m *MemoryStateManager) GetState(stateID string) (*eap.EAPState, error) {
m.mu.Lock()
defer m.mu.Unlock()

entry, ok := m.states[stateID]
if !ok {
return nil, errors.New("state not found")
}
if time.Now().After(entry.expiresAt) {
delete(m.states, stateID)
return nil, errors.New("state not found")
}

// Return a copy to avoid concurrent modification.
stateCopy := *entry.state
if entry.state.Data != nil {
stateCopy.Data = make(map[string]interface{})
for k, v := range entry.state.Data {
stateCopy.Data[k] = v
}
}

return &stateCopy, nil
}

// SetState stores the EAP state and (re)sets its expiry deadline.
func (m *MemoryStateManager) SetState(stateID string, state *eap.EAPState) error {
// Store a copy to avoid external modification.
stateCopy := *state
if state.Data != nil {
stateCopy.Data = make(map[string]interface{})
for k, v := range state.Data {
stateCopy.Data[k] = v
}
}

m.mu.Lock()
defer m.mu.Unlock()
m.states[stateID] = &stateEntry{
state:     &stateCopy,
expiresAt: time.Now().Add(m.ttl),
}
return nil
}

// DeleteState removes the EAP state for the given ID.
func (m *MemoryStateManager) DeleteState(stateID string) error {
m.mu.Lock()
defer m.mu.Unlock()

delete(m.states, stateID)
return nil
}
