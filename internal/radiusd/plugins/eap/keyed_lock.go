package eap

import "sync"

// keyedMutex provides per-key mutual exclusion. Goroutines contending for the
// same key are serialized, while goroutines holding different keys proceed
// concurrently.
//
// It is used to serialize the GetState → handler → SetState sequence per EAP
// state ID. The per-handshake *tlsengine.Engine stored in EAPState.Data is
// shared by reference across the shallow state copies returned by the state
// manager, and the engine is documented as not safe for concurrent Process
// calls. Without serialization a NAS retransmit processed concurrently with the
// original request can desync the TLS handshake (lost outbound fragments,
// inbound reassembler drift, silently dropped SetState writes).
type keyedMutex struct {
	mu    sync.Mutex
	locks map[string]*refLock
}

// refLock is a mutex with a reference count so the map entry can be removed once
// no goroutine is using or waiting on it.
type refLock struct {
	mu  sync.Mutex
	ref int
}

// newKeyedMutex returns a ready-to-use keyedMutex.
func newKeyedMutex() *keyedMutex {
	return &keyedMutex{locks: make(map[string]*refLock)}
}

// lock acquires the lock for key and returns an unlock function. The caller must
// invoke the returned function exactly once (typically via defer) to release the
// lock and drop the reference.
func (k *keyedMutex) lock(key string) func() {
	k.mu.Lock()
	rl, ok := k.locks[key]
	if !ok {
		rl = &refLock{}
		k.locks[key] = rl
	}
	rl.ref++
	k.mu.Unlock()

	rl.mu.Lock()

	return func() {
		rl.mu.Unlock()
		k.mu.Lock()
		rl.ref--
		if rl.ref == 0 {
			delete(k.locks, key)
		}
		k.mu.Unlock()
	}
}
