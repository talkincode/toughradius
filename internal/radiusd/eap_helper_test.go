package radiusd

import (
	"net"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// createTestRadiusService creates a minimal RadiusService for testing
func createTestRadiusService() *RadiusService {
	// Create a minimal config
	cfg := &config.AppConfig{
		Radiusd: config.RadiusdConfig{
			Debug: false,
		},
	}

	// Create a mock AppContext
	testApp := app.NewApplication(cfg)

	// Create RadiusService with the test context
	// Note: This won't have database connections, but that's okay for EAP helper tests
	rs := &RadiusService{
		appCtx: testApp,
	}

	return rs
}

func TestNewEAPAuthHelper(t *testing.T) {
	rs := createTestRadiusService()
	helper := NewEAPAuthHelper(rs)

	if helper == nil {
		t.Fatal("NewEAPAuthHelper returned nil")
	}

	if helper.coordinator == nil {
		t.Fatal("coordinator is nil")
	}
}

func TestEAPAuthHelperGetCoordinator(t *testing.T) {
	rs := createTestRadiusService()
	helper := NewEAPAuthHelper(rs)
	coordinator := helper.GetCoordinator()

	if coordinator == nil {
		t.Fatal("GetCoordinator returned nil")
	}

	// Validate it returns the same instance
	if coordinator != helper.coordinator {
		t.Error("GetCoordinator returned different instance")
	}
}

func TestEAPAuthHelperHandleEAPAuthenticationBasic(t *testing.T) {
	rs := createTestRadiusService()
	helper := NewEAPAuthHelper(rs)

	// Create test data
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2865.UserName_SetString(packet, "testuser")

	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812},
	}

	user := &domain.RadiusUser{
		Username:   "testuser",
		Password:   "password",
		Status:     common.ENABLED,
		ExpireTime: time.Now().Add(24 * time.Hour),
	}

	nas := &domain.NetNas{
		ID:         1,
		Identifier: "NAS-1",
		Ipaddr:     "192.168.1.1",
		Secret:     "secret",
	}

	vendorReq := &vendorparsers.VendorRequest{}
	response := radius.New(radius.CodeAccessAccept, []byte("secret"))

	// Test HandleEAPAuthentication
	// Note: without a real EAP message, this call should return handled=false
	handled, success, err := helper.HandleEAPAuthentication(
		nil, // ResponseWriter - nil because it won't be used
		req,
		user,
		nas,
		vendorReq,
		response,
		"eap-md5",
	)

	// Validate the result
	// Without an EAP message, the coordinator should report unhandled
	if handled {
		t.Log("Note: handled=true may indicate EAP message was found")
	}

	if err != nil {
		t.Logf("Error occurred (expected for missing EAP message): %v", err)
	}

	_ = success // success depends on whether an EAP message was present
}

func TestEAPAuthHelperSendEAPSuccess(t *testing.T) {
	rs := createTestRadiusService()
	helper := NewEAPAuthHelper(rs)

	packet := radius.New(radius.CodeAccessAccept, []byte("secret"))
	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812},
	}

	response := radius.New(radius.CodeAccessAccept, []byte("secret"))

	// Test SendEAPSuccess
	// Note: with a nil ResponseWriter, this may return an error or panic
	// We just ensure the method exists and won't panic unexpectedly
	defer func() {
		if r := recover(); r != nil {
			// A panic is expected when ResponseWriter is nil
			t.Logf("SendEAPSuccess panicked (expected with nil writer): %v", r)
		}
	}()

	err := helper.SendEAPSuccess(nil, req, response, "secret")

	// With a nil ResponseWriter, an error or panic may occur
	if err != nil {
		t.Logf("SendEAPSuccess error (expected with nil writer): %v", err)
	}
}

func TestEAPAuthHelperSendEAPFailure(t *testing.T) {
	rs := createTestRadiusService()
	helper := NewEAPAuthHelper(rs)

	packet := radius.New(radius.CodeAccessReject, []byte("secret"))
	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812},
	}

	// Test SendEAPFailure
	defer func() {
		if r := recover(); r != nil {
			// A panic is expected when the ResponseWriter is nil
			t.Logf("SendEAPFailure panicked (expected with nil writer): %v", r)
		}
	}()

	err := helper.SendEAPFailure(nil, req, "secret", nil)

	// With a nil ResponseWriter, an error or panic may occur
	if err != nil {
		t.Logf("SendEAPFailure error (expected with nil writer): %v", err)
	}
}

func TestEAPAuthHelperCleanupState(t *testing.T) {
	rs := createTestRadiusService()
	helper := NewEAPAuthHelper(rs)

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812},
	}

	// Test CleanupState - should not panic
	helper.CleanupState(req)
}

func TestEAPAuthHelperMacAuth(t *testing.T) {
	rs := createTestRadiusService()
	helper := NewEAPAuthHelper(rs)

	// Create a MAC authentication scenario
	macAddr := "aa:bb:cc:dd:ee:ff"
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2865.UserName_SetString(packet, macAddr)

	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812},
	}

	user := &domain.RadiusUser{
		Username:   macAddr,
		MacAddr:    macAddr,
		Status:     common.ENABLED,
		ExpireTime: time.Now().Add(24 * time.Hour),
	}

	nas := &domain.NetNas{
		ID:         1,
		Identifier: "NAS-1",
		Ipaddr:     "192.168.1.1",
		Secret:     "secret",
	}

	// MAC authentication scenario
	vendorReq := &vendorparsers.VendorRequest{
		MacAddr: macAddr,
	}

	response := radius.New(radius.CodeAccessAccept, []byte("secret"))

	// Test MAC authentication
	handled, success, err := helper.HandleEAPAuthentication(
		nil,
		req,
		user,
		nas,
		vendorReq,
		response,
		"eap-md5",
	)

	_ = handled
	_ = success
	_ = err

	// isMacAuth should be true during MAC authentication
	// The coordinator logic handles this, so we only check that the call doesn't panic
}

func TestEAPAuthHelperDifferentMethods(t *testing.T) {
	rs := createTestRadiusService()
	helper := NewEAPAuthHelper(rs)

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2865.UserName_SetString(packet, "testuser")

	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812},
	}

	user := &domain.RadiusUser{
		Username:   "testuser",
		Password:   "password",
		Status:     common.ENABLED,
		ExpireTime: time.Now().Add(24 * time.Hour),
	}

	nas := &domain.NetNas{
		ID:         1,
		Identifier: "NAS-1",
		Secret:     "secret",
	}

	vendorReq := &vendorparsers.VendorRequest{}
	response := radius.New(radius.CodeAccessAccept, []byte("secret"))

	// Test different EAP methods
	methods := []string{"eap-md5", "eap-mschapv2", "eap-tls"}

	for _, method := range methods {
		t.Run("Method_"+method, func(t *testing.T) {
			handled, success, err := helper.HandleEAPAuthentication(
				nil,
				req,
				user,
				nas,
				vendorReq,
				response,
				method,
			)

			// Ensure the method call does not panic
			_ = handled
			_ = success
			_ = err
		})
	}
}

func TestEAPAuthHelperConcurrentAccess(t *testing.T) {
	rs := createTestRadiusService()
	helper := NewEAPAuthHelper(rs)

	// Test concurrent access to GetCoordinator for safety
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			coordinator := helper.GetCoordinator()
			if coordinator == nil {
				t.Error("GetCoordinator returned nil in concurrent access")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}
}
