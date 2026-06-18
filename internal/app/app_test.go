package app

import (
	"testing"

	"github.com/talkincode/toughradius/v9/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewApplication(t *testing.T) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestApp",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/test",
			Debug:    true,
		},
	}
	app := NewApplication(cfg)
	if app == nil {
		t.Fatal("NewApplication returned nil")
	}
	if app.appConfig != cfg {
		t.Error("Application config not set correctly")
	}
}

func TestApplication_Config(t *testing.T) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestApp",
			Location: "UTC",
		},
	}
	app := NewApplication(cfg)
	retrievedCfg := app.Config()
	if retrievedCfg != cfg {
		t.Error("Config() did not return the correct config")
	}
	if retrievedCfg.System.Appid != "TestApp" {
		t.Errorf("Expected Appid 'TestApp', got '%s'", retrievedCfg.System.Appid)
	}
}

func TestAutoRegisterPopNodeIdConstant(t *testing.T) {
	if AutoRegisterPopNodeId != 999999999 {
		t.Errorf("Expected AutoRegisterPopNodeId to be 999999999, got %d", AutoRegisterPopNodeId)
	}
}

func TestWarnInsecureRuntimeDefaults(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		wantWarns int
	}{
		{
			name:      "empty secret warns",
			secret:    "",
			wantWarns: 1,
		},
		{
			name:      "default placeholder warns",
			secret:    config.DefaultWebSecret,
			wantWarns: 1,
		},
		{
			name:      "custom secret is quiet",
			secret:    "local-test-random-secret",
			wantWarns: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zapcore.WarnLevel)
			undo := zap.ReplaceGlobals(zap.New(core))
			defer undo()

			warnInsecureRuntimeDefaults(&config.AppConfig{
				System: config.SysConfig{Debug: true},
				Logger: config.LogConfig{Mode: "development"},
				Web:    config.WebConfig{Secret: tt.secret},
			})

			if logs.Len() != tt.wantWarns {
				t.Fatalf("expected %d warnings, got %d: %+v", tt.wantWarns, logs.Len(), logs.All())
			}
		})
	}
}

func TestIsProductionRuntime(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.AppConfig
		want bool
	}{
		{
			name: "development",
			cfg: &config.AppConfig{
				System: config.SysConfig{Debug: true},
				Logger: config.LogConfig{Mode: "development"},
			},
			want: false,
		},
		{
			name: "debug disabled",
			cfg: &config.AppConfig{
				System: config.SysConfig{Debug: false},
				Logger: config.LogConfig{Mode: "development"},
			},
			want: true,
		},
		{
			name: "production logger",
			cfg: &config.AppConfig{
				System: config.SysConfig{Debug: true},
				Logger: config.LogConfig{Mode: "production"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isProductionRuntime(tt.cfg); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
