package radiusd

import (
	"context"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/ldapauth"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// fakeLDAPSettings is a configurable ldapauth.Settings for the backend tests.
type fakeLDAPSettings struct {
	strs  map[string]string
	ints  map[string]int64
	bools map[string]bool
}

func (f fakeLDAPSettings) GetSettingsStringValue(category, key string) string {
	return f.strs[category+"."+key]
}
func (f fakeLDAPSettings) GetSettingsInt64Value(category, key string) int64 {
	return f.ints[category+"."+key]
}
func (f fakeLDAPSettings) GetSettingsBoolValue(category, key string) bool {
	return f.bools[category+"."+key]
}

// enabledTemplateSettings returns settings with the LDAP backend enabled in a
// template bind mode that is complete enough for ldapauth.Validate to pass, so
// a Verify reaches the (unreachable) dial step and fails as Unavailable.
func enabledTemplateSettings() fakeLDAPSettings {
	return fakeLDAPSettings{
		strs: map[string]string{
			"ldap.ServerURL":      "ldap://127.0.0.1:3899",
			"ldap.BindMode":       "template",
			"ldap.BindDNTemplate": "uid=%s,ou=people,dc=example,dc=com",
		},
		bools: map[string]bool{"ldap.Enabled": true},
	}
}

func authErrorMetric(t *testing.T, err error) string {
	t.Helper()
	authErr, ok := radiuserrors.GetAuthError(err)
	if !ok {
		t.Fatalf("expected an AuthError, got %T: %v", err, err)
	}
	return authErr.MetricsType
}

func TestLDAPBackend_Active(t *testing.T) {
	// nil settings -> inactive (never panics).
	if newLDAPCredentialBackend(nil).Active() {
		t.Fatal("nil settings backend must be inactive")
	}
	// Enabled=false -> inactive.
	off := newLDAPCredentialBackend(fakeLDAPSettings{bools: map[string]bool{"ldap.Enabled": false}})
	if off.Active() {
		t.Fatal("disabled backend must be inactive")
	}
	// Enabled=true -> active.
	on := newLDAPCredentialBackend(enabledTemplateSettings())
	if !on.Active() {
		t.Fatal("enabled backend must be active")
	}
}

func TestLDAPBackend_GetPassword(t *testing.T) {
	user := &domain.RadiusUser{Username: "alice", Password: "localpw", MacAddr: "00:11:22:33:44:55"}

	// Inactive backend mirrors the default provider: local password for
	// interactive auth, MAC for MAC auth.
	off := newLDAPCredentialBackend(fakeLDAPSettings{})
	if got, err := off.GetPassword(user, false); err != nil || got != "localpw" {
		t.Fatalf("inactive non-mac: got (%q,%v), want (localpw,nil)", got, err)
	}
	if got, err := off.GetPassword(user, true); err != nil || got != "00:11:22:33:44:55" {
		t.Fatalf("inactive mac: got (%q,%v), want (mac,nil)", got, err)
	}

	on := newLDAPCredentialBackend(enabledTemplateSettings())
	// THE safety guard: an active backend must NOT disclose a local password for
	// interactive auth, so challenge-response EAP methods that call GetPassword
	// reject instead of comparing against a (now non-authoritative) local secret.
	got, err := on.GetPassword(user, false)
	if err == nil {
		t.Fatal("active backend must refuse to provide a local password for interactive auth")
	}
	if got != "" {
		t.Fatalf("active backend must return an empty password, got %q", got)
	}
	if metric := authErrorMetric(t, err); metric != app.MetricsRadiusRejectLdapError {
		t.Fatalf("active GetPassword metric = %q, want %q", metric, app.MetricsRadiusRejectLdapError)
	}
	// MAC auth is independent of LDAP and must still work when active.
	if mac, err := on.GetPassword(user, true); err != nil || mac != "00:11:22:33:44:55" {
		t.Fatalf("active mac: got (%q,%v), want (mac,nil)", mac, err)
	}
}

func TestMapLDAPVerifyError(t *testing.T) {
	cases := []struct {
		name   string
		in     error
		metric string // "" means expect nil
	}{
		{"success", nil, ""},
		{"invalid_credentials", ldapauth.ErrInvalidCredentials, app.MetricsRadiusRejectPasswdError},
		{"user_not_found", ldapauth.ErrUserNotFound, app.MetricsRadiusRejectPasswdError},
		{"empty_credential", ldapauth.ErrEmptyCredential, app.MetricsRadiusRejectPasswdError},
		{"unavailable", ldapauth.ErrUnavailable, app.MetricsRadiusRejectLdapError},
		{"not_configured", ldapauth.ErrNotConfigured, app.MetricsRadiusRejectLdapError},
		{"disabled", ldapauth.ErrDisabled, app.MetricsRadiusRejectLdapError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapLDAPVerifyError(tc.in)
			if tc.metric == "" {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if metric := authErrorMetric(t, got); metric != tc.metric {
				t.Fatalf("metric = %q, want %q", metric, tc.metric)
			}
		})
	}
}

func TestLDAPBackend_VerifyCleartext_Mapping(t *testing.T) {
	// Enabled but not configured (empty ServerURL): Verify returns
	// ErrNotConfigured before dialing -> backend error metric.
	notConfigured := newLDAPCredentialBackend(fakeLDAPSettings{
		bools: map[string]bool{"ldap.Enabled": true},
	})
	err := notConfigured.VerifyCleartext(context.Background(), "alice", "pw")
	if err == nil {
		t.Fatal("expected an error for an unconfigured backend")
	}
	if metric := authErrorMetric(t, err); metric != app.MetricsRadiusRejectLdapError {
		t.Fatalf("not-configured metric = %q, want %q", metric, app.MetricsRadiusRejectLdapError)
	}

	// Configured + enabled but empty password: Verify rejects before dialing
	// (RFC 4513 §5.1.2 unauthenticated-bind defense) -> credential metric.
	configured := newLDAPCredentialBackend(enabledTemplateSettings())
	err = configured.VerifyCleartext(context.Background(), "alice", "")
	if err == nil {
		t.Fatal("expected an error for an empty password")
	}
	if metric := authErrorMetric(t, err); metric != app.MetricsRadiusRejectPasswdError {
		t.Fatalf("empty-password metric = %q, want %q", metric, app.MetricsRadiusRejectPasswdError)
	}
}

func TestLDAPBackend_VerifyRequestPAP(t *testing.T) {
	secret := []byte("s3cr3t")
	makeAuthCtx := func(setPassword bool, user *domain.RadiusUser) *auth.AuthContext {
		packet := radius.New(radius.CodeAccessRequest, secret)
		if setPassword {
			if err := rfc2865.UserPassword_SetString(packet, "bindpw"); err != nil {
				t.Fatalf("set password: %v", err)
			}
		}
		return &auth.AuthContext{
			Request: &radius.Request{Packet: packet},
			User:    user,
		}
	}

	backend := newLDAPCredentialBackend(enabledTemplateSettings())
	user := &domain.RadiusUser{Username: "alice"}

	// No User-Password (a CHAP/MS-CHAP request): rejected with a diagnostic
	// backend error WITHOUT attempting a bind.
	err := backend.verifyRequestPAP(context.Background(), makeAuthCtx(false, user))
	if err == nil {
		t.Fatal("a non-PAP request must be rejected")
	}
	if metric := authErrorMetric(t, err); metric != app.MetricsRadiusRejectLdapError {
		t.Fatalf("non-PAP metric = %q, want %q", metric, app.MetricsRadiusRejectLdapError)
	}

	// nil user record: rejected with a diagnostic backend error.
	if err := backend.verifyRequestPAP(context.Background(), makeAuthCtx(true, nil)); err == nil {
		t.Fatal("a missing user record must be rejected")
	}

	// A PAP request routes to the bind path; with an unreachable directory it
	// surfaces as an unavailable backend (proving it got past the no-password
	// guard into VerifyCleartext rather than being rejected as non-PAP).
	err = backend.verifyRequestPAP(context.Background(), makeAuthCtx(true, user))
	if err == nil {
		t.Fatal("expected the bind attempt to fail against an unreachable directory")
	}
	if metric := authErrorMetric(t, err); metric != app.MetricsRadiusRejectLdapError {
		t.Fatalf("PAP-bind metric = %q, want %q", metric, app.MetricsRadiusRejectLdapError)
	}
}
