// Package adminapi implements the ToughRADIUS management REST API served under
// /api/v1. Its handlers back the React Admin frontend and expose CRUD and
// action endpoints for operators, NAS devices, network nodes, RADIUS profiles
// and users, online sessions, accounting records, the dashboard, system
// settings, logs, and backups.
//
// Handlers are plain [github.com/labstack/echo/v4] HandlerFuncs registered with
// the shared web server (see internal/webserver). They read their dependencies
// — the application context, the GORM database handle, and the configuration
// manager — from the echo context using [GetAppContext], [GetDB], and
// [GetConfig], rather than holding injected state, so a handler is a stateless
// function of its request.
//
// Successful responses use the unified [Response] envelope (a data object with
// optional pagination [Meta]); failures use [ErrorResponse] with a stable,
// machine-readable error code. Authorization is enforced per route with the
// [RequireLevel] middleware against the operator levels [LevelSuper],
// [LevelAdmin], and [LevelOperator].
package adminapi

import (
	"github.com/talkincode/toughradius/v9/internal/app"
)

// Init registers every admin API route group on the shared web server. It wires
// the handlers in this package to their paths under /api/v1 and must be called
// once during startup, after the application context is ready and before the
// web server begins serving. It is not safe for concurrent use and is not meant
// to be called more than once.
//
// The appCtx parameter is accepted for symmetry with other subsystems' Init
// functions; handlers resolve the live application context per request from the
// echo context (see [GetAppContext]) rather than capturing it here.
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
	registerCertificateRoutes()
	registerSystemLogRoutes()
	registerSystemBackupRoutes()
}
