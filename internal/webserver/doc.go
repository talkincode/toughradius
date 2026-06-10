// Package webserver hosts the ToughRADIUS admin HTTP server and middleware.
//
// It builds the Echo runtime, serves the embedded React Admin UI, applies JWT
// authentication and request guards, and registers operational endpoints such as
// readiness checks. Business logic remains in adminapi/app packages while this
// package provides transport, middleware, and route plumbing.
package webserver
