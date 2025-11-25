package benchmark

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all benchmark configuration parameters.
// It can be loaded from YAML files or populated from command-line flags.
type Config struct {
	// Server configuration
	Server struct {
		Address  string `yaml:"address"`   // RADIUS server address
		AuthPort int    `yaml:"auth_port"` // Authentication port (default: 1812)
		AcctPort int    `yaml:"acct_port"` // Accounting port (default: 1813)
		Secret   string `yaml:"secret"`    // RADIUS shared secret
		Timeout  int    `yaml:"timeout"`   // Request timeout in seconds
	} `yaml:"server"`

	// NAS configuration
	NAS struct {
		IP         string `yaml:"ip"`         // NAS-IP-Address
		Identifier string `yaml:"identifier"` // NAS-Identifier
	} `yaml:"nas"`

	// Test user configuration
	User struct {
		DataFile string `yaml:"data_file"` // Path to JSON file with user list
		Username string `yaml:"username"`  // Single test username
		Password string `yaml:"password"`  // Single test password
		IP       string `yaml:"ip"`        // User IP address
		MAC      string `yaml:"mac"`       // User MAC address
	} `yaml:"user"`

	// Load testing configuration
	Load struct {
		Total       int64 `yaml:"total"`       // Total number of transactions
		Concurrency int64 `yaml:"concurrency"` // Concurrent workers
		Interval    int   `yaml:"interval"`    // Statistics reporting interval (seconds)
	} `yaml:"load"`

	// Output configuration
	Output struct {
		CSVFile string `yaml:"csv_file"` // CSV output file path
	} `yaml:"output"`

	// Protocol settings
	Protocol struct {
		Encryption string `yaml:"encryption"` // Authentication method: pap, chap (currently only pap supported)
	} `yaml:"protocol"`
}

// NewDefaultConfig creates a Config with sensible default values.
func NewDefaultConfig() *Config {
	cfg := &Config{}

	// Server defaults
	cfg.Server.Address = "127.0.0.1"
	cfg.Server.AuthPort = 1812
	cfg.Server.AcctPort = 1813
	cfg.Server.Secret = "secret"
	cfg.Server.Timeout = 10

	// NAS defaults
	cfg.NAS.IP = "127.0.0.1"
	cfg.NAS.Identifier = "benchmark-test"

	// User defaults
	cfg.User.DataFile = "bmdata.json"
	cfg.User.Username = "test01"
	cfg.User.Password = "111111"
	cfg.User.IP = "127.0.0.1"
	cfg.User.MAC = "11:11:11:11:11:11"

	// Load testing defaults
	cfg.Load.Total = 100
	cfg.Load.Concurrency = 10
	cfg.Load.Interval = 5

	// Output defaults
	cfg.Output.CSVFile = "benchmark.csv"

	// Protocol defaults
	cfg.Protocol.Encryption = "pap"

	return cfg
}

// LoadFromFile loads configuration from a YAML file.
//
// Parameters:
//   - path: Path to YAML configuration file
//
// Returns:
//   - *Config: Loaded configuration
//   - error: Error if file cannot be read or parsed
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := NewDefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid and returns an error if not.
func (c *Config) Validate() error {
	if c.Server.Address == "" {
		return fmt.Errorf("server address cannot be empty")
	}
	if c.Server.Secret == "" {
		return fmt.Errorf("server secret cannot be empty")
	}
	if c.Server.AuthPort < 1 || c.Server.AuthPort > 65535 {
		return fmt.Errorf("invalid auth port: %d", c.Server.AuthPort)
	}
	if c.Server.AcctPort < 1 || c.Server.AcctPort > 65535 {
		return fmt.Errorf("invalid acct port: %d", c.Server.AcctPort)
	}
	if c.Server.Timeout < 1 {
		return fmt.Errorf("timeout must be at least 1 second")
	}
	if c.Load.Total < 1 {
		return fmt.Errorf("total transactions must be at least 1")
	}
	if c.Load.Concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1")
	}
	if c.Load.Interval < 1 {
		return fmt.Errorf("interval must be at least 1 second")
	}
	if c.Protocol.Encryption != "pap" {
		return fmt.Errorf("only PAP encryption is currently supported (got: %s)", c.Protocol.Encryption)
	}

	return nil
}

// GetTimeout returns the timeout duration.
func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.Server.Timeout) * time.Second
}

// SaveToFile saves the current configuration to a YAML file.
// This is useful for generating template configuration files.
//
// Parameters:
//   - path: Path where to save the configuration file
//
// Returns:
//   - error: Error if file cannot be written
func (c *Config) SaveToFile(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
