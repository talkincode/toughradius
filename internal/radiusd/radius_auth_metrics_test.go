package radiusd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/pkg/metrics"
	"layeh.com/radius"
)

// metricsTestAppContext supplies a non-nil Config so SendReject's debug check
// does not dereference a nil pointer; all other accessors inherit the empty
// mock behavior.
type metricsTestAppContext struct {
	mockAppContext
}

func (m *metricsTestAppContext) Config() *config.AppConfig { return &config.AppConfig{} }

// captureResponseWriter records the last RADIUS packet written.
type captureResponseWriter struct {
	last *radius.Packet
}

func (w *captureResponseWriter) Write(p *radius.Packet) error {
	w.last = p
	return nil
}

// TestLogAndReject_CountsRejectByReason locks in the M14.3 observability fix:
// the bare PAP/CHAP reject path must increment the per-reason RADIUS metric so
// an LDAP-backed PAP rejection (radus_reject_ldap_error) is visible to operators
// when the directory is unreachable, and an unclassified error falls back to the
// auth-drop counter.
func TestLogAndReject_CountsRejectByReason(t *testing.T) {
	require.NoError(t, metrics.InitMetrics(""))

	svc := &AuthService{RadiusService: &RadiusService{appCtx: &metricsTestAppContext{}}}
	req := &radius.Request{Packet: radius.New(radius.CodeAccessRequest, []byte("secret"))}

	cases := []struct {
		name   string
		err    error
		metric string
	}{
		{"ldap_error", radiuserrors.NewAuthError(app.MetricsRadiusRejectLdapError, "ldap unavailable"), app.MetricsRadiusRejectLdapError},
		{"passwd_error", radiuserrors.NewAuthError(app.MetricsRadiusRejectPasswdError, "wrong password"), app.MetricsRadiusRejectPasswdError},
		{"unclassified_falls_back_to_auth_drop", errors.New("boom"), app.MetricsRadiusAuthDrop},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := app.GetRadiusMetrics(tc.metric)
			w := &captureResponseWriter{}

			svc.logAndReject(w, req, tc.err)

			assert.Equal(t, before+1, app.GetRadiusMetrics(tc.metric),
				"reject must increment %s", tc.metric)
			require.NotNil(t, w.last, "a reject response must be written")
			assert.Equal(t, radius.CodeAccessReject, w.last.Code)
		})
	}
}

// TestLogAndReject_LdapRejectNotCountedAsPasswordError guards the metric
// classification: an LDAP backend outage must not be tallied as a password
// rejection (which would corrupt wrong-password dashboards/alerts).
func TestLogAndReject_LdapRejectNotCountedAsPasswordError(t *testing.T) {
	require.NoError(t, metrics.InitMetrics(""))

	svc := &AuthService{RadiusService: &RadiusService{appCtx: &metricsTestAppContext{}}}
	req := &radius.Request{Packet: radius.New(radius.CodeAccessRequest, []byte("secret"))}

	passwdBefore := app.GetRadiusMetrics(app.MetricsRadiusRejectPasswdError)
	svc.logAndReject(&captureResponseWriter{},
		req, radiuserrors.NewAuthError(app.MetricsRadiusRejectLdapError, "ldap unavailable"))

	assert.Equal(t, int64(1), app.GetRadiusMetrics(app.MetricsRadiusRejectLdapError))
	assert.Equal(t, passwdBefore, app.GetRadiusMetrics(app.MetricsRadiusRejectPasswdError),
		"a backend outage must not increment the password-error counter")
}
