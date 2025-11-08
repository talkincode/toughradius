package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// ListNAS 获取 NAS 设备列表
// @Summary 获取 NAS 设备列表
// @Tags NAS
// @Param page query int false "页码"
// @Param perPage query int false "每页数量"
// @Param sort query string false "排序字段"
// @Param order query string false "排序方向"
// @Param name query string false "设备名称"
// @Param status query string false "设备状态"
// @Success 200 {object} ListResponse
// @Router /api/v1/nas [get]
func ListNAS(c echo.Context) error {
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
	var devices []domain.NetVpe

	query := db.Model(&domain.NetVpe{})

	// 按名称过滤
	if name := c.QueryParam("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// 按状态过滤
	if status := c.QueryParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 按 IP 地址过滤
	if ipaddr := c.QueryParam("ipaddr"); ipaddr != "" {
		query = query.Where("ipaddr = ?", ipaddr)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&devices)

	return ok(c, map[string]interface{}{
		"data":  devices,
		"total": total,
	})
}

// GetNAS 获取单个 NAS 设备
// @Summary 获取 NAS 设备详情
// @Tags NAS
// @Param id path int true "NAS ID"
// @Success 200 {object} domain.NetVpe
// @Router /api/v1/nas/{id} [get]
func GetNAS(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的 NAS ID", nil)
	}

	var device domain.NetVpe
	if err := app.GDB().First(&device, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "NAS 设备不存在", nil)
	}

	return ok(c, device)
}

// CreateNAS 创建 NAS 设备
// @Summary 创建 NAS 设备
// @Tags NAS
// @Param nas body domain.NetVpe true "NAS 设备信息"
// @Success 201 {object} domain.NetVpe
// @Router /api/v1/nas [post]
func CreateNAS(c echo.Context) error {
	var device domain.NetVpe
	if err := c.Bind(&device); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析请求参数", err.Error())
	}

	// 验证必填字段
	if device.Name == "" {
		return fail(c, http.StatusBadRequest, "MISSING_NAME", "设备名称不能为空", nil)
	}
	if device.Ipaddr == "" {
		return fail(c, http.StatusBadRequest, "MISSING_IPADDR", "设备 IP 地址不能为空", nil)
	}
	if device.Secret == "" {
		return fail(c, http.StatusBadRequest, "MISSING_SECRET", "RADIUS 秘钥不能为空", nil)
	}

	// 检查 IP 地址是否已存在
	var count int64
	app.GDB().Model(&domain.NetVpe{}).Where("ipaddr = ?", device.Ipaddr).Count(&count)
	if count > 0 {
		return fail(c, http.StatusConflict, "IPADDR_EXISTS", "IP 地址已存在", nil)
	}

	// 设置默认值
	if device.Status == "" {
		device.Status = "enabled"
	}
	if device.CoaPort == 0 {
		device.CoaPort = 3799
	}

	if err := app.GDB().Create(&device).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "创建 NAS 设备失败", err.Error())
	}

	return ok(c, device)
}

// UpdateNAS 更新 NAS 设备
// @Summary 更新 NAS 设备
// @Tags NAS
// @Param id path int true "NAS ID"
// @Param nas body domain.NetVpe true "NAS 设备信息"
// @Success 200 {object} domain.NetVpe
// @Router /api/v1/nas/{id} [put]
func UpdateNAS(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的 NAS ID", nil)
	}

	var device domain.NetVpe
	if err := app.GDB().First(&device, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "NAS 设备不存在", nil)
	}

	var updateData domain.NetVpe
	if err := c.Bind(&updateData); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析请求参数", err.Error())
	}

	// 验证 IP 唯一性
	if updateData.Ipaddr != "" && updateData.Ipaddr != device.Ipaddr {
		var count int64
		app.GDB().Model(&domain.NetVpe{}).Where("ipaddr = ? AND id != ?", updateData.Ipaddr, id).Count(&count)
		if count > 0 {
			return fail(c, http.StatusConflict, "IPADDR_EXISTS", "IP 地址已存在", nil)
		}
	}

	// 更新字段
	updates := map[string]interface{}{}
	if updateData.Name != "" {
		updates["name"] = updateData.Name
	}
	if updateData.Identifier != "" {
		updates["identifier"] = updateData.Identifier
	}
	if updateData.Hostname != "" {
		updates["hostname"] = updateData.Hostname
	}
	if updateData.Ipaddr != "" {
		updates["ipaddr"] = updateData.Ipaddr
	}
	if updateData.Secret != "" {
		updates["secret"] = updateData.Secret
	}
	if updateData.CoaPort > 0 {
		updates["coa_port"] = updateData.CoaPort
	}
	if updateData.Model != "" {
		updates["model"] = updateData.Model
	}
	if updateData.VendorCode != "" {
		updates["vendor_code"] = updateData.VendorCode
	}
	if updateData.Status != "" {
		updates["status"] = updateData.Status
	}
	if updateData.Tags != "" {
		updates["tags"] = updateData.Tags
	}
	if updateData.Remark != "" {
		updates["remark"] = updateData.Remark
	}
	if updateData.NodeId > 0 {
		updates["node_id"] = updateData.NodeId
	}

	if err := app.GDB().Model(&device).Updates(updates).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "更新 NAS 设备失败", err.Error())
	}

	// 重新查询最新数据
	app.GDB().First(&device, id)

	return ok(c, device)
}

// DeleteNAS 删除 NAS 设备
// @Summary 删除 NAS 设备
// @Tags NAS
// @Param id path int true "NAS ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/nas/{id} [delete]
func DeleteNAS(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的 NAS ID", nil)
	}

	// 检查是否有在线会话
	var onlineCount int64
	app.GDB().Model(&domain.RadiusOnline{}).Joins("JOIN net_vpe ON radius_online.nas_addr = net_vpe.ipaddr").Where("net_vpe.id = ?", id).Count(&onlineCount)
	if onlineCount > 0 {
		return fail(c, http.StatusConflict, "HAS_ONLINE_SESSIONS", "该设备有在线会话，无法删除", map[string]interface{}{
			"online_count": onlineCount,
		})
	}

	if err := app.GDB().Delete(&domain.NetVpe{}, id).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "删除 NAS 设备失败", err.Error())
	}

	return ok(c, map[string]interface{}{
		"message": "删除成功",
	})
}

// registerNASRoutes 注册 NAS 路由
func registerNASRoutes() {
	webserver.ApiGET("/nas", ListNAS)
	webserver.ApiGET("/nas/:id", GetNAS)
	webserver.ApiPOST("/nas", CreateNAS)
	webserver.ApiPUT("/nas/:id", UpdateNAS)
	webserver.ApiDELETE("/nas/:id", DeleteNAS)
}
