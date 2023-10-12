package app

import (
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/v8/assets"
	"github.com/talkincode/toughradius/v8/common/zaplog"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"github.com/talkincode/toughradius/v8/config"
	"github.com/talkincode/toughradius/v8/models"
	bolt "go.etcd.io/bbolt"
	"gorm.io/gorm"
)

const (
	AutoRegisterPopNodeId int64 = 999999999
)

var app *Application

type Application struct {
	appConfig *config.AppConfig
	gormDB    *gorm.DB
	sched     *cron.Cron
	// tsdb      tstorage.Storage
	cwmpTable *CwmpEventTable
	transDB   *bolt.DB
}

func GApp() *Application {
	return app
}

func GDB() *gorm.DB {
	return app.gormDB
}

func GConfig() *config.AppConfig {
	return app.appConfig
}

// func GTsdb() tstorage.Storage {
// 	return app.tsdb
// }

func InitGlobalApplication(cfg *config.AppConfig) {
	app = NewApplication(cfg)
	app.Init(cfg)
}

func NewApplication(appConfig *config.AppConfig) *Application {
	return &Application{appConfig: appConfig}
}

func (a *Application) Config() *config.AppConfig {
	return a.appConfig
}

func (a *Application) DB() *gorm.DB {
	return a.gormDB
}

func (a *Application) CwmpTable() *CwmpEventTable {
	return a.cwmpTable
}

func (a *Application) Init(cfg *config.AppConfig) {
	loc, err := time.LoadLocation(cfg.System.Location)
	if err != nil {
		log.Error("timezone config error")
	} else {
		time.Local = loc
	}

	zaplog.InitGlobalLogger(zaplog.LogConfig{
		Mode:           cfg.Logger.Mode,
		ConsoleEnable:  true,
		LokiEnable:     cfg.Logger.LokiEnable,
		FileEnable:     cfg.Logger.FileEnable,
		Filename:       cfg.Logger.Filename,
		LokiApi:        cfg.Logger.LokiApi,
		LokiUser:       cfg.Logger.LokiUser,
		LokiPwd:        cfg.Logger.LokiPwd,
		LokiJob:        cfg.Logger.LokiJob,
		QueueSize:      cfg.Logger.QueueSize,
		MetricsHistory: cfg.Logger.MetricsHistory,
		MetricsStorage: cfg.Logger.MetricsStorage,
	})
	switch cfg.Database.Type {
	case "postgres":
		a.gormDB = getPgDatabase(cfg.Database)
	default:
		panic("not support database type")
	}
	// wait for database initialization to complete
	go func() {
		time.Sleep(3 * time.Second)
		a.checkSuper()
		a.checkSettings()
		a.checkDefaultPNode()
	}()

	a.cwmpTable = NewCwmpEventTable()
	a.initJob()
	a.RenderTranslateFiles()
}

func (a *Application) MigrateDB(track bool) (err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			if os.Getenv("GO_DEGUB_TRACE") != "" {
				debug.PrintStack()
			}
			err2, ok := err1.(error)
			if ok {
				err = err2
				log.Error(err2.Error())
			}
		}
	}()
	if track {
		_ = a.gormDB.Debug().Migrator().AutoMigrate(models.Tables...)
	} else {
		_ = a.gormDB.Migrator().AutoMigrate(models.Tables...)
	}
	return nil
}

func (a *Application) DropAll() {
	_ = a.gormDB.Migrator().DropTable(models.Tables...)
}

func (a *Application) InitDb() {
	err := a.gormDB.Migrator().DropTable(models.Tables...)
	err = a.gormDB.Migrator().AutoMigrate(models.Tables...)
	if err != nil {
		log.Error(err)
	}
}

// checkDefaultPNode check default node
func (a *Application) checkDefaultPNode() {
	var pnode models.NetNode
	err := a.gormDB.Where("id=?", AutoRegisterPopNodeId).First(&pnode).Error
	if err != nil {
		a.gormDB.Create(&models.NetNode{
			ID:     AutoRegisterPopNodeId,
			Name:   "default",
			Tags:   "system",
			Remark: "Device auto-registration node",
		})
	}
}

// GetSettingsStringValue Get settings string value
func (a *Application) GetSettingsStringValue(stype string, name string) string {
	var value string
	a.gormDB.Raw("SELECT value FROM sys_config WHERE type = ? and name = ? limit 1", stype, name).Scan(&value)
	return value
}

func (a *Application) GetSettingsInt64Value(stype string, name string) int64 {
	var value = a.GetSettingsStringValue(stype, name)
	return cast.ToInt64(value)
}

func (a *Application) GetSystemTheme() string {
	var value string
	a.gormDB.Raw("SELECT value FROM sys_config WHERE type = 'system' and name = 'SystemTheme' limit 1").Scan(&value)
	if value == "" {
		a.SetSystemTheme("light")
		return "light"
	}
	return value
}

func (a *Application) SetSystemTheme(value string) {
	a.gormDB.Exec("UPDATE sys_config set value = ? WHERE type = 'system' and name = 'SystemTheme'", value)
}

// GetTr069SettingsStringValue Get tr069 settings string value
func (a *Application) GetTr069SettingsStringValue(name string) string {
	return a.GetSettingsStringValue("tr069", name)
}

func (a *Application) GetRadiusSettingsStringValue(name string) string {
	return a.GetSettingsStringValue("radius", name)
}

func (a *Application) GetSystemSettingsStringValue(name string) string {
	return a.GetSettingsStringValue("system", name)
}

// BackupDatabase Backup database
func (a *Application) BackupDatabase() error {
	scriptsh := assets.PgdumpShell
	scriptsh = strings.ReplaceAll(scriptsh, "{dbhost}", a.appConfig.Database.Host)
	scriptsh = strings.ReplaceAll(scriptsh, "{dbport}", strconv.FormatInt(int64(a.appConfig.Database.Port), 10))
	scriptsh = strings.ReplaceAll(scriptsh, "{dbuser}", a.appConfig.Database.User)
	scriptsh = strings.ReplaceAll(scriptsh, "{dbpwd}", a.appConfig.Database.Passwd)
	scriptsh = strings.ReplaceAll(scriptsh, "{dbname}", a.appConfig.Database.Name)
	_ = os.WriteFile("/tmp/databackup.sh", []byte(scriptsh), 0777)
	defer os.Remove("/tmp/databackup.sh")
	rbs, err := exec.Command("/bin/sh", "/tmp/databackup.sh").CombinedOutput()
	log.Info(string(rbs))
	if err != nil {
		return err
	}
	return nil
}

// checkAppVersion Check version
func (a *Application) checkAppVersion() {
	cver := a.GetSettingsStringValue("system", "TeamsacsVersion")
	buildVersion := assets.BuildVersion()
	if buildVersion != cver {
		_ = a.gormDB.Exec("UPDATE sys_config SET value = ? WHERE type = ? and name = ?", buildVersion, "system", "TeamsacsVersion")
	}
}

func Release() {
	app.sched.Stop()
	_ = app.transDB.Close()
	zaplog.Release()
}
