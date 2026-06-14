package radiusd

import (
	"errors"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/app"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/pkg/metrics"
	"layeh.com/radius"
)

// errAcctResponseWriter always fails the write so SendResponse takes its
// response-I/O drop path.
type errAcctResponseWriter struct{}

func (errAcctResponseWriter) Write(*radius.Packet) error { return errors.New("write failed") }

// TestAcctMetricsKeyForStage locks in the accounting failure taxonomy: each
// ingress stage maps to its own metric and any unknown stage falls back to the
// catch-all drop counter.
func TestAcctMetricsKeyForStage(t *testing.T) {
	cases := []struct {
		stage string
		want  string
	}{
		{"nas_lookup", app.MetricsRadiusAcctDropNas},
		{"validate_username", app.MetricsRadiusAcctDropUsername},
		{"verify_secret", app.MetricsRadiusAcctDropSecret},
		{"something_else", app.MetricsRadiusAcctDrop},
		{"", app.MetricsRadiusAcctDrop},
	}
	for _, tc := range cases {
		assert.Equalf(t, tc.want, acctMetricsKeyForStage(tc.stage),
			"stage %q must map to %s", tc.stage, tc.want)
	}
}

// TestLogAcctError_CountsDropByStage is the accounting analog of the auth-side
// M14.3 fix: a dropped Accounting-Request must increment the per-stage failure
// metric (so an unknown NAS is distinguishable from a missing username or a bad
// authenticator) instead of being only logged. It also guards the AuthError
// leak: the shared GetNas helper returns an auth-typed error carrying
// radus_reject_unauthorized, which must NOT bleed into the auth reject counter
// when an accounting packet is dropped.
func TestLogAcctError_CountsDropByStage(t *testing.T) {
	require.NoError(t, metrics.InitMetrics(""))

	acct := &AcctService{RadiusService: &RadiusService{}}

	cases := []struct {
		name   string
		stage  string
		err    error
		metric string
	}{
		{
			name:   "nas_lookup",
			stage:  "nas_lookup",
			err:    radiuserrors.NewUnauthorizedNasError("10.0.0.1", "nas-1", errors.New("record not found")),
			metric: app.MetricsRadiusAcctDropNas,
		},
		{
			name:   "validate_username",
			stage:  "validate_username",
			err:    radiuserrors.NewAcctUsernameEmptyError(),
			metric: app.MetricsRadiusAcctDropUsername,
		},
		{
			name:   "verify_secret",
			stage:  "verify_secret",
			err:    ErrSecretMismatch,
			metric: app.MetricsRadiusAcctDropSecret,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := app.GetRadiusMetrics(tc.metric)
			unauthorizedBefore := app.GetRadiusMetrics(app.MetricsRadiusRejectUnauthorized)

			acct.logAcctError(tc.stage, "10.0.0.1", "bob", tc.err)

			assert.Equal(t, before+1, app.GetRadiusMetrics(tc.metric),
				"stage %q must increment %s", tc.stage, tc.metric)
			assert.Equal(t, unauthorizedBefore, app.GetRadiusMetrics(app.MetricsRadiusRejectUnauthorized),
				"an accounting drop must not increment the auth reject counter")
		})
	}
}

// TestSubmitAcctTaskCountsDropOnOverload verifies that a back-pressure drop
// (worker pool saturated) is tallied under the catch-all radus_acct_drop, which
// was previously never incremented at all.
func TestSubmitAcctTaskCountsDropOnOverload(t *testing.T) {
	require.NoError(t, metrics.InitMetrics(""))

	pool, err := ants.NewPool(1, ants.WithNonblocking(true))
	require.NoError(t, err)
	defer pool.Release()

	acct := &AcctService{RadiusService: &RadiusService{TaskPool: pool}}

	started := make(chan struct{})
	block := make(chan struct{})
	require.True(t, acct.submitAcctTask(func() {
		close(started)
		<-block
	}, "busy"))
	<-started

	before := app.GetRadiusMetrics(app.MetricsRadiusAcctDrop)
	accepted := acct.submitAcctTask(func() {}, "overflow")
	assert.False(t, accepted, "task should be dropped when pool is saturated")
	assert.Equal(t, before+1, app.GetRadiusMetrics(app.MetricsRadiusAcctDrop),
		"a saturated-pool drop must increment radus_acct_drop")

	close(block)
	// Let the unblocked worker drain so the deferred pool release is clean.
	time.Sleep(20 * time.Millisecond)
}

// TestSendResponseCountsDropOnWriteError verifies that a failed Accounting-
// Response write is counted under the catch-all radus_acct_drop.
func TestSendResponseCountsDropOnWriteError(t *testing.T) {
	require.NoError(t, metrics.InitMetrics(""))

	acct := &AcctService{RadiusService: &RadiusService{}}
	req := &radius.Request{Packet: radius.New(radius.CodeAccountingRequest, []byte("secret"))}

	before := app.GetRadiusMetrics(app.MetricsRadiusAcctDrop)
	acct.SendResponse(errAcctResponseWriter{}, req)

	assert.Equal(t, before+1, app.GetRadiusMetrics(app.MetricsRadiusAcctDrop),
		"a failed accounting response write must increment radus_acct_drop")
}
