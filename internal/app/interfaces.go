package app

import (
	"github.com/robfig/cron/v3"
	"github.com/talkincode/toughradius/v9/config"
	"gorm.io/gorm"
)

// DBProvider provides database access
type DBProvider interface {
	DB() *gorm.DB
}

// ConfigProvider provides application configuration
type ConfigProvider interface {
	Config() *config.AppConfig
}

// SettingsProvider provides system settings access
type SettingsProvider interface {
	GetSettingsStringValue(category, key string) string
	GetSettingsInt64Value(category, key string) int64
	GetSettingsBoolValue(category, key string) bool
	SaveSettings(settings map[string]interface{}) error
}

// SchedulerProvider provides task scheduling capability
type SchedulerProvider interface {
	Scheduler() *cron.Cron
}

// ConfigManagerProvider provides configuration manager access
type ConfigManagerProvider interface {
	ConfigMgr() *ConfigManager
}

// AppContext combines all provider interfaces for full application context
// Services should depend on specific providers or this combined interface
type AppContext interface {
	DBProvider
	ConfigProvider
	SettingsProvider
	SchedulerProvider
	ConfigManagerProvider
	
	// Application lifecycle methods
	MigrateDB(track bool) error
	InitDb()
	DropAll()
}
