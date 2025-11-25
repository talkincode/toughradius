package adminapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
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
	db := GetDB(c)

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

	// Support filtering by name (case-insensitive)
	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		if strings.EqualFold(db.Dialector.Name(), "postgres") {
			query = query.Where("name ILIKE ?", "%"+name+"%")
		} else {
			query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
		}
	}

	// Support filtering by status
	if status := strings.TrimSpace(c.QueryParam("status")); status != "" {
		query = query.Where("status = ?", status)
	}

	// Support filtering by addr_pool (case-insensitive)
	if addrPool := strings.TrimSpace(c.QueryParam("addr_pool")); addrPool != "" {
		if strings.EqualFold(db.Dialector.Name(), "postgres") {
			query = query.Where("addr_pool ILIKE ?", "%"+addrPool+"%")
		} else {
			query = query.Where("LOWER(addr_pool) LIKE ?", "%"+strings.ToLower(addrPool)+"%")
		}
	}

	// Support filtering by domain (case-insensitive)
	if domain := strings.TrimSpace(c.QueryParam("domain")); domain != "" {
		if strings.EqualFold(db.Dialector.Name(), "postgres") {
			query = query.Where("domain ILIKE ?", "%"+domain+"%")
		} else {
			query = query.Where("LOWER(domain) LIKE ?", "%"+strings.ToLower(domain)+"%")
		}
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
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid profile ID", nil)
	}

	var profile domain.RadiusProfile
	if err := GetDB(c).First(&profile, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Profile not found", nil)
	}

	return ok(c, profile)
}

// ProfileRequest represents the mixed-type JSON sent from the frontend
type ProfileRequest struct {
	Name           string      `json:"name" validate:"required,min=1,max=100"`
	Status         interface{} `json:"status"` // Can be string or boolean
	AddrPool       string      `json:"addr_pool" validate:"omitempty,addrpool"`
	ActiveNum      int         `json:"active_num" validate:"gte=0,lte=100"`
	UpRate         int         `json:"up_rate" validate:"gte=0,lte=10000000"`
	DownRate       int         `json:"down_rate" validate:"gte=0,lte=10000000"`
	Domain         string      `json:"domain" validate:"omitempty,max=50"`
	IPv6PrefixPool string      `json:"ipv6_prefix_pool" validate:"omitempty"`
	BindMac        interface{} `json:"bind_mac"`  // Can be int or boolean
	BindVlan       interface{} `json:"bind_vlan"` // Can be int or boolean
	Remark         string      `json:"remark" validate:"omitempty,max=500"`
	NodeId         interface{} `json:"node_id"` // Can be int64 or string
}

// toRadiusProfile Convert ProfileRequest Convert to RadiusProfile
func (pr *ProfileRequest) toRadiusProfile() *domain.RadiusProfile {
	profile := &domain.RadiusProfile{
		Name:           pr.Name,
		AddrPool:       pr.AddrPool,
		ActiveNum:      pr.ActiveNum,
		UpRate:         pr.UpRate,
		DownRate:       pr.DownRate,
		Domain:         pr.Domain,
		IPv6PrefixPool: pr.IPv6PrefixPool,
		Remark:         pr.Remark,
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

// ProfileUpdateRequest represents the mixed-type JSON sent from the frontend for updates
type ProfileUpdateRequest struct {
	Name           string      `json:"name" validate:"omitempty,min=1,max=100"`
	Status         interface{} `json:"status"` // Can be string or boolean
	AddrPool       string      `json:"addr_pool" validate:"omitempty,addrpool"`
	ActiveNum      int         `json:"active_num" validate:"gte=0,lte=100"`
	UpRate         int         `json:"up_rate" validate:"gte=0,lte=10000000"`
	DownRate       int         `json:"down_rate" validate:"gte=0,lte=10000000"`
	Domain         string      `json:"domain" validate:"omitempty,max=50"`
	IPv6PrefixPool string      `json:"ipv6_prefix_pool" validate:"omitempty"`
	BindMac        interface{} `json:"bind_mac"`  // Can be int or boolean
	BindVlan       interface{} `json:"bind_vlan"` // Can be int or boolean
	Remark         string      `json:"remark" validate:"omitempty,max=500"`
	NodeId         interface{} `json:"node_id"` // Can be int64 or string
}

// toRadiusProfile Convert ProfileUpdateRequest Convert to RadiusProfile
func (pr *ProfileUpdateRequest) toRadiusProfile() *domain.RadiusProfile {
	profile := &domain.RadiusProfile{
		Name:           pr.Name,
		AddrPool:       pr.AddrPool,
		ActiveNum:      pr.ActiveNum,
		UpRate:         pr.UpRate,
		DownRate:       pr.DownRate,
		Domain:         pr.Domain,
		IPv6PrefixPool: pr.IPv6PrefixPool,
		Remark:         pr.Remark,
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
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request parameters", err.Error())
	}

	// Auto-validate request parameters
	if err := c.Validate(&req); err != nil {
		return err // Validation errors already formatted
	}

	// Convert to RadiusProfile
	profile := req.toRadiusProfile()

	// Check whether a profile with the same name already exists (business logic validation)
	var count int64
	GetDB(c).Model(&domain.RadiusProfile{}).Where("name = ?", profile.Name).Count(&count)
	if count > 0 {
		return fail(c, http.StatusConflict, "NAME_EXISTS", "Profile name already exists", nil)
	}

	// Set default values
	if profile.Status == "" {
		profile.Status = "enabled"
	}

	if err := GetDB(c).Create(profile).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create profile", err.Error())
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
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid profile ID", nil)
	}

	var profile domain.RadiusProfile
	if err := GetDB(c).First(&profile, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Profile not found", nil)
	}

	var req ProfileUpdateRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request parameters", err.Error())
	}

	// Auto-validate request parameters
	if err := c.Validate(&req); err != nil {
		return err // Validation errors already formatted
	}

	updateData := req.toRadiusProfile()

	// Validate name uniqueness (business logic validation)
	if updateData.Name != "" && updateData.Name != profile.Name {
		var count int64
		GetDB(c).Model(&domain.RadiusProfile{}).Where("name = ? AND id != ?", updateData.Name, id).Count(&count)
		if count > 0 {
			return fail(c, http.StatusConflict, "NAME_EXISTS", "Profile name already exists", nil)
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
	if updateData.IPv6PrefixPool != "" {
		updates["ipv6_prefix_pool"] = updateData.IPv6PrefixPool
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

	if err := GetDB(c).Model(&profile).Updates(updates).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update profile", err.Error())
	}

	// Invalidate profile cache for dynamic users
	GetAppContext(c).ProfileCache().Invalidate(id)

	// Re-query latest data
	GetDB(c).First(&profile, id)

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
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid profile ID", nil)
	}

	// Check whether any users are currently using this profile
	var userCount int64
	GetDB(c).Model(&domain.RadiusUser{}).Where("profile_id = ?", id).Count(&userCount)
	if userCount > 0 {
		return fail(c, http.StatusConflict, "IN_USE", "Profile is in use and cannot be deleted", map[string]interface{}{
			"user_count": userCount,
		})
	}

	if err := GetDB(c).Delete(&domain.RadiusProfile{}, id).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete profile", err.Error())
	}

	// Invalidate profile cache
	GetAppContext(c).ProfileCache().Invalidate(id)

	return ok(c, map[string]interface{}{
		"message": "Deletion successful",
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
