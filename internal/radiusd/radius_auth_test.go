package radiusd

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"gorm.io/gorm"
	"layeh.com/radius"
)

// mockAppContext implements app.AppContext for testing
type mockAppContext struct{}

func (m *mockAppContext) DB() *gorm.DB                                       { return nil }
func (m *mockAppContext) Config() *config.AppConfig                          { return nil }
func (m *mockAppContext) GetSettingsStringValue(category, key string) string { return "" }
func (m *mockAppContext) GetSettingsInt64Value(category, key string) int64   { return 0 }
func (m *mockAppContext) GetSettingsBoolValue(category, key string) bool     { return false }
func (m *mockAppContext) SaveSettings(settings map[string]interface{}) error { return nil }
func (m *mockAppContext) Scheduler() *cron.Cron                              { return nil }
func (m *mockAppContext) ConfigMgr() *app.ConfigManager                      { return nil }
func (m *mockAppContext) ProfileCache() *app.ProfileCache                    { return nil }
func (m *mockAppContext) MigrateDB(track bool) error                         { return nil }
func (m *mockAppContext) InitDb()                                            {}
func (m *mockAppContext) DropAll()                                           {}

type testEnhancer struct {
	name  string
	calls *int32
}

func (e *testEnhancer) Name() string {
	return e.name
}

func (e *testEnhancer) Enhance(ctx context.Context, authCtx *auth.AuthContext) error {
	atomic.AddInt32(e.calls, 1)
	return nil
}

type testGuard struct {
	name   string
	calls  *int32
	result error
}

func (g *testGuard) Name() string {
	return g.name
}

func (g *testGuard) OnError(ctx context.Context, authCtx *auth.AuthContext, stage string, err error) error {
	atomic.AddInt32(g.calls, 1)
	return g.result
}

func (g *testGuard) OnAuthError(ctx context.Context, authCtx *auth.AuthContext, stage string, err error) *auth.GuardResult {
	return nil // Use fallback to OnError
}

func TestApplyAcceptEnhancersInvokesRegisteredEnhancers(t *testing.T) {
	registry.ResetForTest()
	t.Cleanup(registry.ResetForTest)

	var called int32
	registry.RegisterResponseEnhancer(&testEnhancer{
		name:  "test-enhancer",
		calls: &called,
	})

	authSvc := &AuthService{RadiusService: &RadiusService{}}
	user := &domain.RadiusUser{
		Username:   "alice",
		ExpireTime: time.Now().Add(time.Hour),
	}
	nas := &domain.NetNas{
		ID:         1,
		Identifier: "NAS-1",
	}
	resp := radius.New(radius.CodeAccessAccept, []byte("secret"))
	vendorReq := &vendorparsers.VendorRequest{}

	authSvc.ApplyAcceptEnhancers(user, nas, vendorReq, resp)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Fatalf("expected enhancer to be called once, got %d", got)
	}
}

func TestProcessAuthErrorInvokesGuards(t *testing.T) {
	registry.ResetForTest()
	t.Cleanup(registry.ResetForTest)

	var called int32
	guardErr := errors.New("guard rejected")

	registry.RegisterAuthGuard(&testGuard{
		name:   "test-guard",
		calls:  &called,
		result: guardErr,
	})

	// Create a mock RadiusService with AppContext
	mockAppCtx := &mockAppContext{}
	authSvc := &AuthService{RadiusService: &RadiusService{appCtx: mockAppCtx}}

	// processAuthError should return error instead of panic
	finalErr := authSvc.processAuthError(
		"test-stage",
		nil,
		nil,
		nil,
		nil,
		false,
		"bob",
		"10.0.0.1",
		errors.New("original error"),
	)

	// Verify guard was called
	if got := atomic.LoadInt32(&called); got != 1 {
		t.Fatalf("expected guard to be called once, got %d", got)
	}

	// Verify error is returned (not panic)
	if finalErr == nil {
		t.Fatal("expected error to be returned")
	}
	if !errors.Is(finalErr, guardErr) {
		t.Fatalf("expected guard error, got %v", finalErr)
	}
}

func TestProcessAuthErrorWithNilError(t *testing.T) {
	authSvc := &AuthService{RadiusService: &RadiusService{}}

	// nil error should return nil
	finalErr := authSvc.processAuthError(
		"test-stage",
		nil, nil, nil, nil, false, "", "",
		nil,
	)

	if finalErr != nil {
		t.Fatalf("expected nil error, got %v", finalErr)
	}
}

func TestProcessAuthErrorNoGuards(t *testing.T) {
	registry.ResetForTest()
	t.Cleanup(registry.ResetForTest)

	authSvc := &AuthService{RadiusService: &RadiusService{}}
	originalErr := errors.New("original error")

	// Without guards, original error should be returned
	finalErr := authSvc.processAuthError(
		"test-stage",
		nil, nil, nil, nil, false, "", "",
		originalErr,
	)

	if !errors.Is(finalErr, originalErr) {
		t.Fatalf("expected original error, got %v", finalErr)
	}
}
