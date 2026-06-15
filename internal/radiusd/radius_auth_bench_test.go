package radiusd

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	cachepkg "github.com/talkincode/toughradius/v9/internal/radiusd/cache"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type benchmarkAuthAppContext struct {
	metricsTestAppContext

	configMgr    *app.ConfigManager
	profileCache *app.ProfileCache
}

func (m *benchmarkAuthAppContext) ConfigMgr() *app.ConfigManager { return m.configMgr }
func (m *benchmarkAuthAppContext) ProfileCache() *app.ProfileCache {
	return m.profileCache
}

type benchmarkUserRepository struct {
	user *domain.RadiusUser
}

func (r *benchmarkUserRepository) GetByUsername(context.Context, string) (*domain.RadiusUser, error) {
	return r.user, nil
}
func (r *benchmarkUserRepository) GetByMacAddr(context.Context, string) (*domain.RadiusUser, error) {
	return r.user, nil
}
func (r *benchmarkUserRepository) UpdateMacAddr(context.Context, string, string) error {
	return nil
}
func (r *benchmarkUserRepository) UpdateVlanId(context.Context, string, int, int) error {
	return nil
}
func (r *benchmarkUserRepository) UpdateLastOnline(context.Context, string) error {
	return nil
}
func (r *benchmarkUserRepository) UpdateField(context.Context, string, string, interface{}) error {
	return nil
}

type benchmarkNasRepository struct {
	nas *domain.NetNas
}

func (r *benchmarkNasRepository) GetByIP(context.Context, string) (*domain.NetNas, error) {
	return r.nas, nil
}
func (r *benchmarkNasRepository) GetByIdentifier(context.Context, string) (*domain.NetNas, error) {
	return r.nas, nil
}
func (r *benchmarkNasRepository) GetByIPOrIdentifier(context.Context, string, string) (*domain.NetNas, error) {
	return r.nas, nil
}

type benchmarkPasswordValidator struct{}

func (benchmarkPasswordValidator) Name() string { return "benchmark-password-validator" }
func (benchmarkPasswordValidator) CanHandle(*auth.AuthContext) bool {
	return true
}
func (benchmarkPasswordValidator) Validate(context.Context, *auth.AuthContext, string) error {
	return nil
}

func newBenchmarkAuthAppContext(b *testing.B) *benchmarkAuthAppContext {
	b.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatalf("open benchmark config database: %v", err)
	}
	if err := db.AutoMigrate(&domain.SysConfig{}); err != nil {
		b.Fatalf("migrate benchmark config database: %v", err)
	}

	application := app.NewApplication(&config.AppConfig{})
	application.OverrideDB(db)
	profileCache := app.NewProfileCache(nil, time.Minute)
	b.Cleanup(profileCache.Stop)

	return &benchmarkAuthAppContext{
		configMgr:    app.NewConfigManager(application),
		profileCache: profileCache,
	}
}

func newBenchmarkAuthService(b *testing.B) *AuthService {
	b.Helper()

	user := &domain.RadiusUser{
		Username:   "bench-user",
		Password:   "bench-pass",
		Status:     common.ENABLED,
		ExpireTime: time.Now().Add(time.Hour),
	}
	nas := &domain.NetNas{
		Identifier: "bench-nas",
		Ipaddr:     "127.0.0.1",
		Secret:     "testing123",
		VendorCode: "0",
		Status:     common.ENABLED,
	}
	radiusService := &RadiusService{
		appCtx:    newBenchmarkAuthAppContext(b),
		authRate:  newAuthRateLimiter(defaultAuthRateShards),
		nasCache:  cachepkg.NewTTLCache[*domain.NetNas](time.Minute, 16),
		userCache: cachepkg.NewTTLCache[*domain.RadiusUser](time.Minute, 16),
		UserRepo:  &benchmarkUserRepository{user: user},
		NasRepo:   &benchmarkNasRepository{nas: nas},
	}

	return NewAuthService(radiusService)
}

func newBenchmarkAuthRequest(b *testing.B) *radius.Request {
	b.Helper()

	packet := radius.New(radius.CodeAccessRequest, []byte("testing123"))
	if err := rfc2865.UserName_SetString(packet, "bench-user"); err != nil {
		b.Fatalf("set User-Name: %v", err)
	}
	if err := rfc2865.NASIdentifier_SetString(packet, "bench-nas"); err != nil {
		b.Fatalf("set NAS-Identifier: %v", err)
	}

	return &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1812},
	}
}

// BenchmarkServeRADIUSAccessAccept measures the real authentication hot path
// from Access-Request ingress through NAS lookup, user lookup, plugin password
// validation, response signing, Access-Accept write, bind update, metrics, and
// last-online update. Repositories and plugins are in-memory/no-op so the
// benchmark isolates RADIUS pipeline overhead rather than database latency.
func BenchmarkServeRADIUSAccessAccept(b *testing.B) {
	undoLogger := zap.ReplaceGlobals(zap.NewNop())
	b.Cleanup(undoLogger)
	registry.ResetForTest()
	b.Cleanup(registry.ResetForTest)
	registry.RegisterPasswordValidator(benchmarkPasswordValidator{})

	svc := newBenchmarkAuthService(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := &captureResponseWriter{}
		svc.ServeRADIUS(w, newBenchmarkAuthRequest(b))
		if w.last == nil {
			b.Fatal("ServeRADIUS did not write a response")
		}
		if w.last.Code != radius.CodeAccessAccept {
			b.Fatalf("ServeRADIUS response code = %v, want %v", w.last.Code, radius.CodeAccessAccept)
		}
	}
}
