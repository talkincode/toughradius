// Package domain defines ToughRADIUS persistence models and table registry.
//
// Types in this package map authentication, accounting, user, profile, NAS,
// node, and system entities to GORM-backed storage. The Tables export provides
// the canonical migration set used by bootstrap and tests to keep schema
// evolution consistent across PostgreSQL and SQLite.
package domain
