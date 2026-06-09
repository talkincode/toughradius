package adminapi

import (
	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"gorm.io/gorm"
)

// GetAppContext returns the application context that request middleware stored
// on the echo context under the "appCtx" key. It is the entry point handlers use
// to reach shared services (database, configuration, scheduler).
//
// It panics if no application context is present, which indicates the route was
// registered without the middleware that injects it — a programming error rather
// than a runtime condition, so it is surfaced immediately instead of returning a
// nil context that would fault later.
func GetAppContext(c echo.Context) app.AppContext {
	return c.Get("appCtx").(app.AppContext) //nolint:errcheck // type assertion is safe for middleware-set context
}

// GetDB returns the GORM database handle for the current request. It prefers a
// per-request handle stored under the "db" key (used by tests and request-scoped
// transactions) and otherwise falls back to the shared connection from the
// application context, so handlers get a usable [*gorm.DB] either way.
func GetDB(c echo.Context) *gorm.DB {
	if db, ok := c.Get("db").(*gorm.DB); ok && db != nil {
		return db
	}
	return GetAppContext(c).DB()
}

// GetConfig returns the configuration manager for the current request, resolved
// from the application context. It is the handler-facing accessor for reading and
// updating dynamic system settings.
func GetConfig(c echo.Context) *app.ConfigManager {
	return GetAppContext(c).ConfigMgr()
}
