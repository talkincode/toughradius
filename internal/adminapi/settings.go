package adminapi

import (
	"errors"
	"net/http"
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

	base := app.GDB().Model(&domain.SysConfig{})
	base = applySettingsFilters(base, c)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询系统设置失败", err.Error())
	}

	var settings []domain.SysConfig
	if err := base.
		Order("sort ASC, id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&settings).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询系统设置失败", err.Error())
	}

	return paged(c, settings, total, page, pageSize)
}

// getSettings retrieves a single system setting
func getSettings(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的设置 ID", nil)
	}

	var setting domain.SysConfig
	if err := app.GDB().Where("id = ?", id).First(&setting).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "SETTING_NOT_FOUND", "设置不存在", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询系统设置失败", err.Error())
	}

	return ok(c, setting)
}

// getconfigurationschemas
func getConfigSchemas(c echo.Context) error {
	if app.GApp().ConfigMgr() == nil {
		return fail(c, http.StatusInternalServerError, "CONFIG_MANAGER_NOT_FOUND", "配置管理器未初始化", nil)
	}

	schemas := app.GApp().ConfigMgr().GetAllSchemas()

	// Convert to a frontend-friendly format
	var result []map[string]interface{}
	for key, schema := range schemas {
		schemaData := map[string]interface{}{
			"key":         key,
			"type":        getConfigTypeName(schema.Type),
			"default":     schema.Default,
			"description": schema.Description,
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
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析设置参数", nil)
	}

	payload.Type = strings.TrimSpace(payload.Type)
	payload.Name = strings.TrimSpace(payload.Name)

	if payload.Type == "" || payload.Name == "" || payload.Value == "" {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "type、name、value 不能为空", nil)
	}

	// Check whether a setting with the same type and name already exists (unique constraint)
	var exists int64
	app.GDB().Model(&domain.SysConfig{}).
		Where("type = ? AND name = ?", payload.Type, payload.Name).
		Count(&exists)
	if exists > 0 {
		return fail(c, http.StatusConflict, "SETTING_EXISTS", "该类型下已存在同名配置", nil)
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

	if err := app.GDB().Create(&setting).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "创建系统设置失败", err.Error())
	}

	// Sync to the ConfigManager cache
	if app.GApp().ConfigMgr() != nil {
		app.GApp().ConfigMgr().ReloadAll()
	}

	return ok(c, setting)
}

// updateSettings updates a system setting
func updateSettings(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的设置 ID", nil)
	}

	var payload settingsPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析设置参数", nil)
	}

	var setting domain.SysConfig
	if err := app.GDB().Where("id = ?", id).First(&setting).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "SETTING_NOT_FOUND", "设置不存在", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询系统设置失败", err.Error())
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

	if err := app.GDB().Save(&setting).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "更新系统设置失败", err.Error())
	}

	// Sync to the ConfigManager cache
	if app.GApp().ConfigMgr() != nil {
		app.GApp().ConfigMgr().ReloadAll()
	}

	return ok(c, setting)
}

// deleteSettings deletes a system setting
func deleteSettings(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的设置 ID", nil)
	}

	if err := app.GDB().Where("id = ?", id).Delete(&domain.SysConfig{}).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "删除系统设置失败", err.Error())
	}

	// Sync to the ConfigManager cache
	if app.GApp().ConfigMgr() != nil {
		app.GApp().ConfigMgr().ReloadAll()
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
	app.GApp().ConfigMgr().ReloadAll()

	return ok(c, map[string]interface{}{
		"message": "配置已重新加载",
		"time":    time.Now(),
	})
}
