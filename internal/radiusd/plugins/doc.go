// Package plugins wires ToughRADIUS plugin implementations into the runtime
// registries used by the authentication, accounting, and EAP pipelines.
//
// The package composes concrete validators, policy checkers, response
// enhancers, accounting handlers, and EAP handlers, then registers them through
// the shared registry package during process startup.
//
// Registration is dependency-aware: components that require repositories or
// config providers are only enabled when those dependencies are available.
package plugins
