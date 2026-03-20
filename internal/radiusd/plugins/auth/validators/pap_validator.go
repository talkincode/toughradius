package validators

import (
	"context"
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius/rfc2865"
)

// PAPValidator handles PAP password validation
type PAPValidator struct{}

func (v *PAPValidator) Name() string {
	return "pap"
}

func (v *PAPValidator) CanHandle(authCtx *auth.AuthContext) bool {
	password := rfc2865.UserPassword_GetString(authCtx.Request.Packet)
	return strings.TrimSpace(password) != ""
}

func (v *PAPValidator) Validate(ctx context.Context, authCtx *auth.AuthContext, password string) error {
	requestPassword := rfc2865.UserPassword_GetString(authCtx.Request.Packet)

	// LDAP takes precedence for PAP when enabled; local password comparison is used as fallback.
	handled, err := validatePasswordWithLDAP(authCtx, requestPassword)
	if handled {
		return err
	}

	if strings.TrimSpace(requestPassword) != password {
		return errors.NewPasswordMismatchError()
	}

	return nil
}
