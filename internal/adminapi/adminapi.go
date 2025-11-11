package adminapi

// Init registers all admin API routes
func Init() {
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
