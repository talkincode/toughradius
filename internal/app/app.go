package app

import (
	"os"
	"runtime/debug"
	"time"
	_ "time/tzdata"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/metrics"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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

func (a *Application) Init(cfg *config.AppConfig) {
	loc, err := time.LoadLocation(cfg.System.Location)
	if err != nil {
		zap.S().Error("timezone config error")
	} else {
		time.Local = loc
	}

	// Initialize zap logger
	var zapConfig zap.Config
	if cfg.Logger.Mode == "production" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Configure output paths
	zapConfig.OutputPaths = []string{"stdout"}
	if cfg.Logger.FileEnable {
		zapConfig.OutputPaths = append(zapConfig.OutputPaths, cfg.Logger.Filename)
	}

	// Build logger with file rotation if enabled
	var logger *zap.Logger
	if cfg.Logger.FileEnable {
		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.Logger.Filename,
			MaxSize:    64,
			MaxBackups: 7,
			MaxAge:     7,
			Compress:   false,
		}

		core := zapcore.NewTee(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				zapcore.AddSync(lumberJackLogger),
				zapConfig.Level,
			),
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
				zapcore.AddSync(os.Stdout),
				zapConfig.Level,
			),
		)
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		logger, err = zapConfig.Build(zap.AddCaller(), zap.AddCallerSkip(1))
		if err != nil {
			panic(err)
		}
	}

	zap.ReplaceGlobals(logger)

	// Initialize metrics with workdir convention
	err = metrics.InitMetrics(cfg.System.Workdir)
	if err != nil {
		zap.S().Warn("Failed to initialize metrics:", err)
	}

	// Initialize database connection
	if cfg.Database.Type == "" {
		cfg.Database.Type = "postgres"
	}
	a.gormDB = getDatabase(cfg.Database)
	zap.S().Infof("数据库连接成功，类型: %s", cfg.Database.Type)

	// wait for database initialization to complete
	go func() {
		time.Sleep(3 * time.Second)
		a.checkSuper()
		a.checkSettings()
		a.checkDefaultPNode()
	}()

	a.initJob()
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
				zap.S().Error(err2.Error())
			}
		}
	}()
	if track {
		if err := a.gormDB.Debug().Migrator().AutoMigrate(domain.Tables...); err != nil {
			zap.S().Error(err)
		}
	} else {
		if err := a.gormDB.Migrator().AutoMigrate(domain.Tables...); err != nil {
			zap.S().Error(err)
		}
	}
	return nil
}

func (a *Application) DropAll() {
	_ = a.gormDB.Migrator().DropTable(domain.Tables...)
}

func (a *Application) InitDb() {
	err := a.gormDB.Migrator().DropTable(domain.Tables...)
	err = a.gormDB.Migrator().AutoMigrate(domain.Tables...)
	if err != nil {
		zap.S().Error(err)
	}
}

// checkDefaultPNode check default node
func (a *Application) checkDefaultPNode() {
	var pnode domain.NetNode
	err := a.gormDB.Where("id=?", AutoRegisterPopNodeId).First(&pnode).Error
	if err != nil {
		a.gormDB.Create(&domain.NetNode{
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

func (a *Application) GetRadiusSettingsStringValue(name string) string {
	return a.GetSettingsStringValue("radius", name)
}

func (a *Application) GetSystemSettingsStringValue(name string) string {
	return a.GetSettingsStringValue("system", name)
}

func Release() {
	app.sched.Stop()
	_ = app.transDB.Close()
	_ = metrics.Close()
	_ = zap.L().Sync()
}
