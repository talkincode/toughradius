/*
 * Copyright (c) 2024-2025 TalkingCode
 * Licensed under the MIT License. See LICENSE file in the project root for details.
 */

// Package common provides general-purpose utility functions for the ToughRADIUS server.
//
// This package includes essential helpers for:
//   - UUID generation (both string and int64 snowflake IDs)
//   - Cryptographic hashing with salt
//   - Type checking and validation (empty values, slices, etc.)
//   - JSON serialization utilities
//   - File system operations
//   - Conditional value selection
//
// These utilities are used throughout the codebase to avoid code duplication
// and maintain consistency in common operations.
package common

import (
	"bytes"
	"crypto/rand"
	sha256_ "crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	EmptyList      []interface{}
	EmptyData      map[string]interface{}
	ShangHaiLOC, _ = time.LoadLocation("Asia/Shanghai")
)

const (
	NA       = "N/A"
	ENABLED  = "enabled"
	DISABLED = "disabled"
)

// defaultSecretSalt is used only for development/testing when env var is not set
const defaultSecretSalt = "toughradius-dev-salt-change-me" //nolint:gosec // G101: this is a default dev value, not a credential

// GetSecretSalt returns the secret salt from environment variable TOUGHRADIUS_SECRET_SALT.
// Falls back to a default value for development only.
// IMPORTANT: Always set TOUGHRADIUS_SECRET_SALT in production!
func GetSecretSalt() string {
	if salt := os.Getenv("TOUGHRADIUS_SECRET_SALT"); salt != "" {
		return salt
	}
	return defaultSecretSalt
}

// FileExists checks whether a file exists at the specified path.
// It returns false if the path points to a directory or if an error occurs.
//
// Parameters:
//   - file: Absolute or relative path to check
//
// Returns:
//   - bool: true if path exists and is a file, false otherwise
//
// Example:
//
//	if common.FileExists("/etc/toughradius/config.yml") {
//	    // Load config
//	}
func FileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

// DirExists checks whether a directory exists at the specified path.
// It returns false if the path points to a file or if an error occurs.
//
// Parameters:
//   - file: Absolute or relative path to check
//
// Returns:
//   - bool: true if path exists and is a directory, false otherwise
func DirExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && info.IsDir()
}

// Must panics with a stack trace if the provided error is not nil.
// This is intended for initialization code where errors are unrecoverable.
//
// Parameters:
//   - err: Error to check (panics if non-nil)
//
// Example:
//
//	config, err := loadConfig()
//	common.Must(err)  // Panic on config load failure
func Must(err error) {
	if err != nil {
		panic(errors.WithStack(err))
	}
}

// Must2 returns the provided value if error is nil, otherwise panics.
// This is a convenience wrapper for functions that return (value, error).
//
// Parameters:
//   - v: Value to return if no error
//   - err: Error to check (panics if non-nil)
//
// Returns:
//   - interface{}: The input value v
//
// Example:
//
//	config := common.Must2(loadConfig()).(*Config)
func Must2(v interface{}, err error) interface{} {
	Must(err)
	return v
}

// UUID generates a time-based UUID string with cryptographic randomness.
// The format is: {unix32bits}-{rand}-{rand}-{rand}-{rand}-{rand}
//
// Returns:
//   - string: Hexadecimal UUID (e.g., "5f3a2b1c-1234-5678-90ab-cdef01234567")
//
// Example:
//
//	sessionID := common.UUID()
//	log.Printf("Session ID: %s", sessionID)
func UUID() string {
	unix32bits := uint32(time.Now().UTC().Unix()) //nolint:gosec // G115: Unix timestamp fits in uint32 until 2106
	buff := make([]byte, 12)
	numRead, err := rand.Read(buff)
	if numRead != len(buff) || err != nil {
		Must(err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x-%x", unix32bits, buff[0:2], buff[2:4], buff[4:6], buff[6:8], buff[8:])
}

var snowflakeNode, _ = snowflake.NewNode(int64(mathrand.Intn(1000))) //nolint:gosec // G404: weak random is acceptable for node ID

// UUIDint64 generates a unique 64-bit integer ID using the Snowflake algorithm.
// This is suitable for distributed systems where sortable, collision-resistant IDs are needed.
//
// The generated ID is based on:
//   - Timestamp (millisecond precision)
//   - Node ID (randomly initialized at startup)
//   - Sequence number (incremented for IDs within same millisecond)
//
// Returns:
//   - int64: Unique snowflake ID (always positive)
//
// Example:
//
//	accountingID := common.UUIDint64()
//	db.Model(&RadiusAccounting{ID: accountingID}).Create(...)
func UUIDint64() int64 {
	return snowflakeNode.Generate().Int64()
}

// Sha256HashWithSalt computes a SHA-256 hash of the source string combined with a salt.
// This is used for password hashing and other cryptographic operations.
//
// Parameters:
//   - src: Plain text to hash
//   - salt: Salt value (should be unique per application/user)
//
// Returns:
//   - string: Hexadecimal SHA-256 hash (64 characters)
//
// Example:
//
//	hashed := common.Sha256HashWithSalt(password, common.GetSecretSalt())
//	if user.Password == hashed {
//	    // Authentication success
//	}
func Sha256HashWithSalt(src string, salt string) string {
	h := sha256_.New()
	h.Write([]byte(src))
	h.Write([]byte(salt))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

// HashPassword hashes password with bcrypt and a secret salt (pepper).
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password+GetSecretSalt()), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks password against a bcrypt hash.
func VerifyPassword(password, hashedPassword string) bool {
	if !IsBcryptHash(hashedPassword) {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password+GetSecretSalt())) == nil
}

// IsBcryptHash checks whether the given hash uses bcrypt format.
func IsBcryptHash(hashedPassword string) bool {
	return strings.HasPrefix(hashedPassword, "$2a$") ||
		strings.HasPrefix(hashedPassword, "$2b$") ||
		strings.HasPrefix(hashedPassword, "$2y$")
}

// ConstantTimeEquals compares two strings in constant time.
func ConstantTimeEquals(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// InSlice checks whether a string value exists in a slice of strings.
// Comparison is case-sensitive and uses exact matching.
//
// Parameters:
//   - v: String value to search for
//   - sl: Slice of strings to search in
//
// Returns:
//   - bool: true if v is found in sl, false otherwise
//
// Example:
//
//	if common.InSlice("admin", user.Roles) {
//	    // User has admin role
//	}
func InSlice(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

// If implements a ternary operator pattern, returning one of two values based on a condition.
// This is useful for inline conditional assignments.
//
// Parameters:
//   - condition: Boolean expression to evaluate
//   - trueVal: Value to return if condition is true
//   - falseVal: Value to return if condition is false
//
// Returns:
//   - interface{}: Either trueVal or falseVal (requires type assertion)
//
// Example:
//
//	logLevel := common.If(debug, "DEBUG", "INFO").(string)
//	port := common.If(useSSL, 443, 80).(int)
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

// IfEmptyStr returns a default value if the source string is empty.
// This is a type-safe alternative to If() for string values.
//
// Parameters:
//   - src: String to check
//   - defval: Default value to return if src is empty
//
// Returns:
//   - string: src if non-empty, otherwise defval
//
// Example:
//
//	hostname := common.IfEmptyStr(config.Hostname, "localhost")
//	port := common.IfEmptyStr(config.Port, "1812")
func IfEmptyStr(src string, defval string) string {
	if src == "" {
		return defval
	}
	return src
}

// IsEmpty checks whether a value is considered "empty" using Go semantics.
// This uses reflection to handle different types uniformly.
//
// A value is considered empty if:
//   - Numeric types (int, float, uint): zero value (0)
//   - bool: false
//   - string: empty string ("")
//   - array/slice/map: nil or length == 0
//   - pointer/interface: nil or referenced value is empty
//   - time.Time: IsZero() returns true
//
// Parameters:
//   - value: Value of any type to check
//
// Returns:
//   - bool: true if value is empty by Go conventions
//
// Example:
//
//	if common.IsEmpty(user.Email) {
//	    return errors.New("email required")
//	}
//	if !common.IsEmpty(config.Features) {
//	    // Process features
//	}
func IsEmpty(value interface{}) bool {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Invalid:
		return true
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return IsEmpty(v.Elem().Interface())
	case reflect.Struct:
		v, ok := value.(time.Time)
		if ok && v.IsZero() {
			return true
		}
	}

	return false
}

// IsEmptyOrNA checks whether a string value is empty or the special "N/A" marker.
// The check is performed after trimming whitespace.
//
// Parameters:
//   - val: String value to check
//
// Returns:
//   - bool: true if val is empty or equals "N/A" (after trimming)
//
// Example:
//
//	if common.IsEmptyOrNA(user.Phone) {
//	    // Skip phone validation
//	}
func IsEmptyOrNA(val string) bool {
	val = strings.TrimSpace(val)
	return val == "" || val == NA
}

// IsNotEmptyAndNA checks whether a string value is non-empty and not the special "N/A" marker.
// This is the logical inverse of IsEmptyOrNA.
//
// Parameters:
//   - val: String value to check
//
// Returns:
//   - bool: true if val is non-empty and not "N/A" (after trimming)
//
// Example:
//
//	if common.IsNotEmptyAndNA(user.Description) {
//	    // Process description field
//	}
func IsNotEmptyAndNA(val string) bool {
	val = strings.TrimSpace(val)
	return strings.TrimSpace(val) != "" && val != NA
}

// JsonMarshal is a thin wrapper around json.Marshal for consistency.
// It serializes a Go value into JSON bytes.
//
// Parameters:
//   - v: Value to marshal (any JSON-serializable type)
//
// Returns:
//   - []byte: JSON-encoded bytes
//   - error: Marshaling error (e.g., unsupported type, cyclic reference)
//
// Example:
//
//	data, err := common.JsonMarshal(config)
//	if err != nil {
//	    return err
//	}
//	os.WriteFile("config.json", data, 0644)
func JsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// JsonUnmarshal is a thin wrapper around json.Unmarshal for consistency.
// It deserializes JSON bytes into a Go value.
//
// Parameters:
//   - data: JSON-encoded bytes
//   - v: Pointer to value to unmarshal into
//
// Returns:
//   - error: Unmarshaling error (e.g., syntax error, type mismatch)
//
// Example:
//
//	var config AppConfig
//	if err := common.JsonUnmarshal(data, &config); err != nil {
//	    return fmt.Errorf("invalid config: %w", err)
//	}
func JsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ToJson converts a Go value to a pretty-printed JSON string.
// Indentation uses 2 spaces. This is intended for debugging and logging.
//
// Parameters:
//   - v: Value to convert to JSON
//
// Returns:
//   - string: Indented JSON string (errors are silently ignored, returns "")
//
// Example:
//
//	zap.L().Debug("user data", zap.String("json", common.ToJson(user)))
//	fmt.Println(common.ToJson(config))  // Pretty-print config
func ToJson(v interface{}) string {
	bs, _ := json.MarshalIndent(v, "", "  ")
	return string(bs)
}

// TrimBytes removes UTF-8 BOM (Byte Order Mark) from byte slices.
// This is useful when reading files that may have been edited on Windows.
//
// Parameters:
//   - src: Byte slice to clean (may contain BOM: 0xEF 0xBB 0xBF)
//
// Returns:
//   - []byte: Cleaned byte slice with BOM removed
//
// Example:
//
//	data, _ := os.ReadFile("config.json")
//	data = common.TrimBytes(data)  // Remove BOM if present
//	json.Unmarshal(data, &config)
func TrimBytes(src []byte) []byte {
	s := bytes.ReplaceAll(src, []byte("\xef\xbb\xbf"), []byte(""))
	return s
}
