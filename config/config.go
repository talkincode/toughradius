package config

import (
	"os"
	"path"
	"strconv"

	"github.com/talkincode/toughradius/v9/pkg/common"
	"gopkg.in/yaml.v3"
)

// DBConfig 数据库配置
// 支持的数据库类型: postgres, sqlite
type DBConfig struct {
	Type     string `yaml:"type"`      // 数据库类型: postgres 或 sqlite
	Host     string `yaml:"host"`      // PostgreSQL 主机地址
	Port     int    `yaml:"port"`      // PostgreSQL 端口
	Name     string `yaml:"name"`      // 数据库名称或 SQLite 文件路径
	User     string `yaml:"user"`      // PostgreSQL 用户名
	Passwd   string `yaml:"passwd"`    // PostgreSQL 密码
	MaxConn  int    `yaml:"max_conn"`  // 最大连接数
	IdleConn int    `yaml:"idle_conn"` // 空闲连接数
	Debug    bool   `yaml:"debug"`     // 调试模式
}

// SysConfig 系统配置
type SysConfig struct {
	Appid    string `yaml:"appid"`
	Location string `yaml:"location"`
	Workdir  string `yaml:"workdir"`
	Debug    bool   `yaml:"debug"`
}

// WebConfig WEB 配置
type WebConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	TlsPort int    `yaml:"tls_port"`
	Secret  string `yaml:"secret"`
}

// FreeradiusConfig Freeradius API 配置
type FreeradiusConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Host    string `yaml:"host" json:"host"`
	Port    int    `yaml:"port" json:"port"`
	Debug   bool   `yaml:"debug" json:"debug"`
}

type RadiusdConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	Host         string `yaml:"host" json:"host"`
	AuthPort     int    `yaml:"auth_port" json:"auth_port"`
	AcctPort     int    `yaml:"acct_port" json:"acct_port"`
	RadsecPort   int    `yaml:"radsec_port" json:"radsec_port"`
	RadsecWorker int    `yaml:"radsec_worker" json:"radsec_worker"`
	RadsecCaCert string `yaml:"radsec_ca_cert" json:"radsec_ca_cert"` // RadSec CA 证书路径
	RadsecCert   string `yaml:"radsec_cert" json:"radsec_cert"`       // RadSec 服务器证书路径
	RadsecKey    string `yaml:"radsec_key" json:"radsec_key"`         // RadSec 服务器私钥路径
	Debug        bool   `yaml:"debug" json:"debug"`
}

type LogConfig struct {
	Mode       string `yaml:"mode"`
	FileEnable bool   `yaml:"file_enable"`
	Filename   string `yaml:"filename"`
}

type AppConfig struct {
	System     SysConfig        `yaml:"system" json:"system"`
	Web        WebConfig        `yaml:"web" json:"web"`
	Database   DBConfig         `yaml:"database" json:"database"`
	Freeradius FreeradiusConfig `yaml:"freeradius" json:"freeradius"`
	Radiusd    RadiusdConfig    `yaml:"radiusd" json:"radiusd"`
	Logger     LogConfig        `yaml:"logger" json:"logger"`
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

// GetRadsecCaCertPath 获取 RadSec CA 证书的完整路径
func (c *AppConfig) GetRadsecCaCertPath() string {
	if path.IsAbs(c.Radiusd.RadsecCaCert) {
		return c.Radiusd.RadsecCaCert
	}
	return path.Join(c.System.Workdir, c.Radiusd.RadsecCaCert)
}

// GetRadsecCertPath 获取 RadSec 服务器证书的完整路径
func (c *AppConfig) GetRadsecCertPath() string {
	if path.IsAbs(c.Radiusd.RadsecCert) {
		return c.Radiusd.RadsecCert
	}
	return path.Join(c.System.Workdir, c.Radiusd.RadsecCert)
}

// GetRadsecKeyPath 获取 RadSec 服务器私钥的完整路径
func (c *AppConfig) GetRadsecKeyPath() string {
	if path.IsAbs(c.Radiusd.RadsecKey) {
		return c.Radiusd.RadsecKey
	}
	return path.Join(c.System.Workdir, c.Radiusd.RadsecKey)
}

func (c *AppConfig) initDirs() {
	os.MkdirAll(path.Join(c.System.Workdir, "logs"), 0755)
	os.MkdirAll(path.Join(c.System.Workdir, "public"), 0755)
	os.MkdirAll(path.Join(c.System.Workdir, "data"), 0755)
	os.MkdirAll(path.Join(c.System.Workdir, "data/metrics"), 0755)
	os.MkdirAll(path.Join(c.System.Workdir, "private"), 0644)
	os.MkdirAll(path.Join(c.System.Workdir, "backup"), 0755)
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
		Type:     "sqlite",    // 默认使用 SQLite，适合开发和测试
		Host:     "127.0.0.1", // PostgreSQL 配置（当 type 为 postgres 时使用）
		Port:     5432,
		Name:     "toughradius.db", // SQLite: 数据库文件名；PostgreSQL: 数据库名
		User:     "postgres",
		Passwd:   "myroot",
		MaxConn:  100,
		IdleConn: 10,
		Debug:    false,
	},
	Freeradius: FreeradiusConfig{
		Enabled: true,
		Host:    "0.0.0.0",
		Port:    1818,
		Debug:   true,
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
	// 开发环境首先查找当前目录是否存在自定义配置文件
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

	// FreeRADIUS Config
	setEnvValue("TOUGHRADIUS_FREERADIUS_WEB_HOST", &cfg.Freeradius.Host)
	setEnvIntValue("TOUGHRADIUS_FREERADIUS_WEB_PORT", &cfg.Freeradius.Port)
	setEnvBoolValue("TOUGHRADIUS_FREERADIUS_WEB_DEBUG", &cfg.Freeradius.Debug)
	setEnvBoolValue("TOUGHRADIUS_FREERADIUS_WEB_ENABLED", &cfg.Freeradius.Enabled)

	setEnvValue("TOUGHRADIUS_LOGGER_MODE", &cfg.Logger.Mode)
	setEnvBoolValue("TOUGHRADIUS_LOGGER_FILE_ENABLE", &cfg.Logger.FileEnable)

	return cfg
}
