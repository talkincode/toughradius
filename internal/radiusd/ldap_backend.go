package radiusd

import (
	"context"
	stderrs "errors"
	"strings"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/ldapauth"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"layeh.com/radius/rfc2865"
)

// ldapCredentialBackend adapts the standalone internal/ldapauth verifier (M14.1)
// to the RADIUS authentication pipeline as a PAP-family password backend
// (M14.2, feature TR-F025).
//
// An LDAP/AD directory verifies a password by performing a Bind with the user's
// cleartext password and never discloses the stored password. The server can
// therefore only authenticate methods where it already holds that cleartext:
// bare PAP and the inner PAP of EAP-TTLS (RFC 5281 §11.2.5). Challenge/response
// methods (CHAP, MS-CHAP, MS-CHAPv2, EAP-MD5, PEAP-MSCHAPv2) cannot be backed by
// a Bind because they need the stored secret to recompute the response; when the
// backend is active they are rejected with a diagnostic reason rather than
// silently compared against the (now authoritative-elsewhere) local password,
// which could otherwise be empty and falsely accept a response computed over an
// empty secret.
//
// The local RadiusUser record is still required and supplies authorization
// (profile, rate limits, expiry, concurrency, address pools); the directory only
// replaces the password check. Enrolling users that exist solely in the
// directory is out of scope.
//
// The backend reads ldap.* settings on every attempt so a runtime configuration
// change (enable/disable, server URL, bind mode) takes effect immediately;
// connection reuse and richer observability are deferred to M14.3.
type ldapCredentialBackend struct {
	settings ldapauth.Settings
}

// Compile-time guarantees that the backend can stand in for the default EAP
// password provider and additionally expose the bind-based verifier capability
// the PAP seams use.
var (
	_ eap.PasswordProvider   = (*ldapCredentialBackend)(nil)
	_ eap.CredentialVerifier = (*ldapCredentialBackend)(nil)
)

// newLDAPCredentialBackend builds a backend over the given settings accessor.
// settings is typically the application context (which satisfies
// ldapauth.Settings); a nil accessor yields an inactive backend that defers to
// local password handling.
func newLDAPCredentialBackend(settings ldapauth.Settings) *ldapCredentialBackend {
	return &ldapCredentialBackend{settings: settings}
}

// ldapBackend returns a backend bound to the service's application context.
func (s *AuthService) ldapBackend() *ldapCredentialBackend {
	return newLDAPCredentialBackend(s.AppContext())
}

// config loads the current ldap.* configuration, or a zero Config (Enabled
// false) when no settings accessor is available.
func (b *ldapCredentialBackend) config() ldapauth.Config {
	if b == nil || b.settings == nil {
		return ldapauth.Config{}
	}
	return ldapauth.LoadConfig(b.settings)
}

// Active reports whether the LDAP backend is enabled. When it is, only
// PAP-family authentication can succeed.
func (b *ldapCredentialBackend) Active() bool {
	return b.config().Enabled
}

// VerifyCleartext authenticates username with the presented cleartext password
// by binding against the directory. It returns nil on success, an AuthError
// tagged radus_reject_passwd_error for a credential rejection (wrong password,
// user not in the directory, empty credential), and an AuthError tagged
// radus_reject_ldap_error for a backend that is unreachable or misconfigured.
func (b *ldapCredentialBackend) VerifyCleartext(ctx context.Context, username, password string) error {
	return mapLDAPVerifyError(ldapauth.New(b.config()).Verify(ctx, username, password))
}

// mapLDAPVerifyError translates an internal/ldapauth result into a pipeline
// AuthError. Credential rejections (wrong password, user not found, empty
// credential) map to radus_reject_passwd_error; a disabled, misconfigured, or
// unreachable backend maps to radus_reject_ldap_error so an outage is not
// reported to the client as a wrong password.
func mapLDAPVerifyError(err error) error {
	switch {
	case err == nil:
		return nil
	case stderrs.Is(err, ldapauth.ErrInvalidCredentials),
		stderrs.Is(err, ldapauth.ErrUserNotFound),
		stderrs.Is(err, ldapauth.ErrEmptyCredential):
		return radiuserrors.NewAuthErrorWithCause(app.MetricsRadiusRejectPasswdError,
			"ldap authentication rejected", err)
	default:
		// ErrUnavailable / ErrNotConfigured / ErrDisabled / unknown: a server-side
		// condition, not a credential rejection.
		return radiuserrors.NewAuthErrorWithCause(app.MetricsRadiusRejectLdapError,
			"ldap backend unavailable", err)
	}
}

// GetPassword implements eap.PasswordProvider. When the backend is active it
// refuses to hand out a local password for interactive (non-MAC) authentication:
// the directory is the password authority and challenge/response EAP methods
// that call this to recompute their expected response must reject rather than
// compare against a local secret that no longer governs the account. PAP seams
// never reach this path when the backend is active because they call
// VerifyCleartext instead. MAC authentication is unaffected (it compares the
// MAC address, not a directory password) and local handling is preserved when
// the backend is inactive.
func (b *ldapCredentialBackend) GetPassword(user *domain.RadiusUser, isMacAuth bool) (string, error) {
	if !isMacAuth && b.Active() {
		return "", radiuserrors.NewAuthError(app.MetricsRadiusRejectLdapError,
			"ldap backend active: challenge-response methods (CHAP/MS-CHAP/EAP-MD5/PEAP-MSCHAPv2) "+
				"cannot be verified by LDAP bind; only PAP-family methods are supported")
	}
	if isMacAuth {
		if user.MacAddr != "" {
			return user.MacAddr, nil
		}
		return user.Username, nil
	}
	return user.Password, nil
}

// verifyRequestPAP authenticates a bare RADIUS Access-Request against the
// directory. It accepts only PAP (a non-empty User-Password attribute, RFC 2865
// §5.2); a request without a usable cleartext password is a challenge/response
// method (CHAP/MS-CHAP) or an empty password and is rejected with a diagnostic
// reason, because the directory cannot verify it by binding.
func (b *ldapCredentialBackend) verifyRequestPAP(ctx context.Context, authCtx *auth.AuthContext) error {
	presented := rfc2865.UserPassword_GetString(authCtx.Request.Packet)
	if strings.TrimSpace(presented) == "" {
		return radiuserrors.NewAuthError(app.MetricsRadiusRejectLdapError,
			"ldap backend active: no PAP password present; CHAP/MS-CHAP and empty passwords "+
				"cannot be verified by LDAP bind")
	}
	if authCtx.User == nil {
		return radiuserrors.NewAuthError(app.MetricsRadiusRejectLdapError,
			"ldap backend active: missing user record")
	}
	return b.VerifyCleartext(ctx, authCtx.User.Username, presented)
}
