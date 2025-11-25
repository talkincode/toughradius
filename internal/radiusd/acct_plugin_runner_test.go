package radiusd

import (
	"context"
	"net"
	"sync/atomic"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"layeh.com/radius"
	"layeh.com/radius/rfc2866"
)

type mockAccountingHandler struct {
	name      string
	calls     *int32
	canHandle bool
	err       error
}

func (h *mockAccountingHandler) Name() string { return h.name }
func (h *mockAccountingHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	return h.canHandle
}
func (h *mockAccountingHandler) Handle(ctx *accounting.AccountingContext) error {
	atomic.AddInt32(h.calls, 1)
	return h.err
}

func TestHandleAccountingWithPluginsDispatchesToHandler(t *testing.T) {
	registry.ResetForTest()
	t.Cleanup(registry.ResetForTest)

	var calls int32
	registry.RegisterAccountingHandler(&mockAccountingHandler{
		name:      "handler",
		calls:     &calls,
		canHandle: true,
	})

	acctSvc := &AcctService{RadiusService: &RadiusService{}}
	packet := radius.New(radius.CodeAccountingRequest, []byte("secret"))
	_ = rfc2866.AcctStatusType_Set(packet, rfc2866.AcctStatusType_Value_Start)
	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1813},
	}

	vendorReq := &vendorparsers.VendorRequest{}
	nas := &domain.NetNas{Identifier: "NAS-1"}

	err := acctSvc.HandleAccountingWithPlugins(
		context.Background(),
		req,
		vendorReq,
		"alice",
		nas,
		"10.0.0.1",
	)
	if err != nil {
		t.Fatalf("HandleAccountingWithPlugins returned error: %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected handler to be called once, got %d", got)
	}
}

func TestHandleAccountingWithPluginsNoHandler(t *testing.T) {
	registry.ResetForTest()
	t.Cleanup(registry.ResetForTest)

	acctSvc := &AcctService{RadiusService: &RadiusService{}}
	packet := radius.New(radius.CodeAccountingRequest, []byte("secret"))
	_ = rfc2866.AcctStatusType_Set(packet, rfc2866.AcctStatusType_Value_Stop)
	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1813},
	}

	err := acctSvc.HandleAccountingWithPlugins(
		context.Background(),
		req,
		&vendorparsers.VendorRequest{},
		"alice",
		&domain.NetNas{},
		"10.0.0.1",
	)
	if err == nil {
		t.Fatalf("expected error when no handlers registered")
	}
}
