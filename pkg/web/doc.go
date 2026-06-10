// Package web provides HTTP-layer helpers shared by the admin web server.
//
// It contains composable middleware and handlers for request metrics, reverse
// proxying, long-lived server-sent events streams, and prequery caching
// behavior. These helpers are framework-oriented utilities and keep business
// rules in adminapi/app packages.
package web
