package adminapi

import (
	"github.com/talkincode/toughradius/v9/internal/app"
)

// Init registers all admin API routes
func Init(appCtx app.AppContext) {
	registerAuthRoutes()
	registerUserRoutes()
	registerDashboardRoutes()
	registerProfileRoutes()
	registerAccountingRoutes()
	registerSessionRoutes()
	registerNASRoutes()
	registerSettingsRoutes()
	registerNodesRoutes()
	registerOperatorsRoutes()
}
