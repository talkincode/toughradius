package adminapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// ListAccounting retrieves the accounting logs table
// @Summary get accounting logs table
// @Tags Accounting
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort direction"
// @Param username query string false "Username"
// @Param nas_addr query string false "NAS addresses"
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Success 200 {object} ListResponse
// @Router /api/v1/accounting [get]
func ListAccounting(c echo.Context) error {
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
		sortField = "acct_start_time"
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	var total int64
	var records []domain.RadiusAccounting

	query := db.Model(&domain.RadiusAccounting{})

	// Filter by username
	if username := c.QueryParam("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	// Filter by NAS address
	if nasAddr := c.QueryParam("nas_addr"); nasAddr != "" {
		query = query.Where("nas_addr = ?", nasAddr)
	}

	// Filter by session ID
	if sessionId := c.QueryParam("acct_session_id"); sessionId != "" {
		query = query.Where("acct_session_id = ?", sessionId)
	}

	// Filter by time range
	if startTime := c.QueryParam("start_time"); startTime != "" {
		parsedTime, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			query = query.Where("acct_start_time >= ?", parsedTime)
		}
	}
	if endTime := c.QueryParam("end_time"); endTime != "" {
		parsedTime, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			query = query.Where("acct_stop_time <= ?", parsedTime)
		}
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&records)

	return paged(c, records, total, page, perPage)
}

// GetAccounting fetches a single accounting record
// @Summary get accounting record detail
// @Tags Accounting
// @Param id path int true "Accounting ID"
// @Success 200 {object} domain.RadiusAccounting
// @Router /api/v1/accounting/{id} [get]
func GetAccounting(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid Accounting ID", nil)
	}

	var record domain.RadiusAccounting
	if err := GetDB(c).First(&record, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Accounting record not found", nil)
	}

	return ok(c, record)
}

// registerAccountingRoutes registers accounting routes
func registerAccountingRoutes() {
	webserver.ApiGET("/accounting", ListAccounting)
	webserver.ApiGET("/accounting/:id", GetAccounting)
}
