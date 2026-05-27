package benchmark

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name:   "valid default config",
			mutate: func(c *Config) {},
		},
		{
			name: "empty server address",
			mutate: func(c *Config) {
				c.Server.Address = ""
			},
			wantErr: "server address",
		},
		{
			name: "missing secret",
			mutate: func(c *Config) {
				c.Server.Secret = ""
			},
			wantErr: "server secret",
		},
		{
			name: "invalid auth port",
			mutate: func(c *Config) {
				c.Server.AuthPort = 70000
			},
			wantErr: "invalid auth port",
		},
		{
			name: "invalid acct port",
			mutate: func(c *Config) {
				c.Server.AcctPort = 0
			},
			wantErr: "invalid acct port",
		},
		{
			name: "timeout too small",
			mutate: func(c *Config) {
				c.Server.Timeout = 0
			},
			wantErr: "timeout",
		},
		{
			name: "total transactions too small",
			mutate: func(c *Config) {
				c.Load.Total = 0
			},
			wantErr: "total transactions",
		},
		{
			name: "concurrency too small",
			mutate: func(c *Config) {
				c.Load.Concurrency = 0
			},
			wantErr: "concurrency",
		},
		{
			name: "interval too small",
			mutate: func(c *Config) {
				c.Load.Interval = 0
			},
			wantErr: "interval",
		},
		{
			name: "unsupported encryption",
			mutate: func(c *Config) {
				c.Protocol.Encryption = "chap"
			},
			wantErr: "only PAP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewDefaultConfig()
			tt.mutate(cfg)
			err := cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestConfigGetTimeout(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Server.Timeout = 15
	if got := cfg.GetTimeout(); got != 15*time.Second {
		t.Fatalf("expected 15s timeout, got %v", got)
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Server.Address = "10.0.0.1"
	cfg.Server.Secret = "topsecret"
	cfg.Load.Total = 500

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	if err := cfg.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Ensure file exists with non-zero size
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected config file to be non-empty")
	}

	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if !reflect.DeepEqual(cfg, loaded) {
		t.Fatalf("loaded config mismatch: %#v != %#v", loaded, cfg)
	}
}
