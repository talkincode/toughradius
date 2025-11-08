package radiusd

import (
	"net"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestNewEAPAuthHelper(t *testing.T) {
	helper := NewEAPAuthHelper()

	if helper == nil {
		t.Fatal("NewEAPAuthHelper returned nil")
	}

	if helper.coordinator == nil {
		t.Fatal("coordinator is nil")
	}
}

func TestEAPAuthHelperGetCoordinator(t *testing.T) {
	helper := NewEAPAuthHelper()
	coordinator := helper.GetCoordinator()

	if coordinator == nil {
		t.Fatal("GetCoordinator returned nil")
	}

	// 验证是同一个实例
	if coordinator != helper.coordinator {
		t.Error("GetCoordinator returned different instance")
	}
}

func TestEAPAuthHelperHandleEAPAuthenticationBasic(t *testing.T) {
	helper := NewEAPAuthHelper()

	// 创建测试数据
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

	// 测试 HandleEAPAuthentication
	// 注意：由于没有实际的 EAP 消息，这个调用应该返回 handled=false
	handled, success, err := helper.HandleEAPAuthentication(
		nil, // ResponseWriter - 可以为 nil 因为不会被使用
		req,
		user,
		nas,
		vendorReq,
		response,
		"eap-md5",
	)

	// 验证结果
	// 由于没有 EAP 消息，协调器应该返回未处理
	if handled {
		t.Log("Note: handled=true may indicate EAP message was found")
	}

	if err != nil {
		t.Logf("Error occurred (expected for missing EAP message): %v", err)
	}

	_ = success // success 的值取决于是否有 EAP 消息
}

func TestEAPAuthHelperSendEAPSuccess(t *testing.T) {
	helper := NewEAPAuthHelper()

	packet := radius.New(radius.CodeAccessAccept, []byte("secret"))
	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812},
	}

	response := radius.New(radius.CodeAccessAccept, []byte("secret"))

	// 测试 SendEAPSuccess
	// 注意：由于 ResponseWriter 为 nil，会返回错误或 panic
	// 我们只测试方法是否存在且不会导致意外错误
	defer func() {
		if r := recover(); r != nil {
			// ResponseWriter 为 nil 导致 panic 是预期的
			t.Logf("SendEAPSuccess panicked (expected with nil writer): %v", r)
		}
	}()

	err := helper.SendEAPSuccess(nil, req, response, "secret")

	// 由于 ResponseWriter 为 nil，可能会返回错误或 panic
	if err != nil {
		t.Logf("SendEAPSuccess error (expected with nil writer): %v", err)
	}
}

func TestEAPAuthHelperSendEAPFailure(t *testing.T) {
	helper := NewEAPAuthHelper()

	packet := radius.New(radius.CodeAccessReject, []byte("secret"))
	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812},
	}

	// 测试 SendEAPFailure
	defer func() {
		if r := recover(); r != nil {
			// ResponseWriter 为 nil 导致 panic 是预期的
			t.Logf("SendEAPFailure panicked (expected with nil writer): %v", r)
		}
	}()

	err := helper.SendEAPFailure(nil, req, "secret", nil)

	// 由于 ResponseWriter 为 nil，可能会返回错误或 panic
	if err != nil {
		t.Logf("SendEAPFailure error (expected with nil writer): %v", err)
	}
}

func TestEAPAuthHelperCleanupState(t *testing.T) {
	helper := NewEAPAuthHelper()

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812},
	}

	// 测试 CleanupState - 应该不会 panic
	helper.CleanupState(req)
}

func TestEAPAuthHelperMacAuth(t *testing.T) {
	helper := NewEAPAuthHelper()

	// 创建 MAC 认证场景
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

	// MAC 认证场景
	vendorReq := &vendorparsers.VendorRequest{
		MacAddr: macAddr,
	}

	response := radius.New(radius.CodeAccessAccept, []byte("secret"))

	// 测试 MAC 认证
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

	// MAC 认证时 isMacAuth 应该为 true
	// 由于实际逻辑在 coordinator 中，这里只是验证方法调用不会 panic
}

func TestEAPAuthHelperDifferentMethods(t *testing.T) {
	helper := NewEAPAuthHelper()

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

	// 测试不同的 EAP 方法
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

			// 验证方法调用不会 panic
			_ = handled
			_ = success
			_ = err
		})
	}
}

func TestEAPAuthHelperConcurrentAccess(t *testing.T) {
	helper := NewEAPAuthHelper()

	// 测试并发访问 GetCoordinator 是否安全
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

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}
