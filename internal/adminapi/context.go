package adminapi

import (
	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"gorm.io/gorm"
)

// GetAppContext gets the application context from echo context
func GetAppContext(c echo.Context) app.AppContext {
	return c.Get("appCtx").(app.AppContext)
}

// GetDB gets the database connection from echo context
func GetDB(c echo.Context) *gorm.DB {
	if db, ok := c.Get("db").(*gorm.DB); ok && db != nil {
		return db
	}
	return GetAppContext(c).DB()
}

// GetConfig gets the configuration from echo context
func GetConfig(c echo.Context) *app.ConfigManager {
	return GetAppContext(c).ConfigMgr()
}
