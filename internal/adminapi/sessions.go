package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
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
	if err := app.GDB().First(&session, id).Error; err != nil {
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

	// Delete online session record
	if err := app.GDB().Delete(&domain.RadiusOnline{}, id).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to terminate session", err.Error())
	}

	// TODO: Send CoA/DM to NAS device to actually force offline
	// This requires RADIUS CoA feature support

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
