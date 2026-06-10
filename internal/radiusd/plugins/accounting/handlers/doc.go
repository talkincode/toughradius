// Package handlers implements accounting event handlers used by the plugin
// runner.
//
// The package maps Accounting-Request packet types and NAS state transitions to
// focused handler functions that update online-session and accounting records
// while preserving idempotent update semantics across start, interim, stop, and
// accounting on/off flows.
package handlers
