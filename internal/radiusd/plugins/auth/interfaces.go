package auth

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"layeh.com/radius"
)

// AuthContext represents the authentication context
type AuthContext struct {
	Request       *radius.Request
	Response      *radius.Packet
	User          *domain.RadiusUser
	Nas           *domain.NetNas
	VendorRequest interface{}
	IsMacAuth     bool                   // whether this is MAC authentication
	Metadata      map[string]interface{} // Additional metadata
}

// PasswordValidator defines the password validation interface
type PasswordValidator interface {
	// Name returns the validator name (pap, chap, mschap, eap-md5, etc.)
	Name() string

	// CanHandle determines whether this validator can handle the request
	CanHandle(ctx *AuthContext) bool

	// Validate performs password validation
	Validate(ctx context.Context, authCtx *AuthContext, password string) error
}

// PolicyChecker defines the profile check interface
type PolicyChecker interface {
	// Name returns the checker's name
	Name() string

	// Check executes the profile check
	Check(ctx context.Context, authCtx *AuthContext) error

	// Order returns the execution order (lower digits run first)
	Order() int
}

// ResponseEnhancer defines the response enhancement interface
type ResponseEnhancer interface {
	// Name returns the enhancer name
	Name() string

	// Enhance augments the response (e.g., add vendor attributes)
	Enhance(ctx context.Context, authCtx *AuthContext) error
}

// GuardAction indicates what action the caller should take after a guard processes an error
type GuardAction int

const (
	// GuardActionContinue indicates the error handling should continue to next guard
	GuardActionContinue GuardAction = iota
	// GuardActionStop indicates error handling should stop, use the returned error
	GuardActionStop
	// GuardActionSuppress indicates the error should be suppressed (treated as success)
	GuardActionSuppress
)

// GuardResult represents the result of a guard's error handling
type GuardResult struct {
	Action GuardAction // What action to take
	Err    error       // The error to use (may be modified, wrapped, or new)
}

// Guard handles authentication errors uniformly (e.g., reject delay, blacklist)
type Guard interface {
	// Name returns the guard name
	Name() string

	// OnError is called when an error occurs during authentication.
	// It can return a new error to abort the flow.
	// Deprecated: Use OnAuthError for more control over error handling flow.
	OnError(ctx context.Context, authCtx *AuthContext, stage string, err error) error

	// OnAuthError is called when an error occurs during authentication.
	// It provides more control over how errors are handled via GuardResult.
	// If not implemented (returns nil result), falls back to OnError behavior.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadlines
	//   - authCtx: Authentication context with request details
	//   - stage: The pipeline stage where the error occurred
	//   - err: The original error
	//
	// Returns:
	//   - *GuardResult: How to handle the error, or nil to use OnError fallback
	OnAuthError(ctx context.Context, authCtx *AuthContext, stage string, err error) *GuardResult
}

// BaseGuard provides a default implementation of Guard interface.
// Embed this in custom guards to get default behavior.
type BaseGuard struct{}

// OnError provides default no-op implementation
func (g *BaseGuard) OnError(ctx context.Context, authCtx *AuthContext, stage string, err error) error {
	return nil
}

// OnAuthError provides default implementation that falls back to OnError
func (g *BaseGuard) OnAuthError(ctx context.Context, authCtx *AuthContext, stage string, err error) *GuardResult {
	return nil // nil means use OnError fallback
}
