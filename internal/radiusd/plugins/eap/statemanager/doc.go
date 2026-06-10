// Package statemanager provides EAP state storage implementations.
//
// The in-memory manager keeps per-session EAP handshake state with TTL-based
// expiry and background cleanup so abandoned handshakes do not accumulate
// unbounded memory.
package statemanager
