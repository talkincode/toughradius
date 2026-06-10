// Package checkers implements post-credential authorization checks in the
// authentication pipeline.
//
// These checks enforce account status, expiration policy, concurrent-session
// limits, and MAC/VLAN bindings after password or EAP validation succeeds, so
// Access-Accept is only produced when policy constraints are satisfied.
package checkers
