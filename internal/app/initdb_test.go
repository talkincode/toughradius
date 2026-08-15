package app

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/gorm"
)

func newTestApplication(t *testing.T) *Application {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(domain.Tables...))

	return &Application{gormDB: db}
}

func TestIsWellKnownBootstrapPassword(t *testing.T) {
	assert.True(t, IsWellKnownBootstrapPassword(WellKnownBootstrapPassword))
	assert.False(t, IsWellKnownBootstrapPassword("unique-pass-123"))
	assert.False(t, IsWellKnownBootstrapPassword(""))
}

func TestCheckSuperCreatesBootstrapAdmin(t *testing.T) {
	app := newTestApplication(t)

	app.checkSuper()

	var admin domain.SysOpr
	err := app.gormDB.Where("username = ?", "admin").First(&admin).Error
	require.NoError(t, err)

	assert.Equal(t, "super", admin.Level)
	assert.Equal(t, common.ENABLED, admin.Status)
	assert.NotEmpty(t, admin.Password)
	assert.False(t, common.VerifyPassword(WellKnownBootstrapPassword, admin.Password))
}

func TestCheckSuperUsesEnvPassword(t *testing.T) {
	t.Setenv(AdminPasswordEnv, "EnvPass123")
	app := newTestApplication(t)

	app.checkSuper()

	var admin domain.SysOpr
	require.NoError(t, app.gormDB.Where("username = ?", "admin").First(&admin).Error)
	assert.True(t, common.VerifyPassword("EnvPass123", admin.Password))
	assert.False(t, common.VerifyPassword(WellKnownBootstrapPassword, admin.Password))
}

func TestCheckSuperIgnoresWellKnownEnvPassword(t *testing.T) {
	t.Setenv(AdminPasswordEnv, WellKnownBootstrapPassword)
	app := newTestApplication(t)

	app.checkSuper()

	var admin domain.SysOpr
	require.NoError(t, app.gormDB.Where("username = ?", "admin").First(&admin).Error)
	assert.False(t, common.VerifyPassword(WellKnownBootstrapPassword, admin.Password))
	assert.NotEmpty(t, admin.Password)
}

func TestCheckSuperDoesNotReenableOrPromote(t *testing.T) {
	app := newTestApplication(t)
	password, err := common.HashPassword("CustomPass123")
	require.NoError(t, err)
	require.NoError(t, app.gormDB.Create(&domain.SysOpr{
		ID:       common.UUIDint64(),
		Username: "admin",
		Password: password,
		Level:    "operator",
		Status:   common.DISABLED,
	}).Error)

	app.checkSuper()

	var admin domain.SysOpr
	require.NoError(t, app.gormDB.Where("username = ?", "admin").First(&admin).Error)
	assert.Equal(t, "operator", admin.Level)
	assert.Equal(t, common.DISABLED, admin.Status)
	assert.True(t, common.VerifyPassword("CustomPass123", admin.Password))
}

func TestCheckSuperRotatesWellKnownPassword(t *testing.T) {
	app := newTestApplication(t)
	password, err := common.HashPassword(WellKnownBootstrapPassword)
	require.NoError(t, err)
	require.NoError(t, app.gormDB.Create(&domain.SysOpr{
		ID:       common.UUIDint64(),
		Username: defaultSuperUsername,
		Password: password,
		Level:    "super",
		Status:   common.ENABLED,
	}).Error)
	core, logs := observer.New(zapcore.WarnLevel)
	undo := zap.ReplaceGlobals(zap.New(core))
	defer undo()

	app.checkSuper()

	var admin domain.SysOpr
	require.NoError(t, app.gormDB.Where("username = ?", defaultSuperUsername).First(&admin).Error)
	assert.False(t, common.VerifyPassword(WellKnownBootstrapPassword, admin.Password))
	assert.Equal(t, 1, logs.FilterMessage("rotated insecure super admin password").Len())
}

func TestCheckSuperRotatesEmptyPassword(t *testing.T) {
	app := newTestApplication(t)
	require.NoError(t, app.gormDB.Create(&domain.SysOpr{
		ID:       common.UUIDint64(),
		Username: "admin",
		Password: "",
		Level:    "operator",
		Status:   common.DISABLED,
	}).Error)

	app.checkSuper()

	var admin domain.SysOpr
	require.NoError(t, app.gormDB.Where("username = ?", "admin").First(&admin).Error)
	assert.Equal(t, "operator", admin.Level)
	assert.Equal(t, common.DISABLED, admin.Status)
	assert.NotEmpty(t, admin.Password)
	assert.False(t, common.VerifyPassword(WellKnownBootstrapPassword, admin.Password))
}

func TestCheckSuperDoesNotCreateWhenOtherSuperExists(t *testing.T) {
	app := newTestApplication(t)
	password, err := common.HashPassword("RootPass123")
	require.NoError(t, err)
	require.NoError(t, app.gormDB.Create(&domain.SysOpr{
		ID:       common.UUIDint64(),
		Username: "root",
		Password: password,
		Level:    "super",
		Status:   common.ENABLED,
	}).Error)

	app.checkSuper()

	var count int64
	require.NoError(t, app.gormDB.Model(&domain.SysOpr{}).Where("username = ?", "admin").Count(&count).Error)
	assert.Zero(t, count)
}

func TestCheckSuperLogsGeneratedPasswordOnce(t *testing.T) {
	app := newTestApplication(t)
	core, logs := observer.New(zapcore.InfoLevel)
	undo := zap.ReplaceGlobals(zap.New(core))
	defer undo()

	app.checkSuper()

	entries := logs.FilterMessage("initialized bootstrap super admin account")
	assert.Equal(t, 1, entries.Len())
	assert.NotEmpty(t, entries.All()[0].ContextMap()["password"])
}

func TestGenerateBootstrapAdminPassword(t *testing.T) {
	password, err := generateBootstrapAdminPassword()
	require.NoError(t, err)
	assert.Len(t, password, bootstrapPasswordLen)
	assert.False(t, IsWellKnownBootstrapPassword(password))

	hasLetter, hasDigit := false, false
	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z':
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	assert.True(t, hasLetter)
	assert.True(t, hasDigit)
}
