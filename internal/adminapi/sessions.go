package adminapi

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

// allowedSessionSortFields defines the whitelist of sortable fields for online sessions
// This prevents SQL injection through the sort parameter
var allowedSessionSortFields = map[string]bool{
	"id":              true,
	"username":        true,
	"nas_addr":        true,
	"framed_ipaddr":   true,
	"acct_start_time": true,
	"acct_session_id": true,
	"mac_addr":        true,
}

// escapeSessionLikePattern escapes special characters in LIKE pattern to prevent wildcard injection
// This escapes %, _, and \ which are SQL LIKE wildcards
func escapeSessionLikePattern(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\") // escape backslash first
	s = strings.ReplaceAll(s, "%", "\\%")   // escape percent
	s = strings.ReplaceAll(s, "_", "\\_")   // escape underscore
	return s
}

// parseSessionTime attempts to parse time in multiple formats
func parseSessionTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04",    // datetime-local format from HTML5
		"2006-01-02 15:04:05", // standard datetime format
		"2006-01-02",          // date only
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, &time.ParseError{Layout: "multiple", Value: s, Message: "unable to parse time"}
}

// ListOnlineSessions List online sessions
// @Summary List online sessions
// @Tags OnlineSession
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort direction"
// @Param username query string false "Username"
// @Param nas_addr query string false "NAS addresses"
// @Param framed_ipaddr query string false "User IP address"
// @Param framed_ipv6addr query string false "User IPv6 address"
// @Param mac_addr query string false "MAC address"
// @Param acct_session_id query string false "Session ID"
// @Param acct_start_time_gte query string false "Start time from (RFC3339 or datetime-local)"
// @Param acct_start_time_lte query string false "Start time to (RFC3339 or datetime-local)"
// @Success 200 {object} ListResponse
// @Router /api/v1/sessions [get]
func ListOnlineSessions(c echo.Context) error {
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

	// Validate sort field against whitelist to prevent SQL injection
	if sortField == "" || !allowedSessionSortFields[sortField] {
		sortField = "acct_start_time"
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	var total int64
	var sessions []domain.RadiusOnline

	query := db.Model(&domain.RadiusOnline{})

	// Filter by username (LIKE with escaped pattern)
	if username := c.QueryParam("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+escapeSessionLikePattern(username)+"%")
	}

	// Filter by NAS address
	if nasAddr := c.QueryParam("nas_addr"); nasAddr != "" {
		query = query.Where("nas_addr = ?", nasAddr)
	}

	// Filter by IP address
	if framedIp := c.QueryParam("framed_ipaddr"); framedIp != "" {
		query = query.Where("framed_ipaddr = ?", framedIp)
	}

	// Filter by IPv6 address (LIKE with escaped pattern)
	if framedIpv6 := c.QueryParam("framed_ipv6addr"); framedIpv6 != "" {
		query = query.Where("framed_ipv6addr LIKE ?", "%"+escapeSessionLikePattern(framedIpv6)+"%")
	}

	// Filter by MAC address (LIKE with escaped pattern)
	if macAddr := c.QueryParam("mac_addr"); macAddr != "" {
		query = query.Where("mac_addr LIKE ?", "%"+escapeSessionLikePattern(macAddr)+"%")
	}

	// Filter by Session ID (LIKE with escaped pattern)
	if acctSessionId := c.QueryParam("acct_session_id"); acctSessionId != "" {
		query = query.Where("acct_session_id LIKE ?", "%"+escapeSessionLikePattern(acctSessionId)+"%")
	}

	// Filter by start time range
	if startTimeGte := c.QueryParam("acct_start_time_gte"); startTimeGte != "" {
		if t, err := parseSessionTime(startTimeGte); err == nil {
			query = query.Where("acct_start_time >= ?", t)
		}
	}
	if startTimeLte := c.QueryParam("acct_start_time_lte"); startTimeLte != "" {
		if t, err := parseSessionTime(startTimeLte); err == nil {
			query = query.Where("acct_start_time <= ?", t)
		}
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&sessions)

	return paged(c, sessions, total, page, perPage)
}

// GetOnlineSession Get single online session
// @Summary Get online session details
// @Tags OnlineSession
// @Param id path int true "Session ID"
// @Success 200 {object} domain.RadiusOnline
// @Router /api/v1/sessions/{id} [get]
func GetOnlineSession(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid Session ID", nil)
	}

	var session domain.RadiusOnline
	if err := GetDB(c).First(&session, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Session not found", nil)
	}

	return ok(c, session)
}

// DeleteOnlineSession Force user offline
// @Summary Force user offline
// @Tags OnlineSession
// @Param id path int true "Session ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/sessions/{id} [delete]
func DeleteOnlineSession(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid Session ID", nil)
	}

	// Fetch session before deletion for CoA
	var session domain.RadiusOnline
	if err := GetDB(c).First(&session, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Session not found", nil)
	}

	// Fetch NAS info for CoA
	var nas domain.NetNas
	nasErr := GetDB(c).Where("ipaddr = ?", session.NasAddr).First(&nas).Error

	// Delete online session record
	if err := GetDB(c).Delete(&domain.RadiusOnline{}, id).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to terminate session", err.Error())
	}

	// Send CoA Disconnect-Request to NAS asynchronously (non-blocking)
	if nasErr == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Build CoA Disconnect-Request packet
			pkt := radius.New(radius.CodeDisconnectRequest, []byte(nas.Secret))
			_ = rfc2866.AcctSessionID_SetString(pkt, session.AcctSessionId) //nolint:errcheck
			_ = rfc2865.UserName_SetString(pkt, session.Username)           //nolint:errcheck

			// Send to NAS CoA port (default 3799)
			coaAddr := net.JoinHostPort(nas.Ipaddr, "3799")
			client := &radius.Client{
				Retry: time.Second * 2,
			}

			response, err := client.Exchange(ctx, pkt, coaAddr)
			if err != nil {
				zap.L().Error("Failed to send CoA Disconnect-Request",
					zap.Error(err),
					zap.String("nas_addr", coaAddr),
					zap.String("username", session.Username),
					zap.String("acct_session_id", session.AcctSessionId),
					zap.String("namespace", "adminapi"))
				return
			}

			if response.Code == radius.CodeDisconnectACK {
				zap.L().Info("CoA Disconnect-Request ACK received",
					zap.String("nas_addr", coaAddr),
					zap.String("username", session.Username),
					zap.String("namespace", "adminapi"))
			} else {
				zap.L().Warn("CoA Disconnect-Request NAK received",
					zap.String("nas_addr", coaAddr),
					zap.String("username", session.Username),
					zap.Uint8("response_code", uint8(response.Code)), //nolint:gosec // G115: RADIUS code is always in uint8 range
					zap.String("namespace", "adminapi"))
			}
		}()
	} else {
		zap.L().Warn("NAS not found for CoA, session deleted from database only",
			zap.String("nas_addr", session.NasAddr),
			zap.String("username", session.Username),
			zap.String("namespace", "adminapi"))
	}

	return ok(c, map[string]interface{}{
		"message": "User has been forced offline",
	})
}

// registerSessionRoutes Register online session routes
func registerSessionRoutes() {
	webserver.ApiGET("/sessions", ListOnlineSessions)
	webserver.ApiGET("/sessions/:id", GetOnlineSession)
	webserver.ApiDELETE("/sessions/:id", DeleteOnlineSession)
}
