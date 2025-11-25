package adminapi

import (
	"net/http"
	"strconv"
	"strings"
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
// @Param nas_addr query string false "NAS address"
// @Param acct_session_id query string false "Session ID"
// @Param framed_ipaddr query string false "User IP address"
// @Param mac_addr query string false "MAC address"
// @Param framed_ipv6addr query string false "IPv6 address"
// @Param acct_start_time_gte query string false "Start time from (RFC3339 or datetime-local format)"
// @Param acct_start_time_lte query string false "Start time to (RFC3339 or datetime-local format)"
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

	// 白名单验证排序字段，防止 SQL 注入
	allowedSortFields := map[string]bool{
		"id": true, "username": true, "acct_session_id": true,
		"acct_start_time": true, "acct_stop_time": true, "acct_session_time": true,
		"nas_addr": true, "framed_ipaddr": true, "acct_input_total": true, "acct_output_total": true,
	}
	if !allowedSortFields[sortField] {
		sortField = "acct_start_time"
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	var total int64
	var records []domain.RadiusAccounting

	query := db.Model(&domain.RadiusAccounting{})

	// Filter by username (使用转义防止 LIKE 通配符注入)
	if username := c.QueryParam("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+escapeLikePattern(username)+"%")
	}

	// Filter by NAS address
	if nasAddr := c.QueryParam("nas_addr"); nasAddr != "" {
		query = query.Where("nas_addr LIKE ?", "%"+escapeLikePattern(nasAddr)+"%")
	}

	// Filter by session ID
	if sessionId := c.QueryParam("acct_session_id"); sessionId != "" {
		query = query.Where("acct_session_id LIKE ?", "%"+escapeLikePattern(sessionId)+"%")
	}

	// Filter by framed IP address
	if framedIp := c.QueryParam("framed_ipaddr"); framedIp != "" {
		query = query.Where("framed_ipaddr LIKE ?", "%"+escapeLikePattern(framedIp)+"%")
	}

	// Filter by MAC address
	if macAddr := c.QueryParam("mac_addr"); macAddr != "" {
		query = query.Where("mac_addr LIKE ?", "%"+escapeLikePattern(macAddr)+"%")
	}

	// Filter by IPv6 address
	if framedIpv6 := c.QueryParam("framed_ipv6addr"); framedIpv6 != "" {
		query = query.Where("framed_ipv6addr LIKE ?", "%"+escapeLikePattern(framedIpv6)+"%")
	}

	// Filter by start time range (支持 RFC3339 和 datetime-local 两种格式)
	if startTimeGte := c.QueryParam("acct_start_time_gte"); startTimeGte != "" {
		parsedTime, err := parseFlexibleTime(startTimeGte)
		if err == nil {
			query = query.Where("acct_start_time >= ?", parsedTime)
		}
	}
	if startTimeLte := c.QueryParam("acct_start_time_lte"); startTimeLte != "" {
		parsedTime, err := parseFlexibleTime(startTimeLte)
		if err == nil {
			query = query.Where("acct_start_time <= ?", parsedTime)
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

// parseFlexibleTime parses time string in RFC3339 or datetime-local format
func parseFlexibleTime(s string) (time.Time, error) {
	// Try RFC3339 first (e.g., "2025-11-01T21:16:00Z")
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// Try datetime-local format (e.g., "2025-11-01T21:16")
	if t, err := time.ParseInLocation("2006-01-02T15:04", s, time.Local); err == nil {
		return t, nil
	}
	// Try date only format (e.g., "2025-11-01")
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return t, nil
	}
	return time.Time{}, &time.ParseError{Layout: "multiple", Value: s, Message: "unable to parse time"}
}

// escapeLikePattern escapes special characters in LIKE pattern to prevent wildcard injection
// This escapes %, _, and \ which are SQL LIKE wildcards
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\") // escape backslash first
	s = strings.ReplaceAll(s, "%", "\\%")   // escape percent
	s = strings.ReplaceAll(s, "_", "\\_")   // escape underscore
	return s
}

// registerAccountingRoutes registers accounting routes
func registerAccountingRoutes() {
	webserver.ApiGET("/accounting", ListAccounting)
	webserver.ApiGET("/accounting/:id", GetAccounting)
}
