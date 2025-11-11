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

func TestHandleAuthErrorInvokesGuards(t *testing.T) {
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

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic from handleAuthError")
		}
		if !errors.Is(r.(error), guardErr) {
			t.Fatalf("expected panic with guard error, got %v", r)
		}
		if got := atomic.LoadInt32(&called); got != 1 {
			t.Fatalf("expected guard to be called once, got %d", got)
		}
	}()

	authSvc.handleAuthError(
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
}
