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

// ProfileRequest 用于处理前端发送的混合类型 JSON
type ProfileRequest struct {
	Name       string      `json:"name" validate:"required,min=1,max=100"`
	Status     interface{} `json:"status"` // 可以是 string 或 boolean
	AddrPool   string      `json:"addr_pool" validate:"omitempty,addrpool"`
	ActiveNum  int         `json:"active_num" validate:"gte=0,lte=100"`
	UpRate     int         `json:"up_rate" validate:"gte=0,lte=10000000"`
	DownRate   int         `json:"down_rate" validate:"gte=0,lte=10000000"`
	Domain     string      `json:"domain" validate:"omitempty,max=50"`
	IPv6Prefix string      `json:"ipv6_prefix" validate:"omitempty"`
	BindMac    interface{} `json:"bind_mac"`  // 可以是 int 或 boolean
	BindVlan   interface{} `json:"bind_vlan"` // 可以是 int 或 boolean
	Remark     string      `json:"remark" validate:"omitempty,max=500"`
	NodeId     interface{} `json:"node_id"` // 可以是 int64 或 string
}

// toRadiusProfile 将 ProfileRequest 转换为 RadiusProfile
func (pr *ProfileRequest) toRadiusProfile() *domain.RadiusProfile {
	profile := &domain.RadiusProfile{
		Name:       pr.Name,
		AddrPool:   pr.AddrPool,
		ActiveNum:  pr.ActiveNum,
		UpRate:     pr.UpRate,
		DownRate:   pr.DownRate,
		Domain:     pr.Domain,
		IPv6Prefix: pr.IPv6Prefix,
		Remark:     pr.Remark,
	}

	// 处理 status 字段：boolean true -> "enabled", false -> "disabled", string 保持不变
	switch v := pr.Status.(type) {
	case bool:
		if v {
			profile.Status = "enabled"
		} else {
			profile.Status = "disabled"
		}
	case string:
		profile.Status = v
	}

	// 处理 bind_mac 字段：boolean -> int (true=1, false=0)
	switch v := pr.BindMac.(type) {
	case bool:
		if v {
			profile.BindMac = 1
		} else {
			profile.BindMac = 0
		}
	case float64:
		profile.BindMac = int(v)
	}

	// 处理 bind_vlan 字段：boolean -> int (true=1, false=0)
	switch v := pr.BindVlan.(type) {
	case bool:
		if v {
			profile.BindVlan = 1
		} else {
			profile.BindVlan = 0
		}
	case float64:
		profile.BindVlan = int(v)
	}

	// 处理 node_id 字段
	switch v := pr.NodeId.(type) {
	case float64:
		profile.NodeId = int64(v)
	case string:
		if v != "" {
			nodeId, _ := strconv.ParseInt(v, 10, 64)
			profile.NodeId = nodeId
		}
	}

	return profile
}

// CreateProfile 创建 RADIUS Profile
// @Summary 创建 RADIUS Profile
// @Tags RadiusProfile
// @Param profile body ProfileRequest true "Profile 信息"
// @Success 201 {object} domain.RadiusProfile
// @Router /api/v1/radius-profiles [post]
func CreateProfile(c echo.Context) error {
	var req ProfileRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析请求参数", err.Error())
	}

	// 自动验证请求参数
	if err := c.Validate(&req); err != nil {
		return err // 验证错误已经格式化
	}

	// 转换为 RadiusProfile
	profile := req.toRadiusProfile()

	// 检查名称是否已存在（业务逻辑验证）
	var count int64
	app.GDB().Model(&domain.RadiusProfile{}).Where("name = ?", profile.Name).Count(&count)
	if count > 0 {
		return fail(c, http.StatusConflict, "NAME_EXISTS", "Profile 名称已存在", nil)
	}

	// 设置默认值
	if profile.Status == "" {
		profile.Status = "enabled"
	}

	if err := app.GDB().Create(profile).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "创建 Profile 失败", err.Error())
	}

	return ok(c, profile)
}

// UpdateProfile 更新 RADIUS Profile
// @Summary 更新 RADIUS Profile
// @Tags RadiusProfile
// @Param id path int true "Profile ID"
// @Param profile body ProfileRequest true "Profile 信息"
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

	var req ProfileRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析请求参数", err.Error())
	}

	// 自动验证请求参数
	if err := c.Validate(&req); err != nil {
		return err // 验证错误已经格式化
	}

	updateData := req.toRadiusProfile()

	// 验证名称唯一性（业务逻辑验证）
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
	if updateData.ActiveNum >= 0 {
		updates["active_num"] = updateData.ActiveNum
	}
	if updateData.UpRate >= 0 {
		updates["up_rate"] = updateData.UpRate
	}
	if updateData.DownRate >= 0 {
		updates["down_rate"] = updateData.DownRate
	}
	if updateData.Domain != "" {
		updates["domain"] = updateData.Domain
	}
	if updateData.IPv6Prefix != "" {
		updates["ipv6_prefix"] = updateData.IPv6Prefix
	}
	if updateData.BindMac >= 0 {
		updates["bind_mac"] = updateData.BindMac
	}
	if updateData.BindVlan >= 0 {
		updates["bind_vlan"] = updateData.BindVlan
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
