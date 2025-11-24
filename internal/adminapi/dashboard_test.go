package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestGetDashboardStats(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	now := time.Now()
	today := startOfDay(now)

	profiles := []*domain.RadiusProfile{
		{Name: "default", Status: "enabled"},
		{Name: "premium", Status: "enabled"},
	}
	for _, profile := range profiles {
		err := db.Create(profile).Error
		require.NoError(t, err)
	}

	users := []*domain.RadiusUser{
		{
			Username:   "alice",
			ProfileId:  profiles[0].ID,
			Status:     "enabled",
			ExpireTime: now.Add(24 * time.Hour),
		},
		{
			Username:   "bob",
			ProfileId:  profiles[1].ID,
			Status:     "disabled",
			ExpireTime: now.Add(48 * time.Hour),
		},
		{
			Username:   "carol",
			ProfileId:  profiles[0].ID,
			Status:     "enabled",
			ExpireTime: now.Add(-24 * time.Hour),
		},
	}
	for _, user := range users {
		err := db.Create(user).Error
		require.NoError(t, err)
	}

	onlineSessions := []*domain.RadiusOnline{
		{
			Username:      "alice",
			NasPortType:   5,
			ServiceType:   2,
			AcctSessionId: "session-alice",
			AcctStartTime: now.Add(-30 * time.Minute),
		},
		{
			Username:      "carol",
			NasPortType:   5,
			ServiceType:   2,
			AcctSessionId: "session-carol",
			AcctStartTime: now.Add(-25 * time.Minute),
		},
		{
			Username:      "bob",
			NasPortType:   19,
			ServiceType:   2,
			AcctSessionId: "session-bob",
			AcctStartTime: now.Add(-15 * time.Minute),
		},
		{
			Username:      "ghost",
			NasPortType:   15,
			ServiceType:   2,
			AcctSessionId: "session-ghost",
			AcctStartTime: now.Add(-10 * time.Minute),
		},
	}
	for _, session := range onlineSessions {
		err := db.Create(session).Error
		require.NoError(t, err)
	}

	accountingRecords := []*domain.RadiusAccounting{
		{
			Username:        "alice",
			AcctSessionId:   "acct-1",
			AcctStartTime:   today.Add(2 * time.Hour),
			AcctInputTotal:  int64(1 * 1024 * 1024 * 1024),
			AcctOutputTotal: int64(2 * 1024 * 1024 * 1024),
		},
		{
			Username:        "alice",
			AcctSessionId:   "acct-2",
			AcctStartTime:   today.Add(-26 * time.Hour),
			AcctInputTotal:  int64(500 * 1024 * 1024),
			AcctOutputTotal: int64(256 * 1024 * 1024),
		},
		{
			Username:        "carol",
			AcctSessionId:   "acct-3",
			AcctStartTime:   today.Add(-5 * 24 * time.Hour),
			AcctInputTotal:  int64(200 * 1024 * 1024),
			AcctOutputTotal: int64(300 * 1024 * 1024),
		},
	}
	for _, record := range accountingRecords {
		err := db.Create(record).Error
		require.NoError(t, err)
	}

	e := setupTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/stats", nil)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)

	err := GetDashboardStats(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	dataBytes, err := json.Marshal(response.Data)
	require.NoError(t, err)

	var stats DashboardStats
	err = json.Unmarshal(dataBytes, &stats)
	require.NoError(t, err)

	assert.Equal(t, int64(3), stats.TotalUsers)
	assert.Equal(t, int64(4), stats.OnlineUsers)
	assert.Equal(t, int64(2), stats.TotalProfiles)
	assert.Equal(t, int64(1), stats.DisabledUsers)
	assert.Equal(t, int64(1), stats.ExpiredUsers)
	require.Len(t, stats.AuthTrend, 7)

	var trendTotal int64
	for _, point := range stats.AuthTrend {
		trendTotal += point.Count
	}
	assert.Equal(t, int64(3), trendTotal)

	require.Len(t, stats.Traffic24h, 24)
	trafficMap := make(map[string]DashboardTrafficPoint, len(stats.Traffic24h))
	for _, point := range stats.Traffic24h {
		trafficMap[point.Hour] = point
	}
	targetHour := today.Add(2 * time.Hour).Format(hourKeyFormat)
	if point, ok := trafficMap[targetHour]; ok {
		assert.InDelta(t, 2.0, point.DownloadGB, 0.01)
	} else {
		t.Fatalf("expected traffic data for hour %s", targetHour)
	}
	assert.GreaterOrEqual(t, stats.TodayInputGB, 1.0)
	assert.GreaterOrEqual(t, stats.TodayOutputGB, 2.0)

	require.GreaterOrEqual(t, len(stats.ProfileDistribution), 2)
	profileMap := make(map[int64]DashboardProfileSlice)
	var unassignedCount int64
	for _, item := range stats.ProfileDistribution {
		profileMap[item.ProfileID] = item
		if item.ProfileID == 0 {
			unassignedCount = item.Value
		}
	}
	defaultProfile := profileMap[profiles[0].ID]
	require.Equal(t, profiles[0].Name, defaultProfile.ProfileName)
	assert.Equal(t, int64(2), defaultProfile.Value)
	premiumProfile := profileMap[profiles[1].ID]
	require.Equal(t, profiles[1].Name, premiumProfile.ProfileName)
	assert.Equal(t, int64(1), premiumProfile.Value)
	assert.Equal(t, int64(1), unassignedCount)
}
