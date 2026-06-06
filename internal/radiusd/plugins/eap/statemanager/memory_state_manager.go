package statemanager

import (
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

// GetState returns the EAP state for the given ID. It is a read-mostly
// operation: the common (non-expired) path holds only a read lock, so
// concurrent EAP handshakes reading distinct or shared states are not
// serialized. Expired states are treated as absent and removed lazily under a
// brief write lock taken only when an expired entry is actually encountered.
//
// When no live state exists it returns eap.ErrStateNotFound, matching the
// EAPStateManager contract and the sentinel used across the EAP subsystem so
// callers can rely on errors.Is(err, eap.ErrStateNotFound).
func (m *MemoryStateManager) GetState(stateID string) (*eap.EAPState, error) {
m.mu.RLock()
entry, ok := m.states[stateID]
if !ok {
m.mu.RUnlock()
return nil, eap.ErrStateNotFound
}
if time.Now().After(entry.expiresAt) {
m.mu.RUnlock()
// Rare path: drop the read lock and remove the expired entry under a
// write lock so the common path above stays read-only.
m.deleteIfExpired(stateID)
return nil, eap.ErrStateNotFound
}

// Return a copy to avoid concurrent modification.
stateCopy := *entry.state
if entry.state.Data != nil {
stateCopy.Data = make(map[string]interface{}, len(entry.state.Data))
for k, v := range entry.state.Data {
stateCopy.Data[k] = v
}
}
m.mu.RUnlock()

return &stateCopy, nil
}

// deleteIfExpired removes the entry for stateID only if it is still present and
// still expired. The re-check under the write lock prevents racing with a
// concurrent SetState that may have refreshed the entry after GetState released
// the read lock.
func (m *MemoryStateManager) deleteIfExpired(stateID string) {
now := time.Now()
m.mu.Lock()
defer m.mu.Unlock()
if entry, ok := m.states[stateID]; ok && now.After(entry.expiresAt) {
delete(m.states, stateID)
}
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
