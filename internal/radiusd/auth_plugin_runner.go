package radiusd

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"layeh.com/radius"
)

// authPluginOptions defines optional settings for authentication plugins
type authPluginOptions struct {
	skipPasswordValidation bool
}

// AuthPluginOption defines an option function for authentication plugins
type AuthPluginOption func(*authPluginOptions)

// SkipPasswordValidation skips password validation (used when authentication is handled elsewhere, e.g., EAP)
func SkipPasswordValidation() AuthPluginOption {
	return func(opts *authPluginOptions) {
		opts.skipPasswordValidation = true
	}
}

// AuthenticateUserWithPlugins uses the plugin system to authenticate a user
func (s *AuthService) AuthenticateUserWithPlugins(
	ctx context.Context,
	r *radius.Request,
	response *radius.Packet,
	user *domain.RadiusUser,
	nas *domain.NetNas,
	vendorReq *vendorparsers.VendorRequest,
	isMacAuth bool,
	opts ...AuthPluginOption,
) error {
	// Parse optional parameters
	options := &authPluginOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	// Create the authentication context
	authCtx := &auth.AuthContext{
		Request:       r,
		Response:      response,
		User:          user,
		Nas:           nas,
		VendorRequest: vendorReq,
		IsMacAuth:     isMacAuth,
		Metadata: map[string]interface{}{
			"profile_cache": s.AppContext().ProfileCache(), // Add profile cache for dynamic attribute resolution
		},
	}

	var password string
	var err error

	// 1. Perform password validation via plugins
	if !isMacAuth && !options.skipPasswordValidation {
		password, err = s.GetLocalPassword(user, isMacAuth)
		if err != nil {
			return errors.WrapError("radus_reject_passwd_error", err)
		}

		if err := s.validatePasswordWithPlugins(ctx, authCtx, password); err != nil {
			return err
		}
	}

	// 2. Perform profile checks via plugins
	if !isMacAuth {
		if err := s.checkPoliciesWithPlugins(ctx, authCtx); err != nil {
			return err
		}
	}

	return nil
}

// validatePasswordWithPlugins uses password validator plugins
func (s *AuthService) validatePasswordWithPlugins(
	ctx context.Context,
	authCtx *auth.AuthContext,
	password string,
) error {
	// Get all registered password validators
	validators := registry.GetPasswordValidators()

	// Iterate over validators to find one that can handle the current request
	for _, validator := range validators {
		if validator.CanHandle(authCtx) {
			return validator.Validate(ctx, authCtx, password)
		}
	}

	// Return an error if no suitable validator is found
	return errors.NewAuthError("radus_reject_other", "no suitable password validator found")
}

// checkPoliciesWithPlugins uses profile checker plugins
func (s *AuthService) checkPoliciesWithPlugins(
	ctx context.Context,
	authCtx *auth.AuthContext,
) error {
	// Get all registered profile checkers (already sorted by order)
	checkers := registry.GetPolicyCheckers()

	// Execute all profile checkers in order
	for _, checker := range checkers {
		if err := checker.Check(ctx, authCtx); err != nil {
			return err
		}
	}

	return nil
}
