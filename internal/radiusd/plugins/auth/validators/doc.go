// Package validators implements credential validation handlers for the RADIUS
// authentication pipeline.
//
// Validators process protocol-specific credential formats, including PAP, CHAP,
// and MS-CHAP variants, and return normalized auth outcomes that upstream
// pipeline stages can map into policy checks, metrics tags, and final RADIUS
// responses.
package validators
