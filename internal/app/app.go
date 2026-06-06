package app

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cast"
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
	// Remove duplicate online rows and drop the legacy non-unique index before
	// AutoMigrate creates the unique index on radius_online.acct_session_id
	// (idempotency backstop, including the upgrade path).
	a.dedupOnlineSessions()
	a.dropLegacyOnlineSessionIndex()
	if track {
		if mErr := a.gormDB.Debug().Migrator().AutoMigrate(domain.Tables...); mErr != nil {
			zap.S().Error(mErr)
			return mErr
		}
	} else {
		if mErr := a.gormDB.Migrator().AutoMigrate(domain.Tables...); mErr != nil {
			zap.S().Error(mErr)
			return mErr
		}
	}
	return nil
}

// dropLegacyOnlineSessionIndex removes the pre-existing non-unique index on
// radius_online.acct_session_id created by older schema versions (the column
// used gorm:"index", named idx_radius_online_acct_session_id). AutoMigrate will
// not convert that index to unique, so it must be dropped first; AutoMigrate
// then creates the new unique index (udx_radius_online_acct_session_id) that the
// ON CONFLICT idempotency clause depends on. Best-effort: failures are logged.
func (a *Application) dropLegacyOnlineSessionIndex() {
	if a.gormDB == nil {
		return
	}
	m := a.gormDB.Migrator()
	if !m.HasTable(&domain.RadiusOnline{}) {
		return
	}
	const legacyIdx = "idx_radius_online_acct_session_id"
	if !m.HasIndex(&domain.RadiusOnline{}, legacyIdx) {
		return
	}
	if err := m.DropIndex(&domain.RadiusOnline{}, legacyIdx); err != nil {
		zap.L().Warn("drop legacy radius_online index failed",
			zap.String("namespace", "radius"), zap.Error(err))
	}
}

// dedupOnlineSessions removes duplicate radius_online rows that share the same
// Acct-Session-Id before AutoMigrate creates the unique index on that column.
// Existing deployments may already contain duplicate online rows produced by
// retransmitted Accounting-Start packets; without this cleanup the unique
// index creation would fail and the idempotency guarantee would silently not
// apply. It keeps the row with the smallest id per Acct-Session-Id and is
// best-effort: failures are logged but never abort startup.
func (a *Application) dedupOnlineSessions() {
	if a.gormDB == nil {
		return
	}
	if !a.gormDB.Migrator().HasTable(&domain.RadiusOnline{}) {
		return
	}
	table := domain.RadiusOnline{}.TableName()
	sql := fmt.Sprintf(
		"DELETE FROM %s WHERE id NOT IN (SELECT min_id FROM (SELECT MIN(id) AS min_id FROM %s GROUP BY acct_session_id) AS keep_ids)",
		table, table,
	)
	if err := a.gormDB.Exec(sql).Error; err != nil {
		zap.L().Warn("dedup radius_online before unique index failed",
			zap.String("namespace", "radius"), zap.Error(err))
	}
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

// SaveSettings persists a batch of configuration settings.
//
// Each map key is a fully-qualified configuration key in "category.name" form
// (matching the keys registered in config_schemas.json) and the value is the
// new value, which is rendered to its string representation before being
// written. Every entry is validated and stored through the ConfigManager, which
// updates both the in-memory cache and the sys_config table atomically per key.
//
// If one or more keys fail (unknown key, validation error, or database error),
// the remaining keys are still attempted and a single joined error describing
// every failure is returned.
func (a *Application) SaveSettings(settings map[string]interface{}) error {
	if len(settings) == 0 {
		return nil
	}

	var errs []error
	for key, raw := range settings {
		category, name, ok := strings.Cut(key, ".")
		if !ok || category == "" || name == "" {
			errs = append(errs, fmt.Errorf("invalid settings key %q: expected \"category.name\"", key))
			continue
		}
		if err := a.configManager.Set(category, name, cast.ToString(raw)); err != nil {
			errs = append(errs, fmt.Errorf("save setting %q: %w", key, err))
		}
	}

	return errors.Join(errs...)
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
