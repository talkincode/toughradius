package radiusd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/app"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	eap "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/pkg/metrics"
)

func TestMapEAPDispatchError(t *testing.T) {
	tests := []struct {
		name        string
		input       error
		metricsType string
		stage       string
		contains    string
	}{
		{
			name:        "password mismatch maps to passwd metric",
			input:       eap.ErrPasswordMismatch,
			metricsType: app.MetricsRadiusRejectPasswdError,
			stage:       StageEAPDispatch,
			contains:    "eap password validation failed",
		},
		{
			name:        "tls identity mismatch maps to unauthorized metric",
			input:       eap.ErrTLSIdentityMismatch,
			metricsType: app.MetricsRadiusRejectUnauthorized,
			stage:       StageEAPDispatch,
			contains:    "identity mismatch",
		},
		{
			name:        "wrapped tls not configured maps to tls config reason",
			input:       fmt.Errorf("wrap: %w", eap.ErrTLSNotConfigured),
			metricsType: app.MetricsRadiusRejectOther,
			stage:       StageEAPDispatch,
			contains:    "eap-tls trust configuration missing",
		},
		{
			name:        "peap inner protocol violation maps to reject-other with clear reason",
			input:       fmt.Errorf("wrap: %w", eap.ErrPEAPInnerProtocol),
			metricsType: app.MetricsRadiusRejectOther,
			stage:       StageEAPDispatch,
			contains:    "peap inner eap-mschapv2 protocol violation",
		},
		{
			name:        "peap inner not implemented maps to reject-other with clear reason",
			input:       eap.ErrPEAPInnerNotImplemented,
			metricsType: app.MetricsRadiusRejectOther,
			stage:       StageEAPDispatch,
			contains:    "peap inner eap method unavailable",
		},
		{
			name:        "peap inner password mismatch maps to passwd metric",
			input:       eap.ErrPasswordMismatch,
			metricsType: app.MetricsRadiusRejectPasswdError,
			stage:       StageEAPDispatch,
			contains:    "eap password validation failed",
		},
		{
			name:        "unknown eap error falls back to reject-other metric",
			input:       errors.New("boom"),
			metricsType: app.MetricsRadiusRejectOther,
			stage:       StageEAPDispatch,
			contains:    "eap authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := mapEAPDispatchError(tt.input)
			authErr, ok := radiuserrors.GetAuthError(out)
			assert.True(t, ok)
			assert.Equal(t, tt.metricsType, authErr.MetricsType)
			assert.Equal(t, tt.stage, authErr.ErrorStage)
			assert.Contains(t, authErr.Error(), tt.contains)
		})
	}
}

func TestMapEAPDispatchError_UsesWrappedSentinel(t *testing.T) {
	input := errors.Join(errors.New("failed handshake"), eap.ErrTLSHandshakeFailed)
	out := mapEAPDispatchError(input)
	authErr, ok := radiuserrors.GetAuthError(out)
	assert.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectOther, authErr.MetricsType)
	assert.Equal(t, StageEAPDispatch, authErr.ErrorStage)
	assert.Contains(t, authErr.Error(), "eap-tls handshake failed")
}

func TestMapEAPDispatchError_PreservesExistingAuthErrorMetric(t *testing.T) {
	input := radiuserrors.NewAuthError(app.MetricsRadiusRejectLimit, "rate limited")
	out := mapEAPDispatchError(input)
	authErr, ok := radiuserrors.GetAuthError(out)
	assert.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectLimit, authErr.MetricsType)
	assert.Equal(t, StageEAPDispatch, authErr.ErrorStage)
	assert.Equal(t, "rate limited", authErr.Message)
}

func TestSafeEAPFailureReason_DoesNotExposeCause(t *testing.T) {
	err := radiuserrors.NewAuthErrorWithCause(app.MetricsRadiusRejectOther, "eap-tls handshake failed", errors.New("sensitive password leak"))
	reason := safeEAPFailureReason(err)
	assert.Equal(t, "eap-tls handshake failed", reason)
	assert.NotContains(t, reason, "sensitive")
}

func TestLogEAPFailure_IncrementsMappedMetric(t *testing.T) {
	assert.NoError(t, metrics.InitMetrics(""))

	s := &AuthService{}
	ctx := &AuthPipelineContext{
		Username: "alice",
		RemoteIP: "10.0.0.1",
	}
	err := radiuserrors.NewAuthErrorWithStage(app.MetricsRadiusRejectUnauthorized, "eap-tls certificate identity mismatch", StageEAPDispatch)

	before := app.GetRadiusMetrics(app.MetricsRadiusRejectUnauthorized)
	s.logEAPFailure(ctx, err)
	after := app.GetRadiusMetrics(app.MetricsRadiusRejectUnauthorized)

	assert.Equal(t, before+1, after)
}

func TestLogEAPFailure_UsesRejectOtherForUnknownError(t *testing.T) {
	assert.NoError(t, metrics.InitMetrics(""))

	s := &AuthService{}
	ctx := &AuthPipelineContext{
		Username: "alice",
		RemoteIP: "10.0.0.1",
	}

	before := app.GetRadiusMetrics(app.MetricsRadiusRejectOther)
	s.logEAPFailure(ctx, errors.New("unexpected failure"))
	after := app.GetRadiusMetrics(app.MetricsRadiusRejectOther)

	assert.Equal(t, before+1, after)
}
