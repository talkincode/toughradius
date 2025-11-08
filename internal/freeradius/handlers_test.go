package freeradius

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
)

// setupTestServer 设置测试服务器
func setupTestServer() *FreeradiusServer {
	// 初始化测试配置
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestFreeRADIUS",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/test",
			Debug:    false,
		},
		Freeradius: config.FreeradiusConfig{
			Host:  "127.0.0.1",
			Port:  18181,
			Debug: false,
		},
	}

	// 初始化应用
	if app.GApp() == nil {
		app.InitGlobalApplication(cfg)
	}

	server := NewFreeRADIUSServer()
	server.initRouter()
	return server
}

// TestNewFreeRADIUSServer 测试创建FreeRADIUS服务器
func TestNewFreeRADIUSServer(t *testing.T) {
	server := setupTestServer()
	assert.NotNil(t, server)
	assert.NotNil(t, server.root)
	assert.True(t, server.root.HideBanner)
}

// TestInitRouter 测试路由初始化
func TestInitRouter(t *testing.T) {
	server := setupTestServer()

	// 验证路由是否正确注册
	routes := server.root.Routes()

	var foundAuthorize, foundAuthenticate, foundPostauth, foundAccounting bool
	for _, route := range routes {
		switch route.Path {
		case "/freeradius/authorize":
			foundAuthorize = true
			assert.Equal(t, http.MethodPost, route.Method)
		case "/freeradius/authenticate":
			foundAuthenticate = true
			assert.Equal(t, http.MethodPost, route.Method)
		case "/freeradius/postauth":
			foundPostauth = true
			assert.Equal(t, http.MethodPost, route.Method)
		case "/freeradius/accounting":
			foundAccounting = true
			assert.Equal(t, http.MethodPost, route.Method)
		}
	}

	assert.True(t, foundAuthorize, "Authorize route not found")
	assert.True(t, foundAuthenticate, "Authenticate route not found")
	assert.True(t, foundPostauth, "Postauth route not found")
	assert.True(t, foundAccounting, "Accounting route not found")
}

// TestFreeradiusAuthorize_UserNotExists 测试用户不存在的认证
func TestFreeradiusAuthorize_UserNotExists(t *testing.T) {
	server := setupTestServer()

	// 创建测试请求
	form := make(url.Values)
	form.Set("username", "nonexistentuser")
	form.Set("nasip", "192.168.1.1")

	req := httptest.NewRequest(http.MethodPost, "/freeradius/authorize", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	c := server.root.NewContext(req, rec)

	// 执行处理函数
	err := server.FreeradiusAuthorize(c)

	// 验证响应
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotImplemented, rec.Code)
}

// TestFreeradiusAuthenticate 测试认证处理
func TestFreeradiusAuthenticate(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/freeradius/authenticate", nil)
	rec := httptest.NewRecorder()
	c := server.root.NewContext(req, rec)

	err := server.FreeradiusAuthenticate(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestFreeradiusPostauth 测试后认证处理
func TestFreeradiusPostauth(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/freeradius/postauth", nil)
	rec := httptest.NewRecorder()
	c := server.root.NewContext(req, rec)

	err := server.FreeradiusPostauth(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestFreeradiusAccounting 测试计费处理
func TestFreeradiusAccounting(t *testing.T) {
	server := setupTestServer()

	// 创建测试请求
	form := make(url.Values)
	form.Set("username", "testuser")
	form.Set("nasip", "192.168.1.1")
	form.Set("nasid", "nas01")
	form.Set("acctSessionId", "test-session")
	form.Set("acctStatusType", "Start")
	form.Set("acctSessionTime", "0")

	req := httptest.NewRequest(http.MethodPost, "/freeradius/accounting", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	c := server.root.NewContext(req, rec)

	err := server.FreeradiusAccounting(c)

	// 计费处理总是返回成功
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestFreeradiusAccounting_Stop 测试停止计费
func TestFreeradiusAccounting_Stop(t *testing.T) {
	server := setupTestServer()

	form := make(url.Values)
	form.Set("username", "testuser")
	form.Set("nasip", "192.168.1.1")
	form.Set("nasid", "nas01")
	form.Set("acctSessionId", "test-session-stop")
	form.Set("acctStatusType", "Stop")
	form.Set("acctSessionTime", "3600")
	form.Set("acctInputOctets", "1024000")
	form.Set("acctOutputOctets", "2048000")

	req := httptest.NewRequest(http.MethodPost, "/freeradius/accounting", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	c := server.root.NewContext(req, rec)

	err := server.FreeradiusAccounting(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestFreeradiusAccounting_Update 测试更新计费
func TestFreeradiusAccounting_Update(t *testing.T) {
	server := setupTestServer()

	form := make(url.Values)
	form.Set("username", "testuser")
	form.Set("nasip", "192.168.1.1")
	form.Set("nasid", "nas01")
	form.Set("acctSessionId", "test-session-update")
	form.Set("acctStatusType", "Update")
	form.Set("acctSessionTime", "1800")
	form.Set("framedIPAddress", "10.0.0.1")
	form.Set("macAddr", "00:11:22:33:44:55")

	req := httptest.NewRequest(http.MethodPost, "/freeradius/accounting", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	c := server.root.NewContext(req, rec)

	err := server.FreeradiusAccounting(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestUserAuthorizationFlow 测试用户授权流程（集成测试）
func TestUserAuthorizationFlow(t *testing.T) {
	// 这个测试需要真实的数据库连接
	// 在实际环境中需要先设置测试数据

	server := setupTestServer()

	// 测试场景：用户不存在
	form := make(url.Values)
	form.Set("username", "nonexistentuser")
	form.Set("nasip", "192.168.1.1")

	req := httptest.NewRequest(http.MethodPost, "/freeradius/authorize", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	c := server.root.NewContext(req, rec)
	err := server.FreeradiusAuthorize(c)

	assert.NoError(t, err)
	// 用户不存在应该返回501错误
	assert.Equal(t, http.StatusNotImplemented, rec.Code)
}

// TestResponseHeadersAndContentType 测试响应头
func TestResponseHeadersAndContentType(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/freeradius/authenticate", nil)
	rec := httptest.NewRecorder()
	c := server.root.NewContext(req, rec)

	err := server.FreeradiusAuthenticate(c)

	assert.NoError(t, err)
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), echo.MIMEApplicationJSON)
}

// BenchmarkFreeradiusAuthenticate 基准测试认证处理
func BenchmarkFreeradiusAuthenticate(b *testing.B) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/freeradius/authenticate", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := server.root.NewContext(req, rec)
		server.FreeradiusAuthenticate(c)
	}
}

// BenchmarkFreeradiusPostauth 基准测试后认证处理
func BenchmarkFreeradiusPostauth(b *testing.B) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/freeradius/postauth", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := server.root.NewContext(req, rec)
		server.FreeradiusPostauth(c)
	}
}

// BenchmarkFreeradiusAccounting 基准测试计费处理
func BenchmarkFreeradiusAccounting(b *testing.B) {
	server := setupTestServer()

	form := make(url.Values)
	form.Set("username", "benchuser")
	form.Set("nasip", "192.168.1.1")
	form.Set("nasid", "nas01")
	form.Set("acctSessionId", "bench-session")
	form.Set("acctStatusType", "Update")
	form.Set("acctSessionTime", "1800")

	reqBody := strings.NewReader(form.Encode())
	req := httptest.NewRequest(http.MethodPost, "/freeradius/accounting", reqBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := server.root.NewContext(req, rec)
		server.FreeradiusAccounting(c)
		reqBody.Seek(0, 0) // 重置请求体
	}
}

// TestAccountingStatusTypes 测试所有计费状态类型
func TestAccountingStatusTypes(t *testing.T) {
	server := setupTestServer()

	statusTypes := []string{"Start", "Stop", "Update", "Alive", "Interim-Update", "Accounting-On", "Accounting-Off"}

	for _, statusType := range statusTypes {
		t.Run("StatusType_"+statusType, func(t *testing.T) {
			form := make(url.Values)
			form.Set("username", "testuser")
			form.Set("nasip", "192.168.1.1")
			form.Set("nasid", "nas01")
			form.Set("acctSessionId", "session-"+statusType)
			form.Set("acctStatusType", statusType)
			form.Set("acctSessionTime", "1800")

			req := httptest.NewRequest(http.MethodPost, "/freeradius/accounting", strings.NewReader(form.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rec := httptest.NewRecorder()

			c := server.root.NewContext(req, rec)
			err := server.FreeradiusAccounting(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
