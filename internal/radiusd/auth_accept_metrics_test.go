package radiusd

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"layeh.com/radius"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	repogorm "github.com/talkincode/toughradius/v9/internal/radiusd/repository/gorm"
	"github.com/talkincode/toughradius/v9/pkg/metrics"
)

// newAcceptMetricsTestService builds an AuthService backed by an in-memory
// SQLite UserRepo and the metrics mock app context. The plugin registry is
// reset to empty for the duration of the test so sendAcceptResponse exercises
// its own accept-counting logic without invoking response enhancers (their
// attribute behavior is covered by the enhancer tests). captureResponseWriter
// and metricsTestAppContext are defined in radius_auth_metrics_test.go.
func newAcceptMetricsTestService(t *testing.T) *AuthService {
	t.Helper()
	registry.ResetForTest()
	t.Cleanup(registry.ResetForTest)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.RadiusUser{}))
	require.NoError(t, db.Create(&domain.RadiusUser{Username: "alice"}).Error)

	return &AuthService{RadiusService: &RadiusService{
		appCtx:   &metricsTestAppContext{},
		UserRepo: repogorm.NewGormUserRepository(db),
	}}
}

func newAcceptPipelineContext(w radius.ResponseWriter, user *domain.RadiusUser) *AuthPipelineContext {
	req := &radius.Request{Packet: radius.New(radius.CodeAccessRequest, []byte("secret"))}
	return &AuthPipelineContext{
		Writer:   w,
		Request:  req,
		Response: req.Response(radius.CodeAccessAccept),
		Username: "alice",
		RemoteIP: "10.0.0.1",
		NAS:      &domain.NetNas{},
		User:     user,
	}
}

// TestSendAcceptResponse_CountsAccept locks in the accept-side observability
// counter: a successful (bare PAP/CHAP) authentication must increment
// radus_accept exactly once so it pairs with the radus_reject_* counters for
// success-rate and SLO dashboards. Before this fix radus_accept was only logged
// as a zap field and stayed at 0 forever.
func TestSendAcceptResponse_CountsAccept(t *testing.T) {
	require.NoError(t, metrics.InitMetrics(""))
	svc := newAcceptMetricsTestService(t)
	w := &captureResponseWriter{}
	ctx := newAcceptPipelineContext(w, &domain.RadiusUser{Username: "alice"})

	before := app.GetRadiusMetrics(app.MetricsRadiusAccept)
	svc.sendAcceptResponse(ctx, false)

	require.NotNil(t, w.last, "an Access-Accept packet must be written")
	assert.Equal(t, radius.CodeAccessAccept, w.last.Code)
	assert.Equal(t, before+1, app.GetRadiusMetrics(app.MetricsRadiusAccept),
		"a successful auth must increment radus_accept exactly once")
}

// TestSendAcceptResponse_EapFlowAlsoCounts guards against scoping the increment
// to the bare path only (for example wrapping it in "if !isEapFlow"): the
// EAP-success caller flows through the same chokepoint and must also be counted.
func TestSendAcceptResponse_EapFlowAlsoCounts(t *testing.T) {
	require.NoError(t, metrics.InitMetrics(""))
	svc := newAcceptMetricsTestService(t)
	w := &captureResponseWriter{}
	ctx := newAcceptPipelineContext(w, &domain.RadiusUser{Username: "alice"})

	before := app.GetRadiusMetrics(app.MetricsRadiusAccept)
	svc.sendAcceptResponse(ctx, true)

	assert.Equal(t, before+1, app.GetRadiusMetrics(app.MetricsRadiusAccept),
		"the EAP-success path must increment radus_accept exactly once")
}

// TestSendAcceptResponse_MissingContextNotCounted verifies the early-return
// guard: when the NAS/User context is incomplete no Access-Accept is emitted, so
// radus_accept must NOT be incremented (no phantom successes for dropped requests).
func TestSendAcceptResponse_MissingContextNotCounted(t *testing.T) {
	require.NoError(t, metrics.InitMetrics(""))
	svc := newAcceptMetricsTestService(t)
	w := &captureResponseWriter{}
	ctx := newAcceptPipelineContext(w, nil) // User nil -> guard fires before send

	before := app.GetRadiusMetrics(app.MetricsRadiusAccept)
	svc.sendAcceptResponse(ctx, false)

	assert.Nil(t, w.last, "no packet must be written when context is incomplete")
	assert.Equal(t, before, app.GetRadiusMetrics(app.MetricsRadiusAccept),
		"an incomplete-context drop must not be counted as an accept")
}
