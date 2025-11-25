package config

import (
	"os"
	"path"
	"strconv"

	"github.com/talkincode/toughradius/v9/pkg/common"
	"gopkg.in/yaml.v3"
)

// DBConfig Database configuration
// Supported database types: postgres, sqlite
type DBConfig struct {
	Type     string `yaml:"type"`      // Database type: postgres or sqlite
	Host     string `yaml:"host"`      // PostgreSQL host address
	Port     int    `yaml:"port"`      // PostgreSQL port
	Name     string `yaml:"name"`      // Database name or SQLite file path
	User     string `yaml:"user"`      // PostgreSQL username
	Passwd   string `yaml:"passwd"`    // PostgreSQL password
	MaxConn  int    `yaml:"max_conn"`  // Maximum connections
	IdleConn int    `yaml:"idle_conn"` // Idle connections
	Debug    bool   `yaml:"debug"`     // Debug mode
}

// SysConfig System configuration
type SysConfig struct {
	Appid    string `yaml:"appid"`
	Location string `yaml:"location"`
	Workdir  string `yaml:"workdir"`
	Debug    bool   `yaml:"debug"`
}

// WebConfig Web server configuration
type WebConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	TlsPort int    `yaml:"tls_port"`
	Secret  string `yaml:"secret"`
}

type RadiusdConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	Host         string `yaml:"host" json:"host"`
	AuthPort     int    `yaml:"auth_port" json:"auth_port"`
	AcctPort     int    `yaml:"acct_port" json:"acct_port"`
	RadsecPort   int    `yaml:"radsec_port" json:"radsec_port"`
	RadsecWorker int    `yaml:"radsec_worker" json:"radsec_worker"`
	RadsecCaCert string `yaml:"radsec_ca_cert" json:"radsec_ca_cert"` // RadSec CA certificate path
	RadsecCert   string `yaml:"radsec_cert" json:"radsec_cert"`       // RadSec server certificate path
	RadsecKey    string `yaml:"radsec_key" json:"radsec_key"`         // RadSec server private key path
	Debug        bool   `yaml:"debug" json:"debug"`
}

type LogConfig struct {
	Mode       string `yaml:"mode"`
	FileEnable bool   `yaml:"file_enable"`
	Filename   string `yaml:"filename"`
}

type AppConfig struct {
	System   SysConfig     `yaml:"system" json:"system"`
	Web      WebConfig     `yaml:"web" json:"web"`
	Database DBConfig      `yaml:"database" json:"database"`
	Radiusd  RadiusdConfig `yaml:"radiusd" json:"radiusd"`
	Logger   LogConfig     `yaml:"logger" json:"logger"`
}

func (c *AppConfig) GetLogDir() string {
	return path.Join(c.System.Workdir, "logs")
}

func (c *AppConfig) GetPublicDir() string {
	return path.Join(c.System.Workdir, "public")
}

func (c *AppConfig) GetPrivateDir() string {
	return path.Join(c.System.Workdir, "private")
}

func (c *AppConfig) GetDataDir() string {
	return path.Join(c.System.Workdir, "data")
}
func (c *AppConfig) GetBackupDir() string {
	return path.Join(c.System.Workdir, "backup")
}

// GetRadsecCaCertPath Returns the full path to the RadSec CA certificate
func (c *AppConfig) GetRadsecCaCertPath() string {
	if path.IsAbs(c.Radiusd.RadsecCaCert) {
		return c.Radiusd.RadsecCaCert
	}
	return path.Join(c.System.Workdir, c.Radiusd.RadsecCaCert)
}

// GetRadsecCertPath Returns the full path to the RadSec server certificate
func (c *AppConfig) GetRadsecCertPath() string {
	if path.IsAbs(c.Radiusd.RadsecCert) {
		return c.Radiusd.RadsecCert
	}
	return path.Join(c.System.Workdir, c.Radiusd.RadsecCert)
}

// GetRadsecKeyPath Returns the full path to the RadSec server private key
func (c *AppConfig) GetRadsecKeyPath() string {
	if path.IsAbs(c.Radiusd.RadsecKey) {
		return c.Radiusd.RadsecKey
	}
	return path.Join(c.System.Workdir, c.Radiusd.RadsecKey)
}

func (c *AppConfig) initDirs() {
	_ = os.MkdirAll(path.Join(c.System.Workdir, "logs"), 0755)         //nolint:errcheck
	_ = os.MkdirAll(path.Join(c.System.Workdir, "public"), 0755)       //nolint:errcheck
	_ = os.MkdirAll(path.Join(c.System.Workdir, "data"), 0755)         //nolint:errcheck
	_ = os.MkdirAll(path.Join(c.System.Workdir, "data/metrics"), 0755) //nolint:errcheck
	_ = os.MkdirAll(path.Join(c.System.Workdir, "private"), 0644)      //nolint:errcheck
	_ = os.MkdirAll(path.Join(c.System.Workdir, "backup"), 0755)       //nolint:errcheck
}

func setEnvValue(name string, val *string) {
	var evalue = os.Getenv(name)
	if evalue != "" {
		*val = evalue
	}
}

func setEnvBoolValue(name string, val *bool) {
	var evalue = os.Getenv(name)
	if evalue != "" {
		*val = evalue == "true" || evalue == "1" || evalue == "on"
	}
}

func setEnvInt64Value(name string, val *int64) {
	var evalue = os.Getenv(name)
	if evalue == "" {
		return
	}

	p, err := strconv.ParseInt(evalue, 10, 64)
	if err == nil {
		*val = p
	}
}
func setEnvIntValue(name string, val *int) {
	var evalue = os.Getenv(name)
	if evalue == "" {
		return
	}

	p, err := strconv.ParseInt(evalue, 10, 64)
	if err == nil {
		*val = int(p)
	}
}

var DefaultAppConfig = &AppConfig{
	System: SysConfig{
		Appid:    "ToughRADIUS",
		Location: "Asia/Shanghai",
		Workdir:  "/var/toughradius",
		Debug:    true,
	},
	Web: WebConfig{
		Host:    "0.0.0.0",
		Port:    1816,
		TlsPort: 1817,
		Secret:  "9b6de5cc-0731-1203-xxtt-0f568ac9da37",
	},
	Database: DBConfig{
		Type:     "sqlite",    // Default to SQLite for development and testing
		Host:     "127.0.0.1", // PostgreSQL configuration (used when type is postgres)
		Port:     5432,
		Name:     "toughradius.db", // SQLite: database filename; PostgreSQL: database name
		User:     "postgres",
		Passwd:   "myroot",
		MaxConn:  100,
		IdleConn: 10,
		Debug:    false,
	},
	Radiusd: RadiusdConfig{
		Enabled:      true,
		Host:         "0.0.0.0",
		AuthPort:     1812,
		AcctPort:     1813,
		RadsecPort:   2083,
		RadsecWorker: 100,
		RadsecCaCert: "private/ca.crt",
		RadsecCert:   "private/radsec.tls.crt",
		RadsecKey:    "private/radsec.tls.key",
		Debug:        true,
	},
	Logger: LogConfig{
		Mode:       "development",
		FileEnable: true,
		Filename:   "/var/toughradius/toughradius.log",
	},
}

func LoadConfig(cfile string) *AppConfig {
	// In development environment, first check if custom config file exists in current directory
	if cfile == "" {
		cfile = "toughradius.yml"
	}
	if !common.FileExists(cfile) {
		cfile = "/etc/toughradius.yml"
	}
	cfg := new(AppConfig)
	if common.FileExists(cfile) {
		data := common.Must2(os.ReadFile(cfile))
		common.Must(yaml.Unmarshal(data.([]byte), cfg))
	} else {
		cfg = DefaultAppConfig
	}

	cfg.initDirs()

	setEnvValue("TOUGHRADIUS_SYSTEM_WORKER_DIR", &cfg.System.Workdir)
	setEnvBoolValue("TOUGHRADIUS_SYSTEM_DEBUG", &cfg.System.Debug)

	// WEB
	setEnvValue("TOUGHRADIUS_WEB_HOST", &cfg.Web.Host)
	setEnvValue("TOUGHRADIUS_WEB_SECRET", &cfg.Web.Secret)
	setEnvIntValue("TOUGHRADIUS_WEB_PORT", &cfg.Web.Port)
	setEnvIntValue("TOUGHRADIUS_WEB_TLS_PORT", &cfg.Web.TlsPort)

	// DB
	setEnvValue("TOUGHRADIUS_DB_TYPE", &cfg.Database.Type)
	setEnvValue("TOUGHRADIUS_DB_HOST", &cfg.Database.Host)
	setEnvValue("TOUGHRADIUS_DB_NAME", &cfg.Database.Name)
	setEnvValue("TOUGHRADIUS_DB_USER", &cfg.Database.User)
	setEnvValue("TOUGHRADIUS_DB_PWD", &cfg.Database.Passwd)
	setEnvIntValue("TOUGHRADIUS_DB_PORT", &cfg.Database.Port)
	setEnvBoolValue("TOUGHRADIUS_DB_DEBUG", &cfg.Database.Debug)

	// toughradius
	setEnvValue("TOUGHRADIUS_RADIUS_HOST", &cfg.Radiusd.Host)
	setEnvIntValue("TOUGHRADIUS_RADIUS_AUTHPORT", &cfg.Radiusd.AuthPort)
	setEnvIntValue("TOUGHRADIUS_RADIUS_ACCTPORT", &cfg.Radiusd.AcctPort)
	setEnvIntValue("TOUGHRADIUS_RADIUS_RADSEC_PORT", &cfg.Radiusd.RadsecPort)
	setEnvIntValue("TOUGHRADIUS_RADIUS_RADSEC_WORKER", &cfg.Radiusd.RadsecWorker)
	setEnvValue("TOUGHRADIUS_RADIUS_RADSEC_CA_CERT", &cfg.Radiusd.RadsecCaCert)
	setEnvValue("TOUGHRADIUS_RADIUS_RADSEC_CERT", &cfg.Radiusd.RadsecCert)
	setEnvValue("TOUGHRADIUS_RADIUS_RADSEC_KEY", &cfg.Radiusd.RadsecKey)
	setEnvBoolValue("TOUGHRADIUS_RADIUS_DEBUG", &cfg.Radiusd.Debug)
	setEnvBoolValue("TOUGHRADIUS_RADIUS_ENABLED", &cfg.Radiusd.Enabled)

	setEnvValue("TOUGHRADIUS_LOGGER_MODE", &cfg.Logger.Mode)
	setEnvBoolValue("TOUGHRADIUS_LOGGER_FILE_ENABLE", &cfg.Logger.FileEnable)

	return cfg
}
