// Package errors provides unified error types for RADIUS authentication and accounting.
// It follows Go's error handling philosophy with typed errors for metrics and logging.
package errors

import (
	"errors"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/app"
)

// RadiusError is the base interface for all RADIUS-related errors.
// It provides methods for metrics tracking and error categorization.
type RadiusError interface {
	error
	// MetricsKey returns the metrics key for this error type
	MetricsKey() string
	// Stage returns the processing stage where the error occurred
	Stage() string
}

// AuthError represents a RADIUS authentication error with metrics support.
// Use this type for all authentication-related errors to ensure consistent
// error handling and metrics reporting.
type AuthError struct {
	MetricsType string // Metrics key for Prometheus/monitoring
	Message     string // Human-readable error message
	ErrorStage  string // Pipeline stage where error occurred (optional)
	Cause       error  // Underlying error (optional)
}

func (e *AuthError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// MetricsKey implements RadiusError interface
func (e *AuthError) MetricsKey() string {
	return e.MetricsType
}

// Stage implements RadiusError interface
func (e *AuthError) Stage() string {
	return e.ErrorStage
}

// Unwrap returns the underlying error for errors.Is/As support
func (e *AuthError) Unwrap() error {
	return e.Cause
}

// AcctError represents a RADIUS accounting error with metrics support.
// Use this type for all accounting-related errors.
type AcctError struct {
	MetricsType string // Metrics key for Prometheus/monitoring
	Message     string // Human-readable error message
	ErrorStage  string // Processing stage where error occurred (optional)
	Cause       error  // Underlying error (optional)
}

func (e *AcctError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// MetricsKey implements RadiusError interface
func (e *AcctError) MetricsKey() string {
	return e.MetricsType
}

// Stage implements RadiusError interface
func (e *AcctError) Stage() string {
	return e.ErrorStage
}

// Unwrap returns the underlying error for errors.Is/As support
func (e *AcctError) Unwrap() error {
	return e.Cause
}

// AuthResult represents the outcome of an authentication pipeline stage.
// It provides a structured way to communicate results without using panic.
type AuthResult struct {
	Success     bool        // Whether the stage completed successfully
	Err         error       // Error if Success is false
	ShouldStop  bool        // Whether pipeline execution should stop (success or handled error)
	Response    interface{} // Optional response data to pass to next stage
	MetricsType string      // Metrics key for tracking
}

// NewAuthResult creates a successful result
func NewAuthResult() *AuthResult {
	return &AuthResult{Success: true}
}

// NewAuthResultWithStop creates a successful result that stops the pipeline
func NewAuthResultWithStop() *AuthResult {
	return &AuthResult{Success: true, ShouldStop: true}
}

// NewAuthErrorResult creates a failed result from an error
func NewAuthErrorResult(err error) *AuthResult {
	result := &AuthResult{Success: false, Err: err}
	if authErr, ok := err.(*AuthError); ok {
		result.MetricsType = authErr.MetricsType
	}
	return result
}

// NewAuthError creates an authentication error.
// Parameters:
//   - metricsType: The metrics key for monitoring (e.g., app.MetricsRadiusRejectNotExists)
//   - message: Human-readable error message
func NewAuthError(metricsType string, message string) error {
	return &AuthError{
		MetricsType: metricsType,
		Message:     message,
	}
}

// NewAuthErrorWithStage creates an authentication error with stage information.
// This is useful for tracking where in the pipeline an error occurred.
func NewAuthErrorWithStage(metricsType, message, stage string) error {
	return &AuthError{
		MetricsType: metricsType,
		Message:     message,
		ErrorStage:  stage,
	}
}

// NewAuthErrorWithCause creates an authentication error wrapping an underlying cause.
func NewAuthErrorWithCause(metricsType, message string, cause error) error {
	return &AuthError{
		MetricsType: metricsType,
		Message:     message,
		Cause:       cause,
	}
}

// NewAcctError creates an accounting error.
// Parameters:
//   - metricsType: The metrics key for monitoring (e.g., app.MetricsRadiusAcctDrop)
//   - message: Human-readable error message
func NewAcctError(metricsType string, message string) error {
	return &AcctError{
		MetricsType: metricsType,
		Message:     message,
	}
}

// NewAcctErrorWithStage creates an accounting error with stage information.
func NewAcctErrorWithStage(metricsType, message, stage string) error {
	return &AcctError{
		MetricsType: metricsType,
		Message:     message,
		ErrorStage:  stage,
	}
}

// NewAcctErrorWithCause creates an accounting error wrapping an underlying cause.
func NewAcctErrorWithCause(metricsType, message string, cause error) error {
	return &AcctError{
		MetricsType: metricsType,
		Message:     message,
		Cause:       cause,
	}
}

// IsAuthError checks whether the error is an authentication error
func IsAuthError(err error) bool {
	var authErr *AuthError
	return errors.As(err, &authErr)
}

// GetAuthError retrieves authentication error details using errors.As
func GetAuthError(err error) (*AuthError, bool) {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr, true
	}
	return nil, false
}

// IsAcctError checks whether the error is an accounting error
func IsAcctError(err error) bool {
	var acctErr *AcctError
	return errors.As(err, &acctErr)
}

// GetAcctError retrieves accounting error details using errors.As
func GetAcctError(err error) (*AcctError, bool) {
	var acctErr *AcctError
	if errors.As(err, &acctErr) {
		return acctErr, true
	}
	return nil, false
}

// GetRadiusError retrieves any RadiusError (AuthError or AcctError)
func GetRadiusError(err error) (RadiusError, bool) {
	if authErr, ok := GetAuthError(err); ok {
		return authErr, true
	}
	if acctErr, ok := GetAcctError(err); ok {
		return acctErr, true
	}
	return nil, false
}

// Common auth error constructors for convenience

// NewUserNotExistsError creates an error for non-existent users
func NewUserNotExistsError() error {
	return NewAuthError(app.MetricsRadiusRejectNotExists, "user not exists")
}

// NewUserDisabledError creates an error for disabled user accounts
func NewUserDisabledError() error {
	return NewAuthError(app.MetricsRadiusRejectDisable, "user status is disabled")
}

// NewUserExpiredError creates an error for expired user accounts
func NewUserExpiredError() error {
	return NewAuthError(app.MetricsRadiusRejectExpire, "user expired")
}

// NewPasswordMismatchError creates an error for password validation failures
func NewPasswordMismatchError() error {
	return NewAuthError(app.MetricsRadiusRejectPasswdError, "password mismatch")
}

// NewOnlineLimitError creates an error when online session limit is exceeded
func NewOnlineLimitError(message string) error {
	return NewAuthError(app.MetricsRadiusRejectLimit, message)
}

// NewMacBindError creates an error for MAC address binding failures
func NewMacBindError() error {
	return NewAuthError(app.MetricsRadiusRejectBindError, "mac address binding failed")
}

// NewVlanBindError creates an error for VLAN binding failures
func NewVlanBindError() error {
	return NewAuthError(app.MetricsRadiusRejectBindError, "vlan binding failed")
}

// NewUnauthorizedNasError creates an error for unauthorized NAS access
func NewUnauthorizedNasError(ip, identifier string, err error) error {
	return NewAuthErrorWithCause(app.MetricsRadiusRejectUnauthorized,
		fmt.Sprintf("unauthorized access to device, Ip=%s, Identifier=%s",
			ip, identifier), err)
}

// NewUsernameEmptyError creates an error for empty username
func NewUsernameEmptyError() error {
	return NewAuthError(app.MetricsRadiusRejectNotExists, "username is empty")
}

// Common accounting error constructors

// NewAcctDropError creates an accounting drop error
func NewAcctDropError(message string) error {
	return NewAcctError(app.MetricsRadiusAcctDrop, message)
}

// NewAcctNasNotFoundError creates an error when NAS is not found for accounting
func NewAcctNasNotFoundError(ip, identifier string) error {
	return NewAcctError(app.MetricsRadiusAcctDrop,
		fmt.Sprintf("NAS not found, Ip=%s, Identifier=%s", ip, identifier))
}

// NewAcctUsernameEmptyError creates an error for empty username in accounting
func NewAcctUsernameEmptyError() error {
	return NewAcctError(app.MetricsRadiusAcctDrop, "username is empty")
}

// WrapAuthError converts a general error into an AuthError
// If the error is already an AuthError, it returns it unchanged
func WrapAuthError(metricsType string, err error) error {
	if err == nil {
		return nil
	}
	if IsAuthError(err) {
		return err
	}
	return NewAuthErrorWithCause(metricsType, err.Error(), err)
}

// WrapAcctError converts a general error into an AcctError
// If the error is already an AcctError, it returns it unchanged
func WrapAcctError(metricsType string, err error) error {
	if err == nil {
		return nil
	}
	if IsAcctError(err) {
		return err
	}
	return NewAcctErrorWithCause(metricsType, err.Error(), err)
}

// WrapError converts a general error into an AuthError (deprecated, use WrapAuthError)
// Kept for backward compatibility
func WrapError(metricsType string, err error) error {
	return WrapAuthError(metricsType, err)
}

// NewError creates a generic non-typed error
func NewError(message string) error {
	return errors.New(message)
}

// JoinErrors combines multiple errors into a single error.
// Useful for collecting errors from multiple guards.
func JoinErrors(errs ...error) error {
	var nonNilErrs []error
	for _, err := range errs {
		if err != nil {
			nonNilErrs = append(nonNilErrs, err)
		}
	}
	if len(nonNilErrs) == 0 {
		return nil
	}
	if len(nonNilErrs) == 1 {
		return nonNilErrs[0]
	}
	return errors.Join(nonNilErrs...)
}
