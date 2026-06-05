package app

import (
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

func newSaveSettingsTestApp(t *testing.T) *Application {
	t.Helper()

	dsn := fmt.Sprintf("file:savesettings_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.SysConfig{}))

	app := &Application{gormDB: db, appConfig: &config.AppConfig{}}
	cm := &ConfigManager{
		app:     app,
		configs: make(map[string]string),
		schemas: make(map[string]*ConfigSchema),
	}
	cm.register(&ConfigSchema{
		Key:     "radius.EapMethod",
		Type:    TypeString,
		Default: "eap-md5",
		Enum:    []string{"eap-md5", "eap-mschapv2"},
	})
	cm.register(&ConfigSchema{
		Key:     "radius.AccountingHistoryDays",
		Type:    TypeInt,
		Default: "90",
	})
	cm.register(&ConfigSchema{
		Key:     "radius.IgnorePassword",
		Type:    TypeBool,
		Default: "false",
	})
	app.configManager = cm
	return app
}

func TestSaveSettings_PersistsValidValues(t *testing.T) {
	app := newSaveSettingsTestApp(t)

	err := app.SaveSettings(map[string]interface{}{
		"radius.EapMethod":             "eap-mschapv2",
		"radius.AccountingHistoryDays": 30,
		"radius.IgnorePassword":        true,
	})
	require.NoError(t, err)

	// In-memory cache reflects the new values.
	assert.Equal(t, "eap-mschapv2", app.GetSettingsStringValue("radius", "EapMethod"))
	assert.Equal(t, int64(30), app.GetSettingsInt64Value("radius", "AccountingHistoryDays"))
	assert.True(t, app.GetSettingsBoolValue("radius", "IgnorePassword"))

	// Values are persisted to the sys_config table.
	var rec domain.SysConfig
	require.NoError(t, app.gormDB.Where("type = ? AND name = ?", "radius", "EapMethod").First(&rec).Error)
	assert.Equal(t, "eap-mschapv2", rec.Value)
}

func TestSaveSettings_EmptyIsNoop(t *testing.T) {
	app := newSaveSettingsTestApp(t)
	require.NoError(t, app.SaveSettings(nil))
	require.NoError(t, app.SaveSettings(map[string]interface{}{}))
}

func TestSaveSettings_InvalidKeyFormat(t *testing.T) {
	app := newSaveSettingsTestApp(t)

	err := app.SaveSettings(map[string]interface{}{
		"missingdot": "value",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid settings key")
}

func TestSaveSettings_UnknownAndInvalidValuesRejected(t *testing.T) {
	app := newSaveSettingsTestApp(t)

	// Unknown (unregistered) key is rejected.
	err := app.SaveSettings(map[string]interface{}{"radius.DoesNotExist": "x"})
	require.Error(t, err)

	// Enum violation is rejected and leaves the previous value untouched.
	err = app.SaveSettings(map[string]interface{}{"radius.EapMethod": "eap-bogus"})
	require.Error(t, err)
	assert.Equal(t, "eap-md5", app.GetSettingsStringValue("radius", "EapMethod"))
}

func TestSaveSettings_PartialFailureAppliesValidKeys(t *testing.T) {
	app := newSaveSettingsTestApp(t)

	err := app.SaveSettings(map[string]interface{}{
		"radius.EapMethod":    "eap-mschapv2", // valid
		"radius.DoesNotExist": "x",            // invalid
	})
	require.Error(t, err)

	// The valid key is still applied despite the sibling failure.
	assert.Equal(t, "eap-mschapv2", app.GetSettingsStringValue("radius", "EapMethod"))
}
