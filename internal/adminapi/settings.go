package adminapi

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// settingsPayload defines the system setting request structure
type settingsPayload struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	Sort   int    `json:"sort"`
	Remark string `json:"remark"`
}

// registerSettingsRoutes registers system setting routes
func registerSettingsRoutes() {
	webserver.ApiGET("/system/settings", listSettings)
	webserver.ApiGET("/system/settings/:id", getSettings)
	webserver.ApiGET("/system/config/schemas", getConfigSchemas)
	webserver.ApiPOST("/system/settings", createSettings)
	webserver.ApiPUT("/system/settings/:id", updateSettings)
	webserver.ApiDELETE("/system/settings/:id", deleteSettings)
	webserver.ApiPOST("/system/config/reload", reloadConfig)
}

// listSettings retrieves the system settings list
func listSettings(c echo.Context) error {
	page, pageSize := parsePagination(c)

	base := GetDB(c).Model(&domain.SysConfig{})
	base = applySettingsFilters(base, c)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query system settings", err.Error())
	}

	var settings []domain.SysConfig
	if err := base.
		Order("sort ASC, id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&settings).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query system settings", err.Error())
	}

	return paged(c, settings, total, page, pageSize)
}

// getSettings retrieves a single system setting
func getSettings(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid setting ID", nil)
	}

	var setting domain.SysConfig
	if err := GetDB(c).Where("id = ?", id).First(&setting).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "SETTING_NOT_FOUND", "Setting not found", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query system settings", err.Error())
	}

	return ok(c, setting)
}

// getconfigurationschemas
func getConfigSchemas(c echo.Context) error {
	if GetAppContext(c).ConfigMgr() == nil {
		return fail(c, http.StatusInternalServerError, "CONFIG_MANAGER_NOT_FOUND", "Configuration manager is not initialized", nil)
	}

	schemas := GetAppContext(c).ConfigMgr().GetAllSchemas()

	// Convert to a frontend-friendly format
	var result []map[string]interface{}
	keys := make([]string, 0, len(schemas))
	for key := range schemas {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		schema := schemas[key]
		schemaData := map[string]interface{}{
			"key":         key,
			"type":        getConfigTypeName(schema.Type),
			"default":     schema.Default,
			"description": schema.Description,
		}

		if schema.Title != "" {
			schemaData["title"] = schema.Title
		}
		if schema.TitleI18n != "" {
			schemaData["title_i18n"] = schema.TitleI18n
		}
		if schema.DescI18n != "" {
			schemaData["description_i18n"] = schema.DescI18n
		}

		if len(schema.Enum) > 0 {
			schemaData["enum"] = schema.Enum
		}
		if schema.Min != nil {
			schemaData["min"] = *schema.Min
		}
		if schema.Max != nil {
			schemaData["max"] = *schema.Max
		}

		result = append(result, schemaData)
	}

	return ok(c, result)
}

// getConfigTypeName resolves configuration type names
func getConfigTypeName(configType app.ConfigType) string {
	switch configType {
	case app.TypeString:
		return "string"
	case app.TypeInt:
		return "int"
	case app.TypeBool:
		return "bool"
	case app.TypeDuration:
		return "duration"
	case app.TypeJSON:
		return "json"
	default:
		return "string"
	}
}

// createSettings creates a system setting
func createSettings(c echo.Context) error {
	var payload settingsPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse setting parameters", nil)
	}

	payload.Type = strings.TrimSpace(payload.Type)
	payload.Name = strings.TrimSpace(payload.Name)

	if payload.Type == "" || payload.Name == "" || payload.Value == "" {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "type, name, and value cannot be empty", nil)
	}

	// Check whether a setting with the same type and name already exists (unique constraint)
	var exists int64
	GetDB(c).Model(&domain.SysConfig{}).
		Where("type = ? AND name = ?", payload.Type, payload.Name).
		Count(&exists)
	if exists > 0 {
		return fail(c, http.StatusConflict, "SETTING_EXISTS", "A setting with the same name already exists under this type", nil)
	}

	setting := domain.SysConfig{
		ID:        common.UUIDint64(),
		Type:      payload.Type,
		Name:      payload.Name,
		Value:     payload.Value,
		Sort:      payload.Sort,
		Remark:    payload.Remark,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := GetDB(c).Create(&setting).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create system setting", err.Error())
	}

	// Sync to the ConfigManager cache
	if GetAppContext(c).ConfigMgr() != nil {
		GetAppContext(c).ConfigMgr().ReloadAll()
	}

	return ok(c, setting)
}

// updateSettings updates a system setting
func updateSettings(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid setting ID", nil)
	}

	var payload settingsPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse setting parameters", nil)
	}

	var setting domain.SysConfig
	if err := GetDB(c).Where("id = ?", id).First(&setting).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "SETTING_NOT_FOUND", "Setting not found", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query system settings", err.Error())
	}

	// Update fields
	if payload.Type != "" {
		setting.Type = strings.TrimSpace(payload.Type)
	}
	if payload.Name != "" {
		setting.Name = strings.TrimSpace(payload.Name)
	}
	if payload.Value != "" {
		setting.Value = payload.Value
	}
	if payload.Sort != 0 {
		setting.Sort = payload.Sort
	}
	if payload.Remark != "" {
		setting.Remark = payload.Remark
	}
	setting.UpdatedAt = time.Now()

	if err := GetDB(c).Save(&setting).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update system setting", err.Error())
	}

	// Sync to the ConfigManager cache
	if GetAppContext(c).ConfigMgr() != nil {
		GetAppContext(c).ConfigMgr().ReloadAll()
	}

	return ok(c, setting)
}

// deleteSettings deletes a system setting
func deleteSettings(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid setting ID", nil)
	}

	if err := GetDB(c).Where("id = ?", id).Delete(&domain.SysConfig{}).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete system setting", err.Error())
	}

	// Sync to the ConfigManager cache
	if GetAppContext(c).ConfigMgr() != nil {
		GetAppContext(c).ConfigMgr().ReloadAll()
	}

	return ok(c, map[string]interface{}{
		"id": id,
	})
}

// Filter conditions
func applySettingsFilters(db *gorm.DB, c echo.Context) *gorm.DB {
	if settingType := strings.TrimSpace(c.QueryParam("type")); settingType != "" {
		db = db.Where("type = ?", settingType)
	}

	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		db = db.Where("name ILIKE ?", "%"+name+"%")
	}

	return db
}

// reloadConfig reloads the configuration
func reloadConfig(c echo.Context) error {
	// Reload all configuration
	GetAppContext(c).ConfigMgr().ReloadAll()

	return ok(c, map[string]interface{}{
		"message": "Configuration reloaded",
		"time":    time.Now(),
	})
}
