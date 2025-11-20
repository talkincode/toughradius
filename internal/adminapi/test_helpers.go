package adminapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	customValidator "github.com/talkincode/toughradius/v9/pkg/validator"
	"gorm.io/gorm"
)

// setupTestEcho creates an Echo instance with a validator
func setupTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = customValidator.NewValidator()
	return e
}

// setupTestDB creates an in-memory test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Automatically migrate common tables
	err = db.AutoMigrate(
		&domain.RadiusProfile{},
		&domain.RadiusUser{},
		&domain.NetNode{},
		&domain.NetNas{},
		&domain.RadiusAccounting{},
		&domain.RadiusOnline{},
		&domain.SysOpr{},
		&domain.SysConfig{},
	)
	require.NoError(t, err)

	return db
}

// setupTestApp creates a test application context and sets it globally
// Returns: app context for injecting into echo context
func setupTestApp(_ *testing.T, db *gorm.DB) app.AppContext {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestApp",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/toughradius-test",
			Debug:    true,
		},
		Web: config.WebConfig{
			Secret: "test-secret-key-for-jwt",
		},
		Database: config.DBConfig{
			Type: "sqlite",
			Name: ":memory:",
		},
	}

	// Create application but don't call Init() which would create a new DB
	testApp := app.NewApplication(cfg)
	testApp.Init(cfg)
	testApp.OverrideDB(db)

	return testApp
}

// CreateTestAppContext creates a test application context with an in-memory SQLite database
// Returns: db, echo instance, and app context
func CreateTestAppContext(t *testing.T) (*gorm.DB, *echo.Echo, app.AppContext) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/toughradius-test",
		},
		Database: config.DBConfig{
			Type: "sqlite",
			Name: ":memory:",
		},
		Web: config.WebConfig{
			Secret: "test-secret-key-for-jwt",
		},
	}

	testApp := app.NewApplication(cfg)
	testApp.Init(cfg)

	// Migrate test tables
	db := testApp.DB()
	err := db.AutoMigrate(
		&domain.RadiusProfile{},
		&domain.RadiusUser{},
		&domain.NetNode{},
		&domain.NetNas{},
		&domain.RadiusAccounting{},
		&domain.RadiusOnline{},
		&domain.SysOpr{},
		&domain.SysConfig{},
	)
	require.NoError(t, err)

	e := setupTestEcho()

	return db, e, testApp
}

// CreateTestContext creates an echo context with appCtx injected
func CreateTestContext(e *echo.Echo, db *gorm.DB, req *http.Request, rec *httptest.ResponseRecorder, appCtx app.AppContext) echo.Context {
	c := e.NewContext(req, rec)
	c.Set("appCtx", appCtx)
	c.Set("db", db)
	// Inject a default super admin for tests that require authentication
	c.Set("current_operator", &domain.SysOpr{
		ID:       1,
		Username: "superadmin",
		Level:    "super",
		Status:   "enabled",
	})
	return c
}

// CreateTestContextWithApp is a helper that combines setupTestEcho, setupTestDB, and setupTestApp
// for backward compatibility with existing tests
func CreateTestContextWithApp(t *testing.T, req *http.Request, rec *httptest.ResponseRecorder) (echo.Context, *gorm.DB, app.AppContext) {
	e := setupTestEcho()
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	c := CreateTestContext(e, db, req, rec, appCtx)
	return c, db, appCtx
}
