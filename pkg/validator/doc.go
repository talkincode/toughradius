// Package validator integrates go-playground/validator with Echo and normalizes
// validation errors into a consistent API payload shape.
//
// The package registers project-specific tags (for example RADIUS status and
// port constraints) and is intended for request payload validation in admin API
// handlers.
package validator
