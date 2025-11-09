package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// ListProfiles 获取 RADIUS Profile 列表
// @Summary 获取 RADIUS Profile 列表
// @Tags RadiusProfile
// @Param page query int false "页码"
// @Param perPage query int false "每页数量"
// @Param sort query string false "排序字段"
// @Param order query string false "排序方向"
// @Success 200 {object} ListResponse
// @Router /api/v1/radius-profiles [get]
func ListProfiles(c echo.Context) error {
	db := app.GDB()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	sortField := c.QueryParam("sort")
	order := c.QueryParam("order")
	if sortField == "" {
		sortField = "id"
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	var total int64
	var profiles []domain.RadiusProfile

	query := db.Model(&domain.RadiusProfile{})

	// 支持按名称搜索
	if name := c.QueryParam("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// 支持按状态过滤
	if status := c.QueryParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&profiles)

	return paged(c, profiles, total, page, perPage)
}

// GetProfile 获取单个 RADIUS Profile
// @Summary 获取 RADIUS Profile 详情
// @Tags RadiusProfile
// @Param id path int true "Profile ID"
// @Success 200 {object} domain.RadiusProfile
// @Router /api/v1/radius-profiles/{id} [get]
func GetProfile(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的 Profile ID", nil)
	}

	var profile domain.RadiusProfile
	if err := app.GDB().First(&profile, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Profile 不存在", nil)
	}

	return ok(c, profile)
}

// CreateProfile 创建 RADIUS Profile
// @Summary 创建 RADIUS Profile
// @Tags RadiusProfile
// @Param profile body domain.RadiusProfile true "Profile 信息"
// @Success 201 {object} domain.RadiusProfile
// @Router /api/v1/radius-profiles [post]
func CreateProfile(c echo.Context) error {
	var profile domain.RadiusProfile
	if err := c.Bind(&profile); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析请求参数", err.Error())
	}

	// 验证必填字段
	if profile.Name == "" {
		return fail(c, http.StatusBadRequest, "MISSING_NAME", "Profile 名称不能为空", nil)
	}

	// 检查名称是否已存在
	var count int64
	app.GDB().Model(&domain.RadiusProfile{}).Where("name = ?", profile.Name).Count(&count)
	if count > 0 {
		return fail(c, http.StatusConflict, "NAME_EXISTS", "Profile 名称已存在", nil)
	}

	// 设置默认值
	if profile.Status == "" {
		profile.Status = "enabled"
	}

	if err := app.GDB().Create(&profile).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "创建 Profile 失败", err.Error())
	}

	return ok(c, profile)
}

// UpdateProfile 更新 RADIUS Profile
// @Summary 更新 RADIUS Profile
// @Tags RadiusProfile
// @Param id path int true "Profile ID"
// @Param profile body domain.RadiusProfile true "Profile 信息"
// @Success 200 {object} domain.RadiusProfile
// @Router /api/v1/radius-profiles/{id} [put]
func UpdateProfile(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的 Profile ID", nil)
	}

	var profile domain.RadiusProfile
	if err := app.GDB().First(&profile, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Profile 不存在", nil)
	}

	var updateData domain.RadiusProfile
	if err := c.Bind(&updateData); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析请求参数", err.Error())
	}

	// 验证名称唯一性
	if updateData.Name != "" && updateData.Name != profile.Name {
		var count int64
		app.GDB().Model(&domain.RadiusProfile{}).Where("name = ? AND id != ?", updateData.Name, id).Count(&count)
		if count > 0 {
			return fail(c, http.StatusConflict, "NAME_EXISTS", "Profile 名称已存在", nil)
		}
	}

	// 更新字段
	updates := map[string]interface{}{}
	if updateData.Name != "" {
		updates["name"] = updateData.Name
	}
	if updateData.Status != "" {
		updates["status"] = updateData.Status
	}
	if updateData.AddrPool != "" {
		updates["addr_pool"] = updateData.AddrPool
	}
	if updateData.ActiveNum > 0 {
		updates["active_num"] = updateData.ActiveNum
	}
	if updateData.UpRate >= 0 {
		updates["up_rate"] = updateData.UpRate
	}
	if updateData.DownRate >= 0 {
		updates["down_rate"] = updateData.DownRate
	}
	if updateData.Remark != "" {
		updates["remark"] = updateData.Remark
	}
	if updateData.NodeId > 0 {
		updates["node_id"] = updateData.NodeId
	}

	if err := app.GDB().Model(&profile).Updates(updates).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "更新 Profile 失败", err.Error())
	}

	// 重新查询最新数据
	app.GDB().First(&profile, id)

	return ok(c, profile)
}

// DeleteProfile 删除 RADIUS Profile
// @Summary 删除 RADIUS Profile
// @Tags RadiusProfile
// @Param id path int true "Profile ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/radius-profiles/{id} [delete]
func DeleteProfile(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的 Profile ID", nil)
	}

	// 检查是否有用户在使用此 Profile
	var userCount int64
	app.GDB().Model(&domain.RadiusUser{}).Where("profile_id = ?", id).Count(&userCount)
	if userCount > 0 {
		return fail(c, http.StatusConflict, "IN_USE", "该 Profile 正在被使用，无法删除", map[string]interface{}{
			"user_count": userCount,
		})
	}

	if err := app.GDB().Delete(&domain.RadiusProfile{}, id).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "删除 Profile 失败", err.Error())
	}

	return ok(c, map[string]interface{}{
		"message": "删除成功",
	})
}

// registerProfileRoutes 注册 Profile 路由
func registerProfileRoutes() {
	webserver.ApiGET("/radius-profiles", ListProfiles)
	webserver.ApiGET("/radius-profiles/:id", GetProfile)
	webserver.ApiPOST("/radius-profiles", CreateProfile)
	webserver.ApiPUT("/radius-profiles/:id", UpdateProfile)
	webserver.ApiDELETE("/radius-profiles/:id", DeleteProfile)
}
