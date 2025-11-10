package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// ListAccounting 获取 Accounting 日志列表
// @Summary 获取 Accounting 日志列表
// @Tags Accounting
// @Param page query int false "页码"
// @Param perPage query int false "每页数量"
// @Param sort query string false "排序字段"
// @Param order query string false "排序方向"
// @Param username query string false "用户名"
// @Param nas_addr query string false "NAS 地址"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {object} ListResponse
// @Router /api/v1/accounting [get]
func ListAccounting(c echo.Context) error {
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
	var records []domain.RadiusAccounting

	query := db.Model(&domain.RadiusAccounting{})

	// 按用户名过滤
	if username := c.QueryParam("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	// 按 NAS 地址过滤
	if nasAddr := c.QueryParam("nas_addr"); nasAddr != "" {
		query = query.Where("nas_addr = ?", nasAddr)
	}

	// 按会话ID过滤
	if sessionId := c.QueryParam("acct_session_id"); sessionId != "" {
		query = query.Where("acct_session_id = ?", sessionId)
	}

	// 按时间范围过滤
	if startTime := c.QueryParam("start_time"); startTime != "" {
		query = query.Where("acct_start_time >= ?", startTime)
	}
	if endTime := c.QueryParam("end_time"); endTime != "" {
		query = query.Where("acct_stop_time <= ?", endTime)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&records)

	return paged(c, records, total, page, perPage)
}

// GetAccounting 获取单条 Accounting 记录
// @Summary 获取 Accounting 记录详情
// @Tags Accounting
// @Param id path int true "Accounting ID"
// @Success 200 {object} domain.RadiusAccounting
// @Router /api/v1/accounting/{id} [get]
func GetAccounting(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的 Accounting ID", nil)
	}

	var record domain.RadiusAccounting
	if err := app.GDB().First(&record, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Accounting 记录不存在", nil)
	}

	return ok(c, record)
}

// registerAccountingRoutes 注册 Accounting 路由
func registerAccountingRoutes() {
	webserver.ApiGET("/accounting", ListAccounting)
	webserver.ApiGET("/accounting/:id", GetAccounting)
}
