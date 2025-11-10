package app

import (
	"os"
	"runtime/debug"
	"time"
	_ "time/tzdata"

	"github.com/robfig/cron/v3"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/metrics"
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
	appConfig     *config.AppConfig
	gormDB        *gorm.DB
	sched         *cron.Cron
	configManager *ConfigManager
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

	// 初始化配置管理器
	a.configManager = NewConfigManager(a)

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
	_ = a.gormDB.Migrator().DropTable(domain.Tables...)
	err := a.gormDB.Migrator().AutoMigrate(domain.Tables...)
	if err != nil {
		zap.S().Error(err)
	}
}

// ConfigMgr 获取配置管理器
func (a *Application) ConfigMgr() *ConfigManager {
	return a.configManager
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

func Release() {
	if app.sched != nil {
		app.sched.Stop()
	}

	_ = metrics.Close()
	_ = zap.L().Sync()
}
