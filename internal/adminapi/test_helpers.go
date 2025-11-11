package adminapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	customValidator "github.com/talkincode/toughradius/v9/pkg/validator"
	"gorm.io/driver/sqlite"
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
// This is a compatibility function for existing tests
func setupTestApp(t *testing.T, db *gorm.DB) {
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

	// We need to inject the test DB into the application
	// Since we removed the global setters, we'll need to work with the app directly
	// For now, let's just initialize normally and the tests will use GetDB(c)
	testApp.Init(cfg)
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

	e := setupTestEcho()

	return testApp.DB(), e, testApp
}

// CreateTestContext creates an echo context with appCtx injected
func CreateTestContext(e *echo.Echo, db *gorm.DB, req *http.Request, rec *httptest.ResponseRecorder, appCtx app.AppContext) echo.Context {
	c := e.NewContext(req, rec)
	c.Set("appCtx", appCtx)
	c.Set("db", db)
	return c
}
