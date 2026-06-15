package radiusd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/pkg/metrics"
)

// TestServeRADIUS_PanicCountsAuthDrop locks in the panic-path observability fix:
// ServeRADIUS recovers from an unexpected stage panic, and that recovery must
// increment radus_auth_drop (previously the counter name was only emitted as a
// zap field) so a crash storm is visible on the counter, not just in the logs.
// A pipeline whose single stage panics is injected; ServeRADIUS must recover,
// count exactly one auth drop, and still answer the client with an Access-Reject.
// captureResponseWriter and metricsTestAppContext are defined in
// radius_auth_metrics_test.go.
func TestServeRADIUS_PanicCountsAuthDrop(t *testing.T) {
	require.NoError(t, metrics.InitMetrics(""))

	svc := &AuthService{RadiusService: &RadiusService{appCtx: &metricsTestAppContext{}}}
	// Inject a pipeline whose only stage panics; ensurePipeline keeps a
	// pre-set pipeline, so the default stages are not registered.
	svc.authPipeline = NewAuthPipeline().Use(newStage("boom", func(*AuthPipelineContext) error {
		panic("boom")
	}))

	req := &radius.Request{Packet: radius.New(radius.CodeAccessRequest, []byte("secret"))}
	w := &captureResponseWriter{}

	before := app.GetRadiusMetrics(app.MetricsRadiusAuthDrop)
	require.NotPanics(t, func() { svc.ServeRADIUS(w, req) },
		"ServeRADIUS must recover from a stage panic")

	assert.Equal(t, before+1, app.GetRadiusMetrics(app.MetricsRadiusAuthDrop),
		"a recovered panic must increment radus_auth_drop exactly once")
	require.NotNil(t, w.last, "the recovered request must still be answered")
	assert.Equal(t, radius.CodeAccessReject, w.last.Code,
		"a recovered panic must answer with an Access-Reject")
}
