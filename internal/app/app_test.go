package app

import (
	"testing"

	"github.com/talkincode/toughradius/v9/config"
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
