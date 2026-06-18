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

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(domain.Tables...))

	return &Application{gormDB: db}
}

func TestCheckSuperCreatesDefaultAdmin(t *testing.T) {
	app := newTestApplication(t)

	app.checkSuper()

	var admin domain.SysOpr
	err := app.gormDB.Where("username = ?", "admin").First(&admin).Error
	require.NoError(t, err)

	assert.Equal(t, "super", admin.Level)
	assert.Equal(t, common.ENABLED, admin.Status)
	assert.True(t, common.VerifyPassword("toughradius", admin.Password))
}

func TestCheckSuperRepairsExistingAdmin(t *testing.T) {
	app := newTestApplication(t)

	broken := &domain.SysOpr{
		ID:       common.UUIDint64(),
		Username: "admin",
		Password: "",
		Level:    "operator",
		Status:   common.DISABLED,
	}
	require.NoError(t, app.gormDB.Create(broken).Error)

	app.checkSuper()

	var admin domain.SysOpr
	err := app.gormDB.Where("username = ?", "admin").First(&admin).Error
	require.NoError(t, err)

	assert.Equal(t, "super", admin.Level)
	assert.Equal(t, common.ENABLED, admin.Status)
	assert.True(t, common.VerifyPassword("toughradius", admin.Password))
}

func TestCheckSuperWarnsWhenCreatingDefaultAdmin(t *testing.T) {
	app := newTestApplication(t)
	core, logs := observer.New(zapcore.WarnLevel)
	undo := zap.ReplaceGlobals(zap.New(core))
	defer undo()

	app.checkSuper()

	assert.Equal(t, 1, logs.FilterMessage("default super admin account uses the built-in password").Len())
}

func TestCheckSuperWarnsWhenExistingAdminUsesDefaultPassword(t *testing.T) {
	app := newTestApplication(t)
	password, err := common.HashPassword(defaultSuperPassword)
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

	assert.Equal(t, 1, logs.FilterMessage("default super admin account uses the built-in password").Len())
}
