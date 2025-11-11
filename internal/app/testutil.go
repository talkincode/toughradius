package app

import (
	"gorm.io/gorm"
)

// SetGApp sets the global application instance (for testing only)
func SetGApp(a *Application) {
	app = a
}

// SetGDB sets the global database instance (for testing only)
func SetGDB(db *gorm.DB) {
	if app != nil {
		app.gormDB = db
	}
}
