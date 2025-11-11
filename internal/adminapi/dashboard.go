package adminapi

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// DashboardStats represents the dashboard statistics structure
type DashboardStats struct {
	TotalUsers     int64   `json:"total_users"`      // Total number of users
	OnlineUsers    int64   `json:"online_users"`     // Currently online users
	TodayAuthCount int64   `json:"today_auth_count"` // Authentication count for today
	TodayAcctCount int64   `json:"today_acct_count"` // Accounting record count for today
	TotalProfiles  int64   `json:"total_profiles"`   // Total number of profiles
	DisabledUsers  int64   `json:"disabled_users"`   // Disabled users
	ExpiredUsers   int64   `json:"expired_users"`    // Expired users
	TodayInputGB   float64 `json:"today_input_gb"`   // Today's upstream traffic (GB)
	TodayOutputGB  float64 `json:"today_output_gb"`  // Today's downstream traffic (GB)
}

// GetDashboardStats retrieves dashboard statistics
// @Summary get dashboard statistics
// @Tags Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} DashboardStats
// @Router /api/v1/dashboard/stats [get]
func GetDashboardStats(c echo.Context) error {
	db := app.GDB()

	stats := &DashboardStats{}

	// 1. Total users
	db.Model(&domain.RadiusUser{}).Count(&stats.TotalUsers)

	// 2. Online users
	db.Model(&domain.RadiusOnline{}).Count(&stats.OnlineUsers)

	// 3. Total profiles
	db.Model(&domain.RadiusProfile{}).Count(&stats.TotalProfiles)

	// 4. Disabled users
	db.Model(&domain.RadiusUser{}).Where("status = ?", "disabled").Count(&stats.DisabledUsers)

	// 5. Expired users
	db.Model(&domain.RadiusUser{}).Where("expire_time < ?", time.Now()).Count(&stats.ExpiredUsers)

	// 6. Today's authentication count (estimated from today's new online sessions)
	todayStart := time.Now().Truncate(24 * time.Hour)
	db.Model(&domain.RadiusOnline{}).Where("acct_start_time >= ?", todayStart).Count(&stats.TodayAuthCount)

	// 7. Today's accounting record count
	db.Model(&domain.RadiusAccounting{}).Where("acct_start_time >= ?", todayStart).Count(&stats.TodayAcctCount)

	// 8. Today's traffic statistics (bytes to GB)
	var flowStats struct {
		TotalInput  int64
		TotalOutput int64
	}
	db.Model(&domain.RadiusAccounting{}).
		Select("COALESCE(SUM(acct_input_total), 0) as total_input, COALESCE(SUM(acct_output_total), 0) as total_output").
		Where("acct_start_time >= ?", todayStart).
		Scan(&flowStats)

	stats.TodayInputGB = float64(flowStats.TotalInput) / (1024 * 1024 * 1024)
	stats.TodayOutputGB = float64(flowStats.TotalOutput) / (1024 * 1024 * 1024)

	return ok(c, stats)
}

// registerDashboardRoutes registers the dashboard routes
func registerDashboardRoutes() {
	webserver.ApiGET("/dashboard/stats", GetDashboardStats)
}
