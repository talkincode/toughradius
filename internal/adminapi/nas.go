package adminapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// nasPayload represents the NAS device request structure
type nasPayload struct {
	NodeId     int64  `json:"node_id,string" validate:"gte=0"`
	Name       string `json:"name" validate:"required,min=1,max=100"`
	Identifier string `json:"identifier" validate:"omitempty,max=100"`
	Hostname   string `json:"hostname" validate:"omitempty,max=100"`
	Ipaddr     string `json:"ipaddr" validate:"required,ip"`
	Secret     string `json:"secret" validate:"required,min=6,max=100"`
	CoaPort    *int   `json:"coa_port" validate:"omitempty,port"`
	Model      string `json:"model" validate:"omitempty,max=50"`
	VendorCode string `json:"vendor_code" validate:"omitempty,max=20"`
	Status     string `json:"status" validate:"omitempty,oneof=enabled disabled"`
	Tags       string `json:"tags" validate:"omitempty,max=200"`
	Remark     string `json:"remark" validate:"omitempty,max=500"`
}

// nasUpdatePayload relaxes validation rules for partial updates
type nasUpdatePayload struct {
	NodeId     int64  `json:"node_id,string" validate:"omitempty,gte=0"`
	Name       string `json:"name" validate:"omitempty,min=1,max=100"`
	Identifier string `json:"identifier" validate:"omitempty,max=100"`
	Hostname   string `json:"hostname" validate:"omitempty,max=100"`
	Ipaddr     string `json:"ipaddr" validate:"omitempty,ip"`
	Secret     string `json:"secret" validate:"omitempty,min=6,max=100"`
	CoaPort    *int   `json:"coa_port" validate:"omitempty,port"`
	Model      string `json:"model" validate:"omitempty,max=50"`
	VendorCode string `json:"vendor_code" validate:"omitempty,max=20"`
	Status     string `json:"status" validate:"omitempty,oneof=enabled disabled"`
	Tags       string `json:"tags" validate:"omitempty,max=200"`
	Remark     string `json:"remark" validate:"omitempty,max=500"`
}

// ListNAS retrieves the NAS device list
// @Summary get the NAS device list
// @Tags NAS
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort direction"
// @Param name query string false "Device name"
// @Param status query string false "Device status"
// @Success 200 {object} ListResponse
// @Router /api/v1/network/nas [get]
func ListNAS(c echo.Context) error {
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
	var devices []domain.NetNas

	query := db.Model(&domain.NetNas{})

	// Filter by name (case-insensitive)
	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		if strings.EqualFold(db.Name(), "postgres") { //nolint:staticcheck
			query = query.Where("name ILIKE ?", "%"+name+"%")
		} else {
			query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
		}
	}

	// Filter by status
	if status := strings.TrimSpace(c.QueryParam("status")); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by IP address (exact match or prefix match)
	if ipaddr := strings.TrimSpace(c.QueryParam("ipaddr")); ipaddr != "" {
		query = query.Where("ipaddr LIKE ?", ipaddr+"%")
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&devices)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  devices,
		"total": total,
	})
}

// GetNAS fetches a single NAS device
// @Summary get NAS device detail
// @Tags NAS
// @Param id path int true "NAS ID"
// @Success 200 {object} domain.NetNas
// @Router /api/v1/network/nas/{id} [get]
func GetNAS(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid NAS ID", nil)
	}

	var device domain.NetNas
	if err := GetDB(c).First(&device, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "NAS device not found", nil)
	}

	return ok(c, device)
}

// CreateNAS creates a NAS device
// @Summary create a NAS device
// @Tags NAS
// @Param nas body nasPayload true "NAS device information"
// @Success 201 {object} domain.NetNas
// @Router /api/v1/network/nas [post]
func CreateNAS(c echo.Context) error {
	var payload nasPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request parameters", err.Error())
	}

	// Validate the request payload
	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	// Check whether the IP address already exists
	var count int64
	GetDB(c).Model(&domain.NetNas{}).Where("ipaddr = ?", payload.Ipaddr).Count(&count)
	if count > 0 {
		return fail(c, http.StatusConflict, "IPADDR_EXISTS", "IP address already exists", nil)
	}

	// Set default values
	if payload.Status == "" {
		payload.Status = "enabled"
	}
	coaPort := 3799
	if payload.CoaPort != nil {
		coaPort = *payload.CoaPort
	}

	device := domain.NetNas{
		NodeId:     payload.NodeId,
		Name:       payload.Name,
		Identifier: payload.Identifier,
		Hostname:   payload.Hostname,
		Ipaddr:     payload.Ipaddr,
		Secret:     payload.Secret,
		CoaPort:    coaPort,
		Model:      payload.Model,
		VendorCode: payload.VendorCode,
		Status:     payload.Status,
		Tags:       payload.Tags,
		Remark:     payload.Remark,
	}

	if err := GetDB(c).Create(&device).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create NAS device", err.Error())
	}

	return ok(c, device)
}

// UpdateNAS updates a NAS device
// @Summary update a NAS device
// @Tags NAS
// @Param id path int true "NAS ID"
// @Param nas body nasPayload true "NAS device information"
// @Success 200 {object} domain.NetNas
// @Router /api/v1/network/nas/{id} [put]
func UpdateNAS(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid NAS ID", nil)
	}

	var device domain.NetNas
	if err := GetDB(c).First(&device, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "NAS device not found", nil)
	}

	var payload nasUpdatePayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request parameters", err.Error())
	}

	// Validate the request payload
	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	// Validate IP uniqueness (e.g., if the IP was modified)
	if payload.Ipaddr != "" && payload.Ipaddr != device.Ipaddr {
		var count int64
		GetDB(c).Model(&domain.NetNas{}).Where("ipaddr = ? AND id != ?", payload.Ipaddr, id).Count(&count)
		if count > 0 {
			return fail(c, http.StatusConflict, "IPADDR_EXISTS", "IP address already exists", nil)
		}
		device.Ipaddr = payload.Ipaddr
	}

	// Update fields
	if payload.Name != "" {
		device.Name = payload.Name
	}
	if payload.Identifier != "" {
		device.Identifier = payload.Identifier
	}
	if payload.Hostname != "" {
		device.Hostname = payload.Hostname
	}
	if payload.Secret != "" {
		device.Secret = payload.Secret
	}
	if payload.CoaPort != nil {
		device.CoaPort = *payload.CoaPort
	}
	if payload.Model != "" {
		device.Model = payload.Model
	}
	if payload.VendorCode != "" {
		device.VendorCode = payload.VendorCode
	}
	if payload.Status != "" {
		device.Status = payload.Status
	}
	if payload.Tags != "" {
		device.Tags = payload.Tags
	}
	if payload.Remark != "" {
		device.Remark = payload.Remark
	}
	if payload.NodeId > 0 {
		device.NodeId = payload.NodeId
	}

	if err := GetDB(c).Save(&device).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update NAS device", err.Error())
	}

	return ok(c, device)
}

// DeleteNAS deletes a NAS device
// @Summary delete a NAS device
// @Tags NAS
// @Param id path int true "NAS ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/network/nas/{id} [delete]
func DeleteNAS(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid NAS ID", nil)
	}

	// Check whether there are active online sessions
	var onlineCount int64
	GetDB(c).Model(&domain.RadiusOnline{}).Joins("JOIN net_vpe ON radius_online.nas_addr = net_vpe.ipaddr").Where("net_vpe.id = ?", id).Count(&onlineCount)
	if onlineCount > 0 {
		return fail(c, http.StatusConflict, "HAS_ONLINE_SESSIONS", "Device has active sessions and cannot be deleted", map[string]interface{}{
			"online_count": onlineCount,
		})
	}

	if err := GetDB(c).Delete(&domain.NetNas{}, id).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete NAS device", err.Error())
	}

	return ok(c, map[string]interface{}{
		"message": "Deletion successful",
	})
}

// registerNASRoutes registers NAS routes
func registerNASRoutes() {
	webserver.ApiGET("/network/nas", ListNAS)
	webserver.ApiGET("/network/nas/:id", GetNAS)
	webserver.ApiPOST("/network/nas", CreateNAS)
	webserver.ApiPUT("/network/nas/:id", UpdateNAS)
	webserver.ApiDELETE("/network/nas/:id", DeleteNAS)
}
