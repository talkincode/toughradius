package radiusd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/app"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	eap "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
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
