// Package app wires ToughRADIUS runtime dependencies and process lifecycle.
//
// It owns application bootstrap, database migration, dynamic configuration,
// profile cache refresh, scheduled background jobs, and metrics/logging setup.
// Web handlers and protocol services consume this package through small
// provider interfaces so startup and test wiring remain centralized.
package app
