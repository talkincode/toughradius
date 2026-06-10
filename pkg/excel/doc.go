// Package excel exports tabular structs to XLSX files for admin-side data
// download workflows.
//
// It reflects exported struct fields, derives column names from db/json tags,
// and writes rows to either a caller-provided path or a generated temporary
// file.
package excel
