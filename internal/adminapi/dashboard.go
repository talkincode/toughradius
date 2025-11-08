package adminapi

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// DashboardStats Dashboard 统计数据结构
type DashboardStats struct {
	TotalUsers     int64   `json:"total_users"`      // 用户总数
	OnlineUsers    int64   `json:"online_users"`     // 在线用户数
	TodayAuthCount int64   `json:"today_auth_count"` // 今日认证次数
	TodayAcctCount int64   `json:"today_acct_count"` // 今日计费记录数
	TotalProfiles  int64   `json:"total_profiles"`   // 策略总数
	DisabledUsers  int64   `json:"disabled_users"`   // 禁用用户数
	ExpiredUsers   int64   `json:"expired_users"`    // 过期用户数
	TodayInputGB   float64 `json:"today_input_gb"`   // 今日上行流量(GB)
	TodayOutputGB  float64 `json:"today_output_gb"`  // 今日下行流量(GB)
}

// GetDashboardStats 获取 Dashboard 统计数据
// @Summary 获取 Dashboard 统计数据
// @Tags Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} DashboardStats
// @Router /api/v1/dashboard/stats [get]
func GetDashboardStats(c echo.Context) error {
	db := app.GDB()

	stats := &DashboardStats{}

	// 1. 用户总数
	db.Model(&domain.RadiusUser{}).Count(&stats.TotalUsers)

	// 2. 在线用户数
	db.Model(&domain.RadiusOnline{}).Count(&stats.OnlineUsers)

	// 3. 策略总数
	db.Model(&domain.RadiusProfile{}).Count(&stats.TotalProfiles)

	// 4. 禁用用户数
	db.Model(&domain.RadiusUser{}).Where("status = ?", "disabled").Count(&stats.DisabledUsers)

	// 5. 过期用户数
	db.Model(&domain.RadiusUser{}).Where("expire_time < ?", time.Now()).Count(&stats.ExpiredUsers)

	// 6. 今日认证次数（通过今日新增的在线会话估算）
	todayStart := time.Now().Truncate(24 * time.Hour)
	db.Model(&domain.RadiusOnline{}).Where("acct_start_time >= ?", todayStart).Count(&stats.TodayAuthCount)

	// 7. 今日计费记录数
	db.Model(&domain.RadiusAccounting{}).Where("acct_start_time >= ?", todayStart).Count(&stats.TodayAcctCount)

	// 8. 今日流量统计（字节转GB）
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

// registerDashboardRoutes 注册 Dashboard 路由
func registerDashboardRoutes() {
	webserver.ApiGET("/dashboard/stats", GetDashboardStats)
}
