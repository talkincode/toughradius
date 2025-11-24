package radiusd

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins"
	parsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers/parsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/huawei"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/mikrotik"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
)

func reRegisterVendorParsers() {
	registry.RegisterVendorParser(&parsers.DefaultParser{})
	registry.RegisterVendorParser(&parsers.HuaweiParser{})
	registry.RegisterVendorParser(&parsers.H3CParser{})
	registry.RegisterVendorParser(&parsers.ZTEParser{})
}

func getFreePort() (int, error) {
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenUDP("udp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.LocalAddr().(*net.UDPAddr).Port, nil
}

func setupTestEnv(t *testing.T) (*app.Application, *config.AppConfig) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "toughradius_test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	// Prepare required workdir sub-directories similar to config.AppConfig.initDirs
	requiredDirs := []string{
		"logs",
		"public",
		"private",
		"backup",
		"data",
		filepath.Join("data", "metrics"),
	}
	for _, dir := range requiredDirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0o755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	authPort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}
	acctPort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "ToughRADIUS-Test",
			Location: "Asia/Shanghai",
			Workdir:  tmpDir,
			Debug:    true,
		},
		Database: config.DBConfig{
			Type: "sqlite",
			Name: filepath.Join(tmpDir, "test.db"),
		},
		Radiusd: config.RadiusdConfig{
			Enabled:  true,
			Host:     "127.0.0.1",
			AuthPort: authPort,
			AcctPort: acctPort,
			Debug:    true,
		},
		Logger: config.LogConfig{
			Mode:       "development",
			FileEnable: false,
		},
	}

	application := app.NewApplication(cfg)
	application.Init(cfg)
	application.InitDb()

	return application, cfg
}

func TestRadiusIntegration(t *testing.T) {
	appCtx, cfg := setupTestEnv(t)
	defer appCtx.Release()

	registry.ResetForTest()
	t.Cleanup(registry.ResetForTest)
	reRegisterVendorParsers()

	// Initialize Radius Service
	radiusService := NewRadiusService(appCtx)
	defer radiusService.Release()
	plugins.InitPlugins(appCtx, radiusService.SessionRepo, radiusService.AccountingRepo)
	authService := NewAuthService(radiusService)
	acctService := NewAcctService(radiusService)

	// Start Auth Server
	go func() {
		if err := ListenRadiusAuthServer(appCtx, authService); err != nil {
			t.Logf("Auth server stopped: %v", err)
		}
	}()

	// Start Acct Server
	go func() {
		if err := ListenRadiusAcctServer(appCtx, acctService); err != nil {
			t.Logf("Acct server stopped: %v", err)
		}
	}()

	// Wait for servers to start
	time.Sleep(time.Second)

	// Create Test Data
	db := appCtx.DB()
	nas := &domain.NetNas{
		ID:         1,
		Identifier: "test-nas",
		Ipaddr:     "10.0.0.1",
		Secret:     "secret",
		VendorCode: "0",
		Status:     common.ENABLED,
		Remark:     "Test NAS",
	}
	if err := db.Create(nas).Error; err != nil {
		t.Fatal(err)
	}

	mikrotikNas := &domain.NetNas{
		ID:         2,
		Identifier: "nas-mikrotik",
		Ipaddr:     "10.0.0.2",
		Secret:     "secret",
		VendorCode: vendors.CodeMikrotik,
		Status:     common.ENABLED,
		Remark:     "Mikrotik Test NAS",
	}
	if err := db.Create(mikrotikNas).Error; err != nil {
		t.Fatal(err)
	}

	huaweiNas := &domain.NetNas{
		ID:         3,
		Identifier: "nas-huawei",
		Ipaddr:     "10.0.0.3",
		Secret:     "secret",
		VendorCode: vendors.CodeHuawei,
		Status:     common.ENABLED,
		Remark:     "Huawei Test NAS",
	}
	if err := db.Create(huaweiNas).Error; err != nil {
		t.Fatal(err)
	}

	user := &domain.RadiusUser{
		ID:         1,
		Username:   "testuser",
		Password:   "password",
		Status:     common.ENABLED,
		ExpireTime: time.Now().Add(24 * time.Hour),
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatal(err)
	}

	mikrotikUser := &domain.RadiusUser{
		ID:         2,
		Username:   "mik-user",
		Password:   "password",
		UpRate:     512,
		DownRate:   2048,
		MacAddr:    "AA:BB:CC:DD:EE:01",
		Status:     common.ENABLED,
		ExpireTime: time.Now().Add(24 * time.Hour),
	}
	if err := db.Create(mikrotikUser).Error; err != nil {
		t.Fatal(err)
	}

	huaweiUser := &domain.RadiusUser{
		ID:         3,
		Username:   "huawei-user",
		Password:   "password",
		UpRate:     1024,
		DownRate:   2048,
		MacAddr:    "AA:BB:CC:DD:EE:02",
		Status:     common.ENABLED,
		ExpireTime: time.Now().Add(24 * time.Hour),
	}
	if err := db.Create(huaweiUser).Error; err != nil {
		t.Fatal(err)
	}

	serverAddr := fmt.Sprintf("127.0.0.1:%d", cfg.Radiusd.AuthPort)
	acctAddr := fmt.Sprintf("127.0.0.1:%d", cfg.Radiusd.AcctPort)

	t.Run("Access-Request Success", func(t *testing.T) {
		packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
		rfc2865.UserName_SetString(packet, "testuser")
		rfc2865.UserPassword_SetString(packet, "password")
		rfc2865.NASIdentifier_SetString(packet, "test-nas")
		rfc2865.NASIPAddress_Set(packet, net.ParseIP("10.0.0.1"))

		response, err := radius.Exchange(context.Background(), packet, serverAddr)
		if err != nil {
			t.Fatal(err)
		}

		if response.Code != radius.CodeAccessAccept {
			t.Errorf("Expected Access-Accept, got %v", response.Code)
		}
	})
	radiusService.ReleaseAuthRateLimit("testuser")

	t.Run("Access-Request Failure (Wrong Password)", func(t *testing.T) {
		packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
		rfc2865.UserName_SetString(packet, "testuser")
		rfc2865.UserPassword_SetString(packet, "wrongpassword")
		rfc2865.NASIdentifier_SetString(packet, "test-nas")
		rfc2865.NASIPAddress_Set(packet, net.ParseIP("10.0.0.1"))

		response, err := radius.Exchange(context.Background(), packet, serverAddr)
		if err != nil {
			t.Fatal(err)
		}

		if response.Code != radius.CodeAccessReject {
			t.Errorf("Expected Access-Reject, got %v", response.Code)
		}
	})

	t.Run("Mikrotik Access-Request With Vendor Attributes", func(t *testing.T) {
		packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
		rfc2865.UserName_SetString(packet, "mik-user")
		rfc2865.UserPassword_SetString(packet, "password")
		rfc2865.NASIdentifier_SetString(packet, "nas-mikrotik")
		rfc2865.NASIPAddress_Set(packet, net.ParseIP("10.0.0.2"))
		rfc2865.NASPort_Set(packet, 2001)
		rfc2865.NASPortType_Set(packet, rfc2865.NASPortType_Value_Ethernet)
		rfc2865.ServiceType_Set(packet, rfc2865.ServiceType_Value_FramedUser)
		rfc2865.CallingStationID_SetString(packet, "AA-BB-CC-DD-EE-01")
		rfc2865.CalledStationID_SetString(packet, "Hotspot")
		rfc2869.NASPortID_SetString(packet, "Gi0/1/0:100.200")
		rfc2865.FramedIPAddress_Set(packet, net.ParseIP("192.168.88.10"))

		response, err := radius.Exchange(context.Background(), packet, serverAddr)
		if err != nil {
			t.Fatal(err)
		}

		if response.Code != radius.CodeAccessAccept {
			t.Fatalf("Expected Access-Accept, got %v", response.Code)
		}

		rateLimit := mikrotik.MikrotikRateLimit_GetString(response)
		expected := "512k/2048k"
		if rateLimit != expected {
			t.Fatalf("unexpected Mikrotik rate limit, want %s got %s", expected, rateLimit)
		}
	})

	t.Run("Huawei Access-Request With Vendor Attributes", func(t *testing.T) {
		packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
		rfc2865.UserName_SetString(packet, "huawei-user")
		rfc2865.UserPassword_SetString(packet, "password")
		rfc2865.NASIdentifier_SetString(packet, "nas-huawei")
		rfc2865.NASIPAddress_Set(packet, net.ParseIP("10.0.0.3"))
		rfc2865.NASPort_Set(packet, 3001)
		rfc2865.NASPortType_Set(packet, rfc2865.NASPortType_Value_Ethernet)
		rfc2865.ServiceType_Set(packet, rfc2865.ServiceType_Value_FramedUser)
		rfc2865.CallingStationID_SetString(packet, "AA-BB-CC-DD-EE-02")
		rfc2865.CalledStationID_SetString(packet, "huawei-ap")
		rfc2869.NASPortID_SetString(packet, "vlanid=310;vlanid2=320;")
		rfc2865.FramedIPAddress_Set(packet, net.ParseIP("10.10.10.20"))

		response, err := radius.Exchange(context.Background(), packet, serverAddr)
		if err != nil {
			t.Fatal(err)
		}

		if response.Code != radius.CodeAccessAccept {
			t.Fatalf("Expected Access-Accept, got %v", response.Code)
		}

		inputAvg := huawei.HuaweiInputAverageRate_Get(response)
		outputAvg := huawei.HuaweiOutputAverageRate_Get(response)
		inputPeak := huawei.HuaweiInputPeakRate_Get(response)
		outputPeak := huawei.HuaweiOutputPeakRate_Get(response)

		if uint32(inputAvg) != uint32(huaweiUser.UpRate*1024) {
			t.Fatalf("unexpected Huawei input average rate, got %d", uint32(inputAvg))
		}
		if uint32(outputAvg) != uint32(huaweiUser.DownRate*1024) {
			t.Fatalf("unexpected Huawei output average rate, got %d", uint32(outputAvg))
		}
		if uint32(inputPeak) != uint32(huaweiUser.UpRate*1024*4) {
			t.Fatalf("unexpected Huawei input peak rate, got %d", uint32(inputPeak))
		}
		if uint32(outputPeak) != uint32(huaweiUser.DownRate*1024*4) {
			t.Fatalf("unexpected Huawei output peak rate, got %d", uint32(outputPeak))
		}
	})

	t.Run("Accounting-Request Start", func(t *testing.T) {
		packet := radius.New(radius.CodeAccountingRequest, []byte("secret"))
		rfc2865.UserName_SetString(packet, "testuser")
		rfc2865.NASIdentifier_SetString(packet, "test-nas")
		rfc2865.NASIPAddress_Set(packet, net.ParseIP("127.0.0.1"))
		rfc2866.AcctStatusType_Set(packet, rfc2866.AcctStatusType_Value_Start)
		rfc2866.AcctSessionID_SetString(packet, "session-1")

		response, err := radius.Exchange(context.Background(), packet, acctAddr)
		if err != nil {
			t.Fatal(err)
		}

		if response.Code != radius.CodeAccountingResponse {
			t.Errorf("Expected Accounting-Response, got %v", response.Code)
		}
	})

	t.Run("Accounting-Request Stop", func(t *testing.T) {
		packet := radius.New(radius.CodeAccountingRequest, []byte("secret"))
		rfc2865.UserName_SetString(packet, "testuser")
		rfc2865.NASIdentifier_SetString(packet, "test-nas")
		rfc2865.NASIPAddress_Set(packet, net.ParseIP("127.0.0.1"))
		rfc2866.AcctStatusType_Set(packet, rfc2866.AcctStatusType_Value_Stop)
		rfc2866.AcctSessionID_SetString(packet, "session-1")
		rfc2866.AcctSessionTime_Set(packet, 60)
		rfc2866.AcctInputOctets_Set(packet, 1000)
		rfc2866.AcctOutputOctets_Set(packet, 2000)

		response, err := radius.Exchange(context.Background(), packet, acctAddr)
		if err != nil {
			t.Fatal(err)
		}

		if response.Code != radius.CodeAccountingResponse {
			t.Errorf("Expected Accounting-Response, got %v", response.Code)
		}
	})
}
