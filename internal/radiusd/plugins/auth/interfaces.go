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

// Guard handles authentication errors uniformly (e.g., reject delay, blacklist)
type Guard interface {
	// Name returns the guard name
	Name() string

	// OnError is called when an error occurs during authentication; it can return a new error to abort the flow
	OnError(ctx context.Context, authCtx *AuthContext, stage string, err error) error
}
