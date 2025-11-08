package adminapi

// Init 注册所有管理端 API 路由
func Init() {
	registerAuthRoutes()
	registerUserRoutes()
}
