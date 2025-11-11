package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// ListProfiles retrieves the RADIUS profile list
// @Summary get the RADIUS profile list
// @Tags RadiusProfile
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort direction"
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

	// Support filtering by name
	if name := c.QueryParam("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// Support filtering by status
	if status := c.QueryParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&profiles)

	return paged(c, profiles, total, page, perPage)
}

// GetProfile retrieves a single RADIUS profile
// @Summary get RADIUS profile detail
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

// ProfileRequest represents the mixed-type JSON sent from the frontend
type ProfileRequest struct {
	Name       string      `json:"name" validate:"required,min=1,max=100"`
	Status     interface{} `json:"status"` // Can be string or boolean
	AddrPool   string      `json:"addr_pool" validate:"omitempty,addrpool"`
	ActiveNum  int         `json:"active_num" validate:"gte=0,lte=100"`
	UpRate     int         `json:"up_rate" validate:"gte=0,lte=10000000"`
	DownRate   int         `json:"down_rate" validate:"gte=0,lte=10000000"`
	Domain     string      `json:"domain" validate:"omitempty,max=50"`
	IPv6Prefix string      `json:"ipv6_prefix" validate:"omitempty"`
	BindMac    interface{} `json:"bind_mac"`  // Can be int or boolean
	BindVlan   interface{} `json:"bind_vlan"` // Can be int or boolean
	Remark     string      `json:"remark" validate:"omitempty,max=500"`
	NodeId     interface{} `json:"node_id"` // Can be int64 or string
}

// toRadiusProfile Convert ProfileRequest Convert to RadiusProfile
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

	// Handle status field: boolean true -> "enabled", false -> "disabled", string remains unchanged
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

	// Handle bind_mac field：boolean -> int (true=1, false=0)
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

	// Handle bind_vlan field：boolean -> int (true=1, false=0)
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

	// Handle node_id field
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

// CreateProfile creates a RADIUS profile
// @Summary create a RADIUS profile
// @Tags RadiusProfile
// @Param profile body ProfileRequest true "Profile information"
// @Success 201 {object} domain.RadiusProfile
// @Router /api/v1/radius-profiles [post]
func CreateProfile(c echo.Context) error {
	var req ProfileRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析请求参数", err.Error())
	}

	// Auto-validate request parameters
	if err := c.Validate(&req); err != nil {
		return err // Validation errors already formatted
	}

	// Convert to RadiusProfile
	profile := req.toRadiusProfile()

	// Check whether a profile with the same name already exists (business logic validation)
	var count int64
	app.GDB().Model(&domain.RadiusProfile{}).Where("name = ?", profile.Name).Count(&count)
	if count > 0 {
		return fail(c, http.StatusConflict, "NAME_EXISTS", "Profile 名称已存在", nil)
	}

	// Set default values
	if profile.Status == "" {
		profile.Status = "enabled"
	}

	if err := app.GDB().Create(profile).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "创建 Profile 失败", err.Error())
	}

	return ok(c, profile)
}

// UpdateProfile updates a RADIUS profile
// @Summary update a RADIUS profile
// @Tags RadiusProfile
// @Param id path int true "Profile ID"
// @Param profile body ProfileRequest true "Profile information"
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

	// Auto-validate request parameters
	if err := c.Validate(&req); err != nil {
		return err // Validation errors already formatted
	}

	updateData := req.toRadiusProfile()

	// Validate name uniqueness (business logic validation)
	if updateData.Name != "" && updateData.Name != profile.Name {
		var count int64
		app.GDB().Model(&domain.RadiusProfile{}).Where("name = ? AND id != ?", updateData.Name, id).Count(&count)
		if count > 0 {
			return fail(c, http.StatusConflict, "NAME_EXISTS", "Profile 名称已存在", nil)
		}
	}

	// Update fields
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

	// Re-query latest data
	app.GDB().First(&profile, id)

	return ok(c, profile)
}

// DeleteProfile Delete RADIUS Profile
// @Summary Delete RADIUS Profile
// @Tags RadiusProfile
// @Param id path int true "Profile ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/radius-profiles/{id} [delete]
func DeleteProfile(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的 Profile ID", nil)
	}

	// Check whether any users are currently using this profile
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

// registerProfileRoutes registers profile routes
func registerProfileRoutes() {
	webserver.ApiGET("/radius-profiles", ListProfiles)
	webserver.ApiGET("/radius-profiles/:id", GetProfile)
	webserver.ApiPOST("/radius-profiles", CreateProfile)
	webserver.ApiPUT("/radius-profiles/:id", UpdateProfile)
	webserver.ApiDELETE("/radius-profiles/:id", DeleteProfile)
}
