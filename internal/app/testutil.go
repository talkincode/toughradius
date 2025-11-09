package app

import (
	"gorm.io/gorm"
)

// SetGApp 设置全局应用实例（仅用于测试）
func SetGApp(a *Application) {
	app = a
}

// SetGDB 设置全局数据库实例（仅用于测试）
func SetGDB(db *gorm.DB) {
	if app != nil {
		app.gormDB = db
	}
}
