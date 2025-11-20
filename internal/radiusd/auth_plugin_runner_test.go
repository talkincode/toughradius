package radiusd

import (
	"context"
	"net"
	"sync/atomic"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"layeh.com/radius"
)

type mockValidator struct {
	name   string
	calls  *int32
	handle bool
}

func (v *mockValidator) Name() string { return v.name }
func (v *mockValidator) CanHandle(ctx *auth.AuthContext) bool {
	return v.handle
}
func (v *mockValidator) Validate(ctx context.Context, authCtx *auth.AuthContext, password string) error {
	atomic.AddInt32(v.calls, 1)
	return nil
}

type mockChecker struct {
	name  string
	order int
	calls *int32
}

func (c *mockChecker) Name() string { return c.name }
func (c *mockChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	atomic.AddInt32(c.calls, 1)
	return nil
}
func (c *mockChecker) Order() int { return c.order }

func newTestAuthService() *AuthService {
	return &AuthService{RadiusService: &RadiusService{}}
}

func TestAuthenticateUserWithPluginsRunsValidatorAndPolicies(t *testing.T) {
	registry.ResetForTest()
	t.Cleanup(registry.ResetForTest)

	var validatorCalls, checkerCalls int32

	registry.RegisterPasswordValidator(&mockValidator{
		name:   "validator",
		calls:  &validatorCalls,
		handle: true,
	})

	registry.RegisterPolicyChecker(&mockChecker{
		name:  "checker",
		order: 1,
		calls: &checkerCalls,
	})

	authSvc := newTestAuthService()
	user := &domain.RadiusUser{Username: "alice", Password: "secret"}
	nas := &domain.NetNas{Identifier: "NAS-1"}
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1812},
	}
	resp := packet.Response(radius.CodeAccessAccept)
	vendorReq := &vendorparsers.VendorRequest{}

	err := authSvc.AuthenticateUserWithPlugins(context.Background(), req, resp, user, nas, vendorReq, false)
	if err != nil {
		t.Fatalf("AuthenticateUserWithPlugins returned error: %v", err)
	}

	if got := atomic.LoadInt32(&validatorCalls); got != 1 {
		t.Fatalf("expected validator called once, got %d", got)
	}
	if got := atomic.LoadInt32(&checkerCalls); got != 1 {
		t.Fatalf("expected checker called once, got %d", got)
	}
}

func TestAuthenticateUserWithPluginsSkipsValidatorWhenOptionUsed(t *testing.T) {
	registry.ResetForTest()
	t.Cleanup(registry.ResetForTest)

	var validatorCalls int32
	registry.RegisterPasswordValidator(&mockValidator{
		name:   "validator",
		calls:  &validatorCalls,
		handle: true,
	})

	authSvc := newTestAuthService()
	user := &domain.RadiusUser{Username: "bob", Password: "secret"}
	nas := &domain.NetNas{Identifier: "NAS-1"}
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1812},
	}
	resp := packet.Response(radius.CodeAccessAccept)
	vendorReq := &vendorparsers.VendorRequest{}

	err := authSvc.AuthenticateUserWithPlugins(
		context.Background(),
		req,
		resp,
		user,
		nas,
		vendorReq,
		false,
		SkipPasswordValidation(),
	)
	if err != nil {
		t.Fatalf("AuthenticateUserWithPlugins returned error: %v", err)
	}

	if got := atomic.LoadInt32(&validatorCalls); got != 0 {
		t.Fatalf("expected validator skipped, got %d calls", got)
	}
}
