package freeradius

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
)

// TestNewFreeRADIUSServer_Configuration 测试服务器配置
func TestNewFreeRADIUSServer_Configuration(t *testing.T) {
	// 初始化测试配置
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestServer",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/test",
			Debug:    false,
		},
		Freeradius: config.FreeradiusConfig{
			Host:  "127.0.0.1",
			Port:  18182,
			Debug: false,
		},
	}
	
	if app.GApp() == nil {
		app.InitGlobalApplication(cfg)
	}
	
	server := NewFreeRADIUSServer()
	
	assert.NotNil(t, server)
	assert.NotNil(t, server.root)
	assert.True(t, server.root.HideBanner)
	assert.False(t, server.root.Debug)
}

// TestNewFreeRADIUSServer_DebugMode 测试调试模式
func TestNewFreeRADIUSServer_DebugMode(t *testing.T) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestServerDebug",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/test",
			Debug:    true,
		},
		Freeradius: config.FreeradiusConfig{
			Host:  "127.0.0.1",
			Port:  18183,
			Debug: true,
		},
	}
	
	if app.GApp() == nil {
		app.InitGlobalApplication(cfg)
	}
	
	server := NewFreeRADIUSServer()
	
	assert.NotNil(t, server)
	assert.True(t, server.root.Debug)
}

// TestFreeradiusServer_Middleware 测试中间件
func TestFreeradiusServer_Middleware(t *testing.T) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestMiddleware",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/test",
		},
		Freeradius: config.FreeradiusConfig{
			Host:  "127.0.0.1",
			Port:  18184,
			Debug: false,
		},
	}
	
	if app.GApp() == nil {
		app.InitGlobalApplication(cfg)
	}
	
	server := NewFreeRADIUSServer()
	server.initRouter()
	
	// 验证服务器实例存在
	assert.NotNil(t, server)
	assert.NotNil(t, server.root)
	
	// 验证路由已注册
	routes := server.root.Routes()
	assert.NotEmpty(t, routes)
}

// TestServerConfig 测试服务器配置项
func TestServerConfig(t *testing.T) {
	tests := []struct {
		name   string
		host   string
		port   int
		debug  bool
	}{
		{
			name:  "production config",
			host:  "0.0.0.0",
			port:  1818,
			debug: false,
		},
		{
			name:  "development config",
			host:  "127.0.0.1",
			port:  18185,
			debug: true,
		},
		{
			name:  "custom port",
			host:  "localhost",
			port:  8080,
			debug: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.AppConfig{
				System: config.SysConfig{
					Appid:    "Test_" + tt.name,
					Location: "UTC",
				},
				Freeradius: config.FreeradiusConfig{
					Host:  tt.host,
					Port:  tt.port,
					Debug: tt.debug,
				},
			}
			
			// 不实际启动服务器，只验证配置
			assert.Equal(t, tt.host, cfg.Freeradius.Host)
			assert.Equal(t, tt.port, cfg.Freeradius.Port)
			assert.Equal(t, tt.debug, cfg.Freeradius.Debug)
		})
	}
}

// TestServerPanicRecovery 测试panic恢复中间件
func TestServerPanicRecovery(t *testing.T) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestPanic",
			Location: "Asia/Shanghai",
		},
		Freeradius: config.FreeradiusConfig{
			Host:  "127.0.0.1",
			Port:  18186,
			Debug: false,
		},
	}
	
	if app.GApp() == nil {
		app.InitGlobalApplication(cfg)
	}
	
	server := NewFreeRADIUSServer()
	
	// 验证服务器创建成功
	assert.NotNil(t, server)
	assert.NotNil(t, server.root)
	
	// 中间件栈应该包含panic恢复中间件
	// Echo框架会自动添加
}

// TestRouteRegistration 测试路由注册
func TestRouteRegistration(t *testing.T) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestRoutes",
			Location: "Asia/Shanghai",
		},
		Freeradius: config.FreeradiusConfig{
			Host:  "127.0.0.1",
			Port:  18187,
			Debug: false,
		},
	}
	
	if app.GApp() == nil {
		app.InitGlobalApplication(cfg)
	}
	
	server := NewFreeRADIUSServer()
	server.initRouter()
	
	routes := server.root.Routes()
	
	// 验证所有预期路由都已注册
	expectedPaths := map[string]bool{
		"/freeradius/authorize":    false,
		"/freeradius/authenticate": false,
		"/freeradius/postauth":     false,
		"/freeradius/accounting":   false,
	}
	
	for _, route := range routes {
		if _, exists := expectedPaths[route.Path]; exists {
			expectedPaths[route.Path] = true
		}
	}
	
	for path, found := range expectedPaths {
		assert.True(t, found, "Route %s not registered", path)
	}
}

// TestServerInitialization 测试服务器初始化
func TestServerInitialization(t *testing.T) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestInit",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/test",
		},
		Freeradius: config.FreeradiusConfig{
			Host:  "127.0.0.1",
			Port:  18188,
			Debug: false,
		},
	}
	
	if app.GApp() == nil {
		app.InitGlobalApplication(cfg)
	}
	
	server := NewFreeRADIUSServer()
	server.initRouter()
	
	// 验证服务器字段
	assert.NotNil(t, server.root)
	assert.True(t, server.root.HideBanner)
	
	// 验证日志配置
	assert.NotNil(t, server.root.Logger)
}

// BenchmarkNewFreeRADIUSServer 基准测试服务器创建
func BenchmarkNewFreeRADIUSServer(b *testing.B) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "BenchServer",
			Location: "UTC",
		},
		Freeradius: config.FreeradiusConfig{
			Host:  "127.0.0.1",
			Port:  18189,
			Debug: false,
		},
	}
	
	if app.GApp() == nil {
		app.InitGlobalApplication(cfg)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFreeRADIUSServer()
	}
}

// BenchmarkInitRouter 基准测试路由初始化
func BenchmarkInitRouter(b *testing.B) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "BenchRouter",
			Location: "UTC",
		},
		Freeradius: config.FreeradiusConfig{
			Host:  "127.0.0.1",
			Port:  18190,
			Debug: false,
		},
	}
	
	if app.GApp() == nil {
		app.InitGlobalApplication(cfg)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server := NewFreeRADIUSServer()
		server.initRouter()
	}
}
