package adminapi

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

// ListOnlineSessions List online sessions
// @Summary List online sessions
// @Tags OnlineSession
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort direction"
// @Param username query string false "Username"
// @Param nas_addr query string false "NAS addresses"
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
	if sortField == "" {
		sortField = "acct_start_time"
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	var total int64
	var sessions []domain.RadiusOnline

	query := db.Model(&domain.RadiusOnline{})

	// Filter by username
	if username := c.QueryParam("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	// Filter by NAS address
	if nasAddr := c.QueryParam("nas_addr"); nasAddr != "" {
		query = query.Where("nas_addr = ?", nasAddr)
	}

	// Filter by IP address
	if framedIp := c.QueryParam("framed_ipaddr"); framedIp != "" {
		query = query.Where("framed_ipaddr = ?", framedIp)
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
	nasErr := GetDB(c).Where("ip_addr = ?", session.NasAddr).First(&nas).Error

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
			rfc2866.AcctSessionID_SetString(pkt, session.AcctSessionId)
			rfc2865.UserName_SetString(pkt, session.Username)

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
					zap.Uint8("response_code", uint8(response.Code)),
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
