package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultAppConfig(t *testing.T) {
	cfg := DefaultAppConfig

	// TestSystem configuration
	if cfg.System.Appid != "ToughRADIUS" {
		t.Errorf("Expected Appid 'ToughRADIUS', got '%s'", cfg.System.Appid)
	}

	if cfg.System.Location != "Asia/Shanghai" {
		t.Errorf("Expected Location 'Asia/Shanghai', got '%s'", cfg.System.Location)
	}

	if cfg.System.Workdir != "/var/toughradius" {
		t.Errorf("Expected Workdir '/var/toughradius', got '%s'", cfg.System.Workdir)
	}

	if !cfg.System.Debug {
		t.Error("Expected Debug to be true")
	}

	// Test Web configuration
	if cfg.Web.Host != "0.0.0.0" {
		t.Errorf("Expected Web.Host '0.0.0.0', got '%s'", cfg.Web.Host)
	}

	if cfg.Web.Port != 1816 {
		t.Errorf("Expected Web.Port 1816, got %d", cfg.Web.Port)
	}

	if cfg.Web.TlsPort != 1817 {
		t.Errorf("Expected Web.TlsPort 1817, got %d", cfg.Web.TlsPort)
	}

	// TestDatabase configuration
	if cfg.Database.Type != "sqlite" {
		t.Errorf("Expected Database.Type 'sqlite', got '%s'", cfg.Database.Type)
	}

	if cfg.Database.Name != "toughradius.db" {
		t.Errorf("Expected Database.Name 'toughradius.db', got '%s'", cfg.Database.Name)
	}

	if cfg.Database.MaxConn != 100 {
		t.Errorf("Expected Database.MaxConn 100, got %d", cfg.Database.MaxConn)
	}

	// Test RADIUS configuration
	if !cfg.Radiusd.Enabled {
		t.Error("Expected Radiusd.Enabled to be true")
	}

	if cfg.Radiusd.AuthPort != 1812 {
		t.Errorf("Expected Radiusd.AuthPort 1812, got %d", cfg.Radiusd.AuthPort)
	}

	if cfg.Radiusd.AcctPort != 1813 {
		t.Errorf("Expected Radiusd.AcctPort 1813, got %d", cfg.Radiusd.AcctPort)
	}

	if cfg.Radiusd.RadsecPort != 2083 {
		t.Errorf("Expected Radiusd.RadsecPort 2083, got %d", cfg.Radiusd.RadsecPort)
	}

	// Test logger configuration
	if cfg.Logger.Mode != "development" {
		t.Errorf("Expected Logger.Mode 'development', got '%s'", cfg.Logger.Mode)
	}

	if !cfg.Logger.FileEnable {
		t.Error("Expected Logger.FileEnable to be true")
	}
}

func TestAppConfigGetters(t *testing.T) {
	cfg := &AppConfig{
		System: SysConfig{
			Workdir: "/test/workdir",
		},
	}

	tests := []struct {
		name     string
		getter   func() string
		expected string
	}{
		{
			name:     "GetLogDir",
			getter:   cfg.GetLogDir,
			expected: "/test/workdir/logs",
		},
		{
			name:     "GetPublicDir",
			getter:   cfg.GetPublicDir,
			expected: "/test/workdir/public",
		},
		{
			name:     "GetPrivateDir",
			getter:   cfg.GetPrivateDir,
			expected: "/test/workdir/private",
		},
		{
			name:     "GetDataDir",
			getter:   cfg.GetDataDir,
			expected: "/test/workdir/data",
		},
		{
			name:     "GetBackupDir",
			getter:   cfg.GetBackupDir,
			expected: "/test/workdir/backup",
		},
		{
			name:     "GetRadsecCaCertPath",
			getter:   cfg.GetRadsecCaCertPath,
			expected: "/test/workdir/private/ca.crt",
		},
		{
			name:     "GetRadsecCertPath",
			getter:   cfg.GetRadsecCertPath,
			expected: "/test/workdir/private/radsec.tls.crt",
		},
		{
			name:     "GetRadsecKeyPath",
			getter:   cfg.GetRadsecKeyPath,
			expected: "/test/workdir/private/radsec.tls.key",
		},
	}

	// Set the RadSec certificate paths (relative path)
	cfg.Radiusd.RadsecCaCert = "private/ca.crt"
	cfg.Radiusd.RadsecCert = "private/radsec.tls.crt"
	cfg.Radiusd.RadsecKey = "private/radsec.tls.key"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.getter()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yml")

	// CreateTestconfig file
	configContent := `
system:
  appid: TestApp
  location: Asia/Tokyo
  workdir: /tmp/test
  debug: false

web:
  host: 127.0.0.1
  port: 8080
  tls_port: 8443
  secret: test-secret

database:
  type: postgres
  host: db.example.com
  port: 5433
  name: testdb
  user: testuser
  passwd: testpass
  max_conn: 50
  idle_conn: 5
  debug: true

radiusd:
  enabled: false
  host: 10.0.0.1
  auth_port: 1912
  acct_port: 1913
  radsec_port: 2084
  radsec_worker: 50
  debug: false

logger:
  mode: production
  file_enable: false
  filename: /tmp/test.log
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Loadconfiguration
	cfg := LoadConfig(configFile)

	// ValidateSystem configuration
	if cfg.System.Appid != "TestApp" {
		t.Errorf("Expected Appid 'TestApp', got '%s'", cfg.System.Appid)
	}

	if cfg.System.Location != "Asia/Tokyo" {
		t.Errorf("Expected Location 'Asia/Tokyo', got '%s'", cfg.System.Location)
	}

	if cfg.System.Debug {
		t.Error("Expected Debug to be false")
	}

	// Validate Web configuration
	if cfg.Web.Host != "127.0.0.1" {
		t.Errorf("Expected Web.Host '127.0.0.1', got '%s'", cfg.Web.Host)
	}

	if cfg.Web.Port != 8080 {
		t.Errorf("Expected Web.Port 8080, got %d", cfg.Web.Port)
	}

	if cfg.Web.Secret != "test-secret" {
		t.Errorf("Expected Web.Secret 'test-secret', got '%s'", cfg.Web.Secret)
	}

	// ValidateDatabase configuration
	if cfg.Database.Type != "postgres" {
		t.Errorf("Expected Database.Type 'postgres', got '%s'", cfg.Database.Type)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("Expected Database.Host 'db.example.com', got '%s'", cfg.Database.Host)
	}

	if cfg.Database.Port != 5433 {
		t.Errorf("Expected Database.Port 5433, got %d", cfg.Database.Port)
	}

	if cfg.Database.Name != "testdb" {
		t.Errorf("Expected Database.Name 'testdb', got '%s'", cfg.Database.Name)
	}

	if cfg.Database.MaxConn != 50 {
		t.Errorf("Expected Database.MaxConn 50, got %d", cfg.Database.MaxConn)
	}

	// Validate RADIUS configuration
	if cfg.Radiusd.Enabled {
		t.Error("Expected Radiusd.Enabled to be false")
	}

	if cfg.Radiusd.AuthPort != 1912 {
		t.Errorf("Expected Radiusd.AuthPort 1912, got %d", cfg.Radiusd.AuthPort)
	}

	if cfg.Radiusd.RadsecWorker != 50 {
		t.Errorf("Expected Radiusd.RadsecWorker 50, got %d", cfg.Radiusd.RadsecWorker)
	}

	// Validate logger configuration
	if cfg.Logger.Mode != "production" {
		t.Errorf("Expected Logger.Mode 'production', got '%s'", cfg.Logger.Mode)
	}

	if cfg.Logger.FileEnable {
		t.Error("Expected Logger.FileEnable to be false")
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	// Loading a nonexistent config file should return the default configuration
	cfg := LoadConfig("/nonexistent/path/config.yml")

	// Validate that the defaults are returned
	if cfg.System.Appid != DefaultAppConfig.System.Appid {
		t.Errorf("Expected default Appid '%s', got '%s'", DefaultAppConfig.System.Appid, cfg.System.Appid)
	}

	if cfg.Web.Port != DefaultAppConfig.Web.Port {
		t.Errorf("Expected default Web.Port %d, got %d", DefaultAppConfig.Web.Port, cfg.Web.Port)
	}
}

func TestEnvVariableOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "env-test-config.yml")

	// Create a base config file
	configContent := `
system:
  appid: TestApp
  workdir: /tmp/test
  debug: false

web:
  host: 127.0.0.1
  port: 8080
  secret: original-secret

database:
  type: sqlite
  host: localhost
  port: 5432
  name: testdb
  user: testuser
  passwd: testpass
  debug: false

radiusd:
  enabled: false
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  debug: false

logger:
  mode: development
  file_enable: false
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variables
	testEnvVars := map[string]string{
		"TOUGHRADIUS_SYSTEM_DEBUG":         "true",
		"TOUGHRADIUS_WEB_HOST":             "192.168.1.1",
		"TOUGHRADIUS_WEB_PORT":             "9090",
		"TOUGHRADIUS_WEB_SECRET":           "env-secret",
		"TOUGHRADIUS_DB_TYPE":              "postgres",
		"TOUGHRADIUS_DB_HOST":              "db.server.com",
		"TOUGHRADIUS_DB_PORT":              "5433",
		"TOUGHRADIUS_DB_DEBUG":             "true",
		"TOUGHRADIUS_RADIUS_ENABLED":       "true",
		"TOUGHRADIUS_RADIUS_AUTHPORT":      "1912",
		"TOUGHRADIUS_RADIUS_ACCTPORT":      "1913",
		"TOUGHRADIUS_RADIUS_DEBUG":         "true",
		"TOUGHRADIUS_LOGGER_MODE":          "production",
		"TOUGHRADIUS_LOGGER_FILE_ENABLE":   "true",
		"TOUGHRADIUS_RADIUS_RADSEC_PORT":   "2084",
		"TOUGHRADIUS_RADIUS_RADSEC_WORKER": "200",
	}

	// Preserve the original environment variables
	originalEnvVars := make(map[string]string)
	for key := range testEnvVars {
		originalEnvVars[key] = os.Getenv(key)
	}

	// Set test-specific environment variables
	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}

	// Restore environment variables after the test
	defer func() {
		for key, value := range originalEnvVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Load the configuration (environment variables should override file settings)
	cfg := LoadConfig(configFile)

	// Validate that environment variables override file values
	if !cfg.System.Debug {
		t.Error("Expected System.Debug to be true (from env)")
	}

	if cfg.Web.Host != "192.168.1.1" {
		t.Errorf("Expected Web.Host '192.168.1.1' (from env), got '%s'", cfg.Web.Host)
	}

	if cfg.Web.Port != 9090 {
		t.Errorf("Expected Web.Port 9090 (from env), got %d", cfg.Web.Port)
	}

	if cfg.Web.Secret != "env-secret" {
		t.Errorf("Expected Web.Secret 'env-secret' (from env), got '%s'", cfg.Web.Secret)
	}

	if cfg.Database.Type != "postgres" {
		t.Errorf("Expected Database.Type 'postgres' (from env), got '%s'", cfg.Database.Type)
	}

	if cfg.Database.Host != "db.server.com" {
		t.Errorf("Expected Database.Host 'db.server.com' (from env), got '%s'", cfg.Database.Host)
	}

	if cfg.Database.Port != 5433 {
		t.Errorf("Expected Database.Port 5433 (from env), got %d", cfg.Database.Port)
	}

	if !cfg.Database.Debug {
		t.Error("Expected Database.Debug to be true (from env)")
	}

	if !cfg.Radiusd.Enabled {
		t.Error("Expected Radiusd.Enabled to be true (from env)")
	}

	if cfg.Radiusd.AuthPort != 1912 {
		t.Errorf("Expected Radiusd.AuthPort 1912 (from env), got %d", cfg.Radiusd.AuthPort)
	}

	if cfg.Radiusd.AcctPort != 1913 {
		t.Errorf("Expected Radiusd.AcctPort 1913 (from env), got %d", cfg.Radiusd.AcctPort)
	}

	if cfg.Radiusd.RadsecPort != 2084 {
		t.Errorf("Expected Radiusd.RadsecPort 2084 (from env), got %d", cfg.Radiusd.RadsecPort)
	}

	if cfg.Radiusd.RadsecWorker != 200 {
		t.Errorf("Expected Radiusd.RadsecWorker 200 (from env), got %d", cfg.Radiusd.RadsecWorker)
	}

	if !cfg.Radiusd.Debug {
		t.Error("Expected Radiusd.Debug to be true (from env)")
	}

	if cfg.Logger.Mode != "production" {
		t.Errorf("Expected Logger.Mode 'production' (from env), got '%s'", cfg.Logger.Mode)
	}

	if !cfg.Logger.FileEnable {
		t.Error("Expected Logger.FileEnable to be true (from env)")
	}
}

func TestInitDirs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &AppConfig{
		System: SysConfig{
			Workdir: tmpDir,
		},
	}

	cfg.initDirs()

	// Validate that the directories were created
	expectedDirs := []string{
		"logs",
		"public",
		"data",
		"data/metrics",
		"private",
		"backup",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(tmpDir, dir)
		info, err := os.Stat(dirPath)
		if err != nil {
			t.Errorf("Directory %s was not created: %v", dir, err)
			continue
		}

		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestSetEnvValue(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		initial  string
		expected string
	}{
		{
			name:     "Set value from env",
			envValue: "test-value",
			initial:  "original",
			expected: "test-value",
		},
		{
			name:     "Keep original when env empty",
			envValue: "",
			initial:  "original",
			expected: "original",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envKey := "TEST_ENV_VALUE"
			defer os.Unsetenv(envKey)

			if tt.envValue != "" {
				os.Setenv(envKey, tt.envValue)
			}

			value := tt.initial
			setEnvValue(envKey, &value)

			if value != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, value)
			}
		})
	}
}

func TestSetEnvBoolValue(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		initial  bool
		expected bool
	}{
		{
			name:     "Set true from 'true'",
			envValue: "true",
			initial:  false,
			expected: true,
		},
		{
			name:     "Set true from '1'",
			envValue: "1",
			initial:  false,
			expected: true,
		},
		{
			name:     "Set true from 'on'",
			envValue: "on",
			initial:  false,
			expected: true,
		},
		{
			name:     "Set false from 'false'",
			envValue: "false",
			initial:  true,
			expected: false,
		},
		{
			name:     "Keep original when env empty",
			envValue: "",
			initial:  true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envKey := "TEST_ENV_BOOL_VALUE"
			defer os.Unsetenv(envKey)

			if tt.envValue != "" {
				os.Setenv(envKey, tt.envValue)
			}

			value := tt.initial
			setEnvBoolValue(envKey, &value)

			if value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, value)
			}
		})
	}
}

func TestSetEnvIntValue(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		initial  int
		expected int
	}{
		{
			name:     "Set value from env",
			envValue: "8080",
			initial:  3000,
			expected: 8080,
		},
		{
			name:     "Keep original when env empty",
			envValue: "",
			initial:  3000,
			expected: 3000,
		},
		{
			name:     "Keep original when env invalid",
			envValue: "invalid",
			initial:  3000,
			expected: 3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envKey := "TEST_ENV_INT_VALUE"
			defer os.Unsetenv(envKey)

			if tt.envValue != "" {
				os.Setenv(envKey, tt.envValue)
			}

			value := tt.initial
			setEnvIntValue(envKey, &value)

			if value != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, value)
			}
		})
	}
}

func TestSetEnvInt64Value(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		initial  int64
		expected int64
	}{
		{
			name:     "Set value from env",
			envValue: "9999999999",
			initial:  1000,
			expected: 9999999999,
		},
		{
			name:     "Keep original when env empty",
			envValue: "",
			initial:  1000,
			expected: 1000,
		},
		{
			name:     "Keep original when env invalid",
			envValue: "invalid",
			initial:  1000,
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envKey := "TEST_ENV_INT64_VALUE"
			defer os.Unsetenv(envKey)

			if tt.envValue != "" {
				os.Setenv(envKey, tt.envValue)
			}

			value := tt.initial
			setEnvInt64Value(envKey, &value)

			if value != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, value)
			}
		})
	}
}

func TestDatabaseConfig(t *testing.T) {
	// Test SQLite configuration
	sqliteCfg := DBConfig{
		Type:     "sqlite",
		Name:     "test.db",
		MaxConn:  10,
		IdleConn: 2,
		Debug:    false,
	}

	if sqliteCfg.Type != "sqlite" {
		t.Errorf("Expected Type 'sqlite', got '%s'", sqliteCfg.Type)
	}

	// Test PostgreSQL configuration
	postgresCfg := DBConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     5432,
		Name:     "mydb",
		User:     "admin",
		Passwd:   "secret",
		MaxConn:  100,
		IdleConn: 10,
		Debug:    true,
	}

	if postgresCfg.Type != "postgres" {
		t.Errorf("Expected Type 'postgres', got '%s'", postgresCfg.Type)
	}

	if postgresCfg.Host != "localhost" {
		t.Errorf("Expected Host 'localhost', got '%s'", postgresCfg.Host)
	}

	if postgresCfg.Port != 5432 {
		t.Errorf("Expected Port 5432, got %d", postgresCfg.Port)
	}
}

func TestCompleteConfigurationCycle(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "complete-test.yml")

	// Create a complete configuration
	configContent := `
system:
  appid: CompleteTest
  location: UTC
  workdir: ` + tmpDir + `
  debug: true

web:
  host: 0.0.0.0
  port: 1816
  tls_port: 1817
  secret: complete-secret

database:
  type: sqlite
  name: complete.db
  max_conn: 100
  idle_conn: 10
  debug: false

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083
  radsec_worker: 100
  debug: true

logger:
  mode: development
  file_enable: true
  filename: ` + filepath.Join(tmpDir, "app.log") + `
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load and validate the configuration
	cfg := LoadConfig(configFile)

	// Validate all getter methods
	if cfg.GetLogDir() != filepath.Join(tmpDir, "logs") {
		t.Errorf("GetLogDir mismatch")
	}

	if cfg.GetPublicDir() != filepath.Join(tmpDir, "public") {
		t.Errorf("GetPublicDir mismatch")
	}

	if cfg.GetPrivateDir() != filepath.Join(tmpDir, "private") {
		t.Errorf("GetPrivateDir mismatch")
	}

	if cfg.GetDataDir() != filepath.Join(tmpDir, "data") {
		t.Errorf("GetDataDir mismatch")
	}

	if cfg.GetBackupDir() != filepath.Join(tmpDir, "backup") {
		t.Errorf("GetBackupDir mismatch")
	}

	// Validate that the directories were created
	dirs := []string{
		cfg.GetLogDir(),
		cfg.GetPublicDir(),
		cfg.GetDataDir(),
		cfg.GetBackupDir(),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", dir)
		}
	}
}

func TestRadsecCertPathsAbsolute(t *testing.T) {
	cfg := &AppConfig{
		System: SysConfig{
			Workdir: "/test/workdir",
		},
		Radiusd: RadiusdConfig{
			RadsecCaCert: "/etc/ssl/certs/ca.crt",
			RadsecCert:   "/etc/ssl/certs/server.crt",
			RadsecKey:    "/etc/ssl/private/server.key",
		},
	}

	// Test that absolute paths are not modified
	if cfg.GetRadsecCaCertPath() != "/etc/ssl/certs/ca.crt" {
		t.Errorf("Expected absolute path '/etc/ssl/certs/ca.crt', got '%s'", cfg.GetRadsecCaCertPath())
	}

	if cfg.GetRadsecCertPath() != "/etc/ssl/certs/server.crt" {
		t.Errorf("Expected absolute path '/etc/ssl/certs/server.crt', got '%s'", cfg.GetRadsecCertPath())
	}

	if cfg.GetRadsecKeyPath() != "/etc/ssl/private/server.key" {
		t.Errorf("Expected absolute path '/etc/ssl/private/server.key', got '%s'", cfg.GetRadsecKeyPath())
	}
}

func TestRadsecCertPathsRelative(t *testing.T) {
	cfg := &AppConfig{
		System: SysConfig{
			Workdir: "/var/toughradius",
		},
		Radiusd: RadiusdConfig{
			RadsecCaCert: "certs/ca.crt",
			RadsecCert:   "certs/server.crt",
			RadsecKey:    "certs/server.key",
		},
	}

	// Test that relative paths are joined to the workdir
	if cfg.GetRadsecCaCertPath() != "/var/toughradius/certs/ca.crt" {
		t.Errorf("Expected path '/var/toughradius/certs/ca.crt', got '%s'", cfg.GetRadsecCaCertPath())
	}

	if cfg.GetRadsecCertPath() != "/var/toughradius/certs/server.crt" {
		t.Errorf("Expected path '/var/toughradius/certs/server.crt', got '%s'", cfg.GetRadsecCertPath())
	}

	if cfg.GetRadsecKeyPath() != "/var/toughradius/certs/server.key" {
		t.Errorf("Expected path '/var/toughradius/certs/server.key', got '%s'", cfg.GetRadsecKeyPath())
	}
}

func TestDefaultRadsecCertPaths(t *testing.T) {
	cfg := DefaultAppConfig

	// Validate the default RadSec certificate paths in the configuration
	if cfg.Radiusd.RadsecCaCert != "private/ca.crt" {
		t.Errorf("Expected default RadsecCaCert 'private/ca.crt', got '%s'", cfg.Radiusd.RadsecCaCert)
	}

	if cfg.Radiusd.RadsecCert != "private/radsec.tls.crt" {
		t.Errorf("Expected default RadsecCert 'private/radsec.tls.crt', got '%s'", cfg.Radiusd.RadsecCert)
	}

	if cfg.Radiusd.RadsecKey != "private/radsec.tls.key" {
		t.Errorf("Expected default RadsecKey 'private/radsec.tls.key', got '%s'", cfg.Radiusd.RadsecKey)
	}
}
