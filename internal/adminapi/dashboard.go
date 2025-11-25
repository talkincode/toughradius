package adminapi

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DashboardStats represents the dashboard statistics structure
type DashboardStats struct {
	TotalUsers          int64                     `json:"total_users"`          // Total number of users
	OnlineUsers         int64                     `json:"online_users"`         // Currently online users
	TodayAuthCount      int64                     `json:"today_auth_count"`     // Authentication count for today
	TodayAcctCount      int64                     `json:"today_acct_count"`     // Accounting record count for today
	TotalProfiles       int64                     `json:"total_profiles"`       // Total number of profiles
	DisabledUsers       int64                     `json:"disabled_users"`       // Disabled users
	ExpiredUsers        int64                     `json:"expired_users"`        // Expired users
	TodayInputGB        float64                   `json:"today_input_gb"`       // Today's upstream traffic (GB)
	TodayOutputGB       float64                   `json:"today_output_gb"`      // Today's downstream traffic (GB)
	AuthTrend           []DashboardAuthTrendPoint `json:"auth_trend"`           // Daily authentication trend (last 7 days)
	Traffic24h          []DashboardTrafficPoint   `json:"traffic_24h"`          // Hourly traffic statistics (last 24 hours)
	ProfileDistribution []DashboardProfileSlice   `json:"profile_distribution"` // Online users grouped by profile
}

// DashboardAuthTrendPoint represents authentication count per day
type DashboardAuthTrendPoint struct {
	Date  string `json:"date"`  // Date label formatted as YYYY-MM-DD
	Count int64  `json:"count"` // Authentication count for the day
}

// DashboardTrafficPoint represents hourly upload/download traffic
type DashboardTrafficPoint struct {
	Hour       string  `json:"hour"`        // Hour label formatted as YYYY-MM-DD HH:00
	UploadGB   float64 `json:"upload_gb"`   // Upload traffic in GB within the hour
	DownloadGB float64 `json:"download_gb"` // Download traffic in GB within the hour
}

// DashboardProfileSlice represents online user distribution grouped by profile
type DashboardProfileSlice struct {
	ProfileID   int64  `json:"profile_id"`
	ProfileName string `json:"profile_name"`
	Value       int64  `json:"value"`
}

const (
	dateKeyFormat = "2006-01-02"
	hourKeyFormat = "2006-01-02 15:00"
	bytesInGB     = float64(1024 * 1024 * 1024)
)

// GetDashboardStats retrieves dashboard statistics
// @Summary get dashboard statistics
// @Tags Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} DashboardStats
// @Router /api/v1/dashboard/stats [get]
func GetDashboardStats(c echo.Context) error {
	db := GetDB(c).WithContext(c.Request().Context())
	now := time.Now()
	todayStart := startOfDay(now)

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
	db.Model(&domain.RadiusUser{}).Where("expire_time < ?", now).Count(&stats.ExpiredUsers)

	// 6. Today's authentication count (estimated from today's new online sessions)
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

	stats.TodayInputGB = float64(flowStats.TotalInput) / bytesInGB
	stats.TodayOutputGB = float64(flowStats.TotalOutput) / bytesInGB
	stats.AuthTrend = fetchAuthTrend(db, now)
	stats.Traffic24h = fetchTrafficStats(db, now)
	stats.ProfileDistribution = fetchProfileDistribution(db)

	return ok(c, stats)
}

// registerDashboardRoutes registers the dashboard routes
func registerDashboardRoutes() {
	webserver.ApiGET("/dashboard/stats", GetDashboardStats)
}

func fetchAuthTrend(db *gorm.DB, now time.Time) []DashboardAuthTrendPoint {
	const days = 7
	result := make([]DashboardAuthTrendPoint, days)
	seriesEnd := startOfDay(now).Add(24 * time.Hour)
	seriesStart := seriesEnd.AddDate(0, 0, -days)
	bucketExpr := dateBucketExpression(db, "acct_start_time", "day")
	var rows []struct {
		Bucket string
		Count  int64
	}
	if err := db.Model(&domain.RadiusAccounting{}).
		Select(fmt.Sprintf("%s AS bucket, COUNT(*) AS count", bucketExpr)).
		Where("acct_start_time >= ? AND acct_start_time < ?", seriesStart, seriesEnd).
		Group("bucket").
		Order("bucket").
		Scan(&rows).Error; err != nil {
		logDashboardQueryError("fetch auth trend", err)
	}
	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Bucket] = row.Count
	}
	for i := 0; i < days; i++ {
		day := seriesStart.AddDate(0, 0, i)
		key := day.Format(dateKeyFormat)
		result[i] = DashboardAuthTrendPoint{
			Date:  key,
			Count: counts[key],
		}
	}
	return result
}

func fetchTrafficStats(db *gorm.DB, now time.Time) []DashboardTrafficPoint {
	const hours = 24
	result := make([]DashboardTrafficPoint, hours)
	hourEnd := startOfHour(now).Add(time.Hour)
	hourStart := hourEnd.Add(-hours * time.Hour)
	bucketExpr := dateBucketExpression(db, "acct_start_time", "hour")
	var rows []struct {
		Bucket   string
		Upload   float64
		Download float64
	}
	if err := db.Model(&domain.RadiusAccounting{}).
		Select(fmt.Sprintf("%s AS bucket, COALESCE(SUM(acct_input_total), 0) AS upload, COALESCE(SUM(acct_output_total), 0) AS download", bucketExpr)).
		Where("acct_start_time >= ? AND acct_start_time < ?", hourStart, hourEnd).
		Group("bucket").
		Order("bucket").
		Scan(&rows).Error; err != nil {
		logDashboardQueryError("fetch traffic stats", err)
	}
	lookup := make(map[string]struct {
		Upload   float64
		Download float64
	})
	for _, row := range rows {
		lookup[row.Bucket] = struct {
			Upload   float64
			Download float64
		}{Upload: row.Upload / bytesInGB, Download: row.Download / bytesInGB}
	}
	for i := 0; i < hours; i++ {
		hour := hourStart.Add(time.Duration(i) * time.Hour)
		key := hour.Format(hourKeyFormat)
		if val, ok := lookup[key]; ok {
			result[i] = DashboardTrafficPoint{Hour: key, UploadGB: val.Upload, DownloadGB: val.Download}
			continue
		}
		result[i] = DashboardTrafficPoint{Hour: key, UploadGB: 0, DownloadGB: 0}
	}
	return result
}

func fetchProfileDistribution(db *gorm.DB) []DashboardProfileSlice {
	var rows []struct {
		ProfileID   int64
		ProfileName string
		Count       int64
	}
	onlineTable := domain.RadiusOnline{}.TableName()
	userTable := domain.RadiusUser{}.TableName()
	profileTable := domain.RadiusProfile{}.TableName()
	if err := db.Table(fmt.Sprintf("%s AS o", onlineTable)).
		Select("COALESCE(u.profile_id, 0) AS profile_id, COALESCE(p.name, '') AS profile_name, COUNT(*) AS count").
		Joins(fmt.Sprintf("LEFT JOIN %s AS u ON u.username = o.username", userTable)).
		Joins(fmt.Sprintf("LEFT JOIN %s AS p ON p.id = u.profile_id", profileTable)).
		Group("profile_id, profile_name").
		Order("count DESC").
		Scan(&rows).Error; err != nil {
		logDashboardQueryError("fetch profile distribution", err)
	}
	result := make([]DashboardProfileSlice, 0, len(rows))
	for _, row := range rows {
		result = append(result, DashboardProfileSlice{
			ProfileID:   row.ProfileID,
			ProfileName: row.ProfileName,
			Value:       row.Count,
		})
	}
	return result
}

func dateBucketExpression(db *gorm.DB, field, granularity string) string {
	switch db.Name() { //nolint:staticcheck
	case "postgres":
		switch granularity {
		case "day":
			return fmt.Sprintf("DATE(%s)", field)
		case "hour":
			return fmt.Sprintf("TO_CHAR(date_trunc('hour', %s), 'YYYY-MM-DD HH24:00')", field)
		}
	default:
		switch granularity {
		case "day":
			return fmt.Sprintf("strftime('%%Y-%%m-%%d', %s, 'localtime')", field)
		case "hour":
			return fmt.Sprintf("strftime('%%Y-%%m-%%d %%H:00', %s, 'localtime')", field)
		}
	}
	return field
}

func startOfDay(t time.Time) time.Time {
	loc := t.Location()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func startOfHour(t time.Time) time.Time {
	loc := t.Location()
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc)
}

func logDashboardQueryError(action string, err error) {
	if err == nil {
		return
	}
	zap.L().Warn("dashboard query failed",
		zap.Error(err),
		zap.String("action", action),
		zap.String("namespace", "adminapi"))
}
