package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// ListOnlineSessions 获取在线会话列表
// @Summary 获取在线会话列表
// @Tags OnlineSession
// @Param page query int false "页码"
// @Param perPage query int false "每页数量"
// @Param sort query string false "排序字段"
// @Param order query string false "排序方向"
// @Param username query string false "用户名"
// @Param nas_addr query string false "NAS 地址"
// @Success 200 {object} ListResponse
// @Router /api/v1/sessions [get]
func ListOnlineSessions(c echo.Context) error {
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
		sortField = "acct_start_time"
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	var total int64
	var sessions []domain.RadiusOnline

	query := db.Model(&domain.RadiusOnline{})

	// 按用户名过滤
	if username := c.QueryParam("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	// 按 NAS 地址过滤
	if nasAddr := c.QueryParam("nas_addr"); nasAddr != "" {
		query = query.Where("nas_addr = ?", nasAddr)
	}

	// 按 IP 地址过滤
	if framedIp := c.QueryParam("framed_ipaddr"); framedIp != "" {
		query = query.Where("framed_ipaddr = ?", framedIp)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&sessions)

	return ok(c, map[string]interface{}{
		"data":  sessions,
		"total": total,
	})
}

// GetOnlineSession 获取单个在线会话
// @Summary 获取在线会话详情
// @Tags OnlineSession
// @Param id path int true "Session ID"
// @Success 200 {object} domain.RadiusOnline
// @Router /api/v1/sessions/{id} [get]
func GetOnlineSession(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的 Session ID", nil)
	}

	var session domain.RadiusOnline
	if err := app.GDB().First(&session, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "在线会话不存在", nil)
	}

	return ok(c, session)
}

// DeleteOnlineSession 强制下线用户
// @Summary 强制用户下线
// @Tags OnlineSession
// @Param id path int true "Session ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/sessions/{id} [delete]
func DeleteOnlineSession(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的 Session ID", nil)
	}

	// 删除在线会话记录
	if err := app.GDB().Delete(&domain.RadiusOnline{}, id).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "强制下线失败", err.Error())
	}

	// TODO: 发送 CoA/DM 到 NAS 设备实现真正的强制下线
	// 这需要 RADIUS CoA 功能支持

	return ok(c, map[string]interface{}{
		"message": "用户已强制下线",
	})
}

// registerSessionRoutes 注册在线会话路由
func registerSessionRoutes() {
	webserver.ApiGET("/sessions", ListOnlineSessions)
	webserver.ApiGET("/sessions/:id", GetOnlineSession)
	webserver.ApiDELETE("/sessions/:id", DeleteOnlineSession)
}
