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

type Application struct {
	appConfig     *config.AppConfig
	gormDB        *gorm.DB
	sched         *cron.Cron
	configManager *ConfigManager
	profileCache  *ProfileCache
}

// Ensure Application implements all interfaces
var (
	_ DBProvider            = (*Application)(nil)
	_ ConfigProvider        = (*Application)(nil)
	_ SettingsProvider      = (*Application)(nil)
	_ SchedulerProvider     = (*Application)(nil)
	_ ConfigManagerProvider = (*Application)(nil)
	_ AppContext            = (*Application)(nil)
)

func NewApplication(appConfig *config.AppConfig) *Application {
	return &Application{appConfig: appConfig}
}

func (a *Application) Config() *config.AppConfig {
	return a.appConfig
}

func (a *Application) DB() *gorm.DB {
	return a.gormDB
}

// OverrideDB replaces the application's database handle (used in tests).
func (a *Application) OverrideDB(db *gorm.DB) {
	a.gormDB = db
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
	a.gormDB = getDatabase(cfg.Database, cfg.System.Workdir)
	zap.S().Infof("Database connection successful, type: %s", cfg.Database.Type)

	// Ensure database schema is migrated before loading configs
	if err := a.MigrateDB(false); err != nil {
		zap.S().Errorf("database migration failed: %v", err)
	}

	// wait for database initialization to complete
	go func() {
		time.Sleep(3 * time.Second)
		a.checkSuper()
		a.checkSettings()
		a.checkDefaultPNode()
	}()

	// Initialize the configuration manager
	a.configManager = NewConfigManager(a)

	// Initialize profile cache for dynamic profile linking
	a.profileCache = NewProfileCache(a.gormDB, DefaultProfileCacheTTL)

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

// ConfigMgr returns the configuration manager
func (a *Application) ConfigMgr() *ConfigManager {
	return a.configManager
}

// Scheduler returns the cron scheduler
func (a *Application) Scheduler() *cron.Cron {
	return a.sched
}

// GetSettingsStringValue retrieves a string configuration value
func (a *Application) GetSettingsStringValue(category, key string) string {
	return a.configManager.GetString(category, key)
}

// GetSettingsInt64Value retrieves an int64 configuration value
func (a *Application) GetSettingsInt64Value(category, key string) int64 {
	return a.configManager.GetInt64(category, key)
}

// GetSettingsBoolValue retrieves a boolean configuration value
func (a *Application) GetSettingsBoolValue(category, key string) bool {
	return a.configManager.GetBool(category, key)
}

// SaveSettings saves configuration settings
func (a *Application) SaveSettings(settings map[string]interface{}) error {
	// TODO: Implement proper settings save logic
	// This is a placeholder to satisfy the interface
	return nil
}

// ProfileCache returns the profile cache instance
func (a *Application) ProfileCache() *ProfileCache {
	return a.profileCache
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

// Release releases application resources
func (a *Application) Release() {
	if a.sched != nil {
		a.sched.Stop()
	}

	if a.profileCache != nil {
		a.profileCache.Stop()
	}

	_ = metrics.Close()
	_ = zap.L().Sync()
}
