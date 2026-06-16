package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talkincode/toughradius/v9/internal/domain"
)

func newSettingsTestCtx(t *testing.T, method, target, body string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	db, e, appCtx := CreateTestAppContext(t)
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, target, reader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)
	return c, rec
}

func createSettingRow(t *testing.T, c echo.Context, id int64, typ, name, value string) {
	t.Helper()
	require.NoError(t, GetDB(c).Create(&domain.SysConfig{
		ID:        id,
		Type:      typ,
		Name:      name,
		Value:     value,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}).Error)
}

func TestListSettings(t *testing.T) {
	c, rec := newSettingsTestCtx(t, http.MethodGet, "/api/v1/system/settings", "")
	createSettingRow(t, c, 101, "radius", "alpha", "1")
	createSettingRow(t, c, 102, "radius", "beta", "2")
	createSettingRow(t, c, 103, "web", "gamma", "3")

	require.NoError(t, listSettings(c))
	require.Equal(t, http.StatusOK, rec.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotNil(t, resp.Meta)
	assert.Equal(t, int64(3), resp.Meta.Total)
}

func TestListSettings_TypeFilter(t *testing.T) {
	c, rec := newSettingsTestCtx(t, http.MethodGet, "/api/v1/system/settings?type=radius", "")
	createSettingRow(t, c, 111, "radius", "alpha", "1")
	createSettingRow(t, c, 112, "radius", "beta", "2")
	createSettingRow(t, c, 113, "web", "gamma", "3")

	require.NoError(t, listSettings(c))
	require.Equal(t, http.StatusOK, rec.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotNil(t, resp.Meta)
	assert.Equal(t, int64(2), resp.Meta.Total, "only radius-type settings should match")
}

func TestGetSettings(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		seed           bool
		expectedStatus int
	}{
		{"found", "201", true, http.StatusOK},
		{"invalid id", "not-a-number", false, http.StatusBadRequest},
		{"not found", "999999", false, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := newSettingsTestCtx(t, http.MethodGet, "/api/v1/system/settings/"+tt.id, "")
			if tt.seed {
				createSettingRow(t, c, 201, "radius", "alpha", "1")
			}
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			_ = getSettings(c)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestCreateSettings(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		seedDuplicate  bool
		expectedStatus int
	}{
		{"success", `{"type":"radius","name":"NewKey","value":"v1"}`, false, http.StatusOK},
		{"missing value", `{"type":"radius","name":"NoVal","value":""}`, false, http.StatusBadRequest},
		{"missing type", `{"type":"","name":"NoType","value":"v"}`, false, http.StatusBadRequest},
		{"invalid json", `{not-json`, false, http.StatusBadRequest},
		{"duplicate", `{"type":"radius","name":"DupKey","value":"v"}`, true, http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := newSettingsTestCtx(t, http.MethodPost, "/api/v1/system/settings", tt.body)
			if tt.seedDuplicate {
				createSettingRow(t, c, 301, "radius", "DupKey", "existing")
			}
			_ = createSettings(c)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestUpdateSettings(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		seed           bool
		body           string
		expectedStatus int
	}{
		{"success", "401", true, `{"value":"updated","remark":"r"}`, http.StatusOK},
		{"invalid id", "bad", false, `{"value":"x"}`, http.StatusBadRequest},
		{"not found", "888888", false, `{"value":"x"}`, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := newSettingsTestCtx(t, http.MethodPut, "/api/v1/system/settings/"+tt.id, tt.body)
			if tt.seed {
				createSettingRow(t, c, 401, "radius", "UpdKey", "old")
			}
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			_ = updateSettings(c)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var got domain.SysConfig
				require.NoError(t, GetDB(c).Where("id = ?", 401).First(&got).Error)
				assert.Equal(t, "updated", got.Value)
			}
		})
	}
}

func TestDeleteSettings(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c, rec := newSettingsTestCtx(t, http.MethodDelete, "/api/v1/system/settings/501", "")
		createSettingRow(t, c, 501, "radius", "DelKey", "v")
		c.SetParamNames("id")
		c.SetParamValues("501")

		require.NoError(t, deleteSettings(c))
		assert.Equal(t, http.StatusOK, rec.Code)

		var count int64
		GetDB(c).Model(&domain.SysConfig{}).Where("id = ?", 501).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("invalid id", func(t *testing.T) {
		c, rec := newSettingsTestCtx(t, http.MethodDelete, "/api/v1/system/settings/bad", "")
		c.SetParamNames("id")
		c.SetParamValues("bad")

		_ = deleteSettings(c)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetConfigSchemas(t *testing.T) {
	c, rec := newSettingsTestCtx(t, http.MethodGet, "/api/v1/system/config/schemas", "")
	require.NoError(t, getConfigSchemas(c))
	require.Equal(t, http.StatusOK, rec.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	schemas, ok := resp.Data.([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, schemas, "config schemas should be loaded from the embedded JSON")
}

func TestReloadConfig(t *testing.T) {
	c, rec := newSettingsTestCtx(t, http.MethodPost, "/api/v1/system/config/reload", "")
	require.NoError(t, reloadConfig(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestSettingsRequireAdmin_NegativeCases verifies the authorization guard on the
// settings mutating routes: operator-level accounts are rejected (403) before the
// handler runs, while admin/super accounts are allowed through.
func TestSettingsRequireAdmin_NegativeCases(t *testing.T) {
	guard := requireAdmin()
	body := `{"type":"radius","name":"GuardKey","value":"v"}`

	cases := []struct {
		name           string
		level          string
		expectAllowed  bool
		expectedStatus int
	}{
		{"operator denied", LevelOperator, false, http.StatusForbidden},
		{"unknown level denied", "guest", false, http.StatusForbidden},
		{"admin allowed", LevelAdmin, true, http.StatusOK},
		{"super allowed", LevelSuper, true, http.StatusOK},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := newSettingsTestCtx(t, http.MethodPost, "/api/v1/system/settings", body)
			c.Set("current_operator", &domain.SysOpr{ID: 1, Level: tt.level, Status: "enabled"})

			handlerRan := false
			wrapped := guard(func(c echo.Context) error {
				handlerRan = true
				return createSettings(c)
			})
			require.NoError(t, wrapped(c))

			assert.Equal(t, tt.expectAllowed, handlerRan)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
