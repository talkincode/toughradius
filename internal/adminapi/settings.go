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

// 系统设置请求结构
type settingsPayload struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	Sort   int    `json:"sort"`
	Remark string `json:"remark"`
}

// 注册系统设置路由
func registerSettingsRoutes() {
	webserver.ApiGET("/system/settings", listSettings)
	webserver.ApiGET("/system/settings/:id", getSettings)
	webserver.ApiPOST("/system/settings", createSettings)
	webserver.ApiPUT("/system/settings/:id", updateSettings)
	webserver.ApiDELETE("/system/settings/:id", deleteSettings)
	webserver.ApiPOST("/system/config/reload", reloadConfig)
}

// 获取系统设置列表
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

// 获取单个系统设置
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

// 创建系统设置
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

	// 检查是否已存在相同类型和名称的配置
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

	// 同步到 ConfigManager 内存缓存
	if app.GApp().ConfigMgr() != nil {
		app.GApp().ConfigMgr().ReloadAll()
	}

	return ok(c, setting)
}

// 更新系统设置
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

	// 更新字段
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

	// 同步到 ConfigManager 内存缓存
	if app.GApp().ConfigMgr() != nil {
		app.GApp().ConfigMgr().ReloadAll()
	}

	return ok(c, setting)
}

// 删除系统设置
func deleteSettings(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的设置 ID", nil)
	}

	if err := app.GDB().Where("id = ?", id).Delete(&domain.SysConfig{}).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "删除系统设置失败", err.Error())
	}

	// 同步到 ConfigManager 内存缓存
	if app.GApp().ConfigMgr() != nil {
		app.GApp().ConfigMgr().ReloadAll()
	}

	return ok(c, map[string]interface{}{
		"id": id,
	})
}

// 筛选条件
func applySettingsFilters(db *gorm.DB, c echo.Context) *gorm.DB {
	if settingType := strings.TrimSpace(c.QueryParam("type")); settingType != "" {
		db = db.Where("type = ?", settingType)
	}

	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		db = db.Where("name ILIKE ?", "%"+name+"%")
	}

	return db
}

// 重载配置
func reloadConfig(c echo.Context) error {
	// 重新加载所有配置
	app.GApp().ConfigMgr().ReloadAll()

	return ok(c, map[string]interface{}{
		"message": "配置已重新加载",
		"time":    time.Now(),
	})
}
