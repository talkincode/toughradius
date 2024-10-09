package config

import (
	"os"
	"path"
	"strconv"

	"github.com/talkincode/toughradius/v8/common"
	"gopkg.in/yaml.v3"
)

// DBConfig 数据库(PostgreSQL)配置
type DBConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Passwd   string `yaml:"passwd"`
	MaxConn  int    `yaml:"max_conn"`
	IdleConn int    `yaml:"idle_conn"`
	Debug    bool   `yaml:"debug"`
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
	Debug        bool   `yaml:"debug" json:"debug"`
}

// Tr069Config tr069 API 配置
type Tr069Config struct {
	Host   string `yaml:"host" json:"host"`
	Port   int    `yaml:"port" json:"port"`
	Tls    bool   `yaml:"tls" json:"tls"`
	Secret string `yaml:"secret" json:"secret"`
	Debug  bool   `yaml:"debug" json:"debug"`
}

type MqttConfig struct {
	Server   string `yaml:"server" json:"server"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Debug    bool   `yaml:"debug" json:"debug"`
}

type LogConfig struct {
	Mode           string `yaml:"mode"`
	ConsoleEnable  bool   `yaml:"console_enable"`
	LokiEnable     bool   `yaml:"loki_enable"`
	FileEnable     bool   `yaml:"file_enable"`
	Filename       string `yaml:"filename"`
	QueueSize      int    `yaml:"queue_size"`
	LokiApi        string `yaml:"loki_api"`
	LokiUser       string `yaml:"loki_user"`
	LokiPwd        string `yaml:"loki_pwd"`
	LokiJob        string `yaml:"loki_job"`
	MetricsStorage string `yaml:"metrics_storage"`
	MetricsHistory int    `yaml:"metrics_history"`
}

type AppConfig struct {
	System     SysConfig        `yaml:"system" json:"system"`
	Web        WebConfig        `yaml:"web" json:"web"`
	Database   DBConfig         `yaml:"database" json:"database"`
	Freeradius FreeradiusConfig `yaml:"freeradius" json:"freeradius"`
	Radiusd    RadiusdConfig    `yaml:"radiusd" json:"radiusd"`
	Tr069      Tr069Config      `yaml:"tr069" json:"tr069"`
	Mqtt       MqttConfig       `yaml:"mqtt" json:"mqtt"`
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

func (c *AppConfig) initDirs() {
	os.MkdirAll(path.Join(c.System.Workdir, "logs"), 0755)
	os.MkdirAll(path.Join(c.System.Workdir, "public"), 0755)
	os.MkdirAll(path.Join(c.System.Workdir, "data"), 0755)
	os.MkdirAll(path.Join(c.System.Workdir, "data/metrics"), 0755)
	os.MkdirAll(path.Join(c.System.Workdir, "private"), 0644)
	os.MkdirAll(path.Join(c.System.Workdir, "backup"), 0644)
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
		Type:     "postgres",
		Host:     "127.0.0.1",
		Port:     5432,
		Name:     "toughradius_v8",
		User:     "postgres",
		Passwd:   "myroot",
		MaxConn:  100,
		IdleConn: 10,
		Debug:    false,
	},
	Tr069: Tr069Config{
		Host:   "0.0.0.0",
		Tls:    true,
		Port:   1819,
		Secret: "9b6de5cc-0731-1203-xxtt-0f568ac9da37",
		Debug:  true,
	},
	Mqtt: MqttConfig{
		Server:   "",
		Username: "",
		Password: "",
		Debug:    false,
	},
	Freeradius: FreeradiusConfig{
		Enabled: true,
		Host:    "0.0.0.0",
		Port:    1818,
		Debug:   true,
	},
	Radiusd: RadiusdConfig{
		Enabled:    true,
		Host:       "0.0.0.0",
		AuthPort:   1812,
		AcctPort:   1813,
		RadsecPort: 2083,
		RadsecWorker: 100,
		Debug:      true,
	},
	Logger: LogConfig{
		Mode:           "development",
		ConsoleEnable:  true,
		LokiEnable:     false,
		FileEnable:     true,
		Filename:       "/var/toughradius/toughradius.log",
		QueueSize:      4096,
		LokiApi:        "http://127.0.0.1:3100",
		LokiUser:       "toughradius",
		LokiPwd:        "toughradius",
		LokiJob:        "toughradius",
		MetricsStorage: "/var/toughradius/data/metrics",
		MetricsHistory: 24 * 7,
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
	setEnvBoolValue("TOUGHRADIUS_RADIUS_DEBUG", &cfg.Radiusd.Debug)
	setEnvBoolValue("TOUGHRADIUS_RADIUS_ENABLED", &cfg.Radiusd.Enabled)

	// FreeRADIUS Config
	setEnvValue("TOUGHRADIUS_FREERADIUS_WEB_HOST", &cfg.Freeradius.Host)
	setEnvIntValue("TOUGHRADIUS_FREERADIUS_WEB_PORT", &cfg.Freeradius.Port)
	setEnvBoolValue("TOUGHRADIUS_FREERADIUS_WEB_DEBUG", &cfg.Freeradius.Debug)
	setEnvBoolValue("TOUGHRADIUS_FREERADIUS_WEB_ENABLED", &cfg.Freeradius.Enabled)

	// TR069 Config
	setEnvValue("TOUGHRADIUS_TR069_WEB_HOST", &cfg.Tr069.Host)
	setEnvValue("TOUGHRADIUS_TR069_WEB_SECRET", &cfg.Tr069.Secret)
	setEnvBoolValue("TOUGHRADIUS_TR069_WEB_TLS", &cfg.Tr069.Tls)
	setEnvBoolValue("TOUGHRADIUS_TR069_WEB_DEBUG", &cfg.Tr069.Debug)
	setEnvIntValue("TOUGHRADIUS_TR069_WEB_PORT", &cfg.Tr069.Port)

	setEnvValue("TOUGHRADIUS_LOKI_JOB", &cfg.Logger.LokiJob)
	setEnvValue("TOUGHRADIUS_LOKI_SERVER", &cfg.Logger.LokiApi)
	setEnvValue("TOUGHRADIUS_LOKI_USERNAME", &cfg.Logger.LokiUser)
	setEnvValue("TOUGHRADIUS_LOKI_PASSWORD", &cfg.Logger.LokiPwd)
	setEnvValue("TOUGHRADIUS_LOGGER_MODE", &cfg.Logger.Mode)
	setEnvBoolValue("TOUGHRADIUS_LOKI_ENABLE", &cfg.Logger.LokiEnable)
	setEnvBoolValue("TOUGHRADIUS_LOGGER_FILE_ENABLE", &cfg.Logger.FileEnable)

	setEnvValue("TOUGHRADIUS_MQTT_SERVER", &cfg.Mqtt.Server)
	setEnvValue("TOUGHRADIUS_MQTT_USERNAME", &cfg.Mqtt.Username)
	setEnvValue("TOUGHRADIUS_MQTT_PASSWORD", &cfg.Mqtt.Password)
	setEnvBoolValue("TOUGHRADIUS_MQTT_DEBUG", &cfg.Mqtt.Debug)

	return cfg
}
