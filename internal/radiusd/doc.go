// Package radiusd implements ToughRADIUS protocol-serving pipelines.
//
// The package hosts the authentication, accounting, and dynamic authorization
// (CoA/Disconnect) request stages used by UDP RADIUS and RadSec frontends. It
// composes plugin runners, validators, checkers, and response enhancers to keep
// protocol handling extensible while preserving stable error mapping and
// metrics semantics.
//
// The package-level entry points are designed for server startup wiring, while
// subpackages under internal/radiusd/plugins and internal/radiusd/repository
// provide pluggable method handlers and persistence interfaces.
package radiusd
