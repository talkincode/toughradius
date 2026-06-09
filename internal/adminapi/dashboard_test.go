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
			AcctStartTime:   now.Add(-1 * time.Minute), // 1 minute ago - within today and 24h window
			AcctInputTotal:  int64(1 * 1024 * 1024 * 1024),
			AcctOutputTotal: int64(2 * 1024 * 1024 * 1024),
		},
		{
			Username:        "alice",
			AcctSessionId:   "acct-2",
			AcctStartTime:   now.Add(-26 * time.Hour), // 26 hours ago - outside 24h window
			AcctInputTotal:  int64(500 * 1024 * 1024),
			AcctOutputTotal: int64(256 * 1024 * 1024),
		},
		{
			Username:        "carol",
			AcctSessionId:   "acct-3",
			AcctStartTime:   now.Add(-5 * 24 * time.Hour), // 5 days ago - outside both windows
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
	// Calculate total traffic from all 24h points
	// Note: Due to timezone differences between Go's time.Now() and SQLite's localtime,
	// we check total traffic sum instead of specific hour to ensure CI compatibility
	var totalDownloadGB float64
	for _, point := range stats.Traffic24h {
		totalDownloadGB += point.DownloadGB
	}
	// The first accounting record (acct-1) has 2GB download and was created 1 minute ago
	assert.InDelta(t, 2.0, totalDownloadGB, 0.01)
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

// TestGetDashboardIPv6Stats exercises the IPv6 dimension of the dashboard along
// both axes: live adoption across online sessions and static provisioning across
// the user base. It regression-protects the online counters (previously untested)
// and the user-base counters.
func TestGetDashboardIPv6Stats(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	now := time.Now()

	// Users: two with a static IPv6 address, two with a static delegated prefix
	// (one user overlaps with both), and one with no IPv6 provisioning at all.
	users := []*domain.RadiusUser{
		{Username: "v6addr1", Status: "enabled", ExpireTime: now.Add(24 * time.Hour), IpV6Addr: "2001:db8:1::1"},
		{Username: "v6both", Status: "enabled", ExpireTime: now.Add(24 * time.Hour), IpV6Addr: "2001:db8:2::1", DelegatedIpv6Prefix: "2001:db8:2:100::/56"},
		{Username: "v6pd", Status: "enabled", ExpireTime: now.Add(24 * time.Hour), DelegatedIpv6Prefix: "2001:db8:3::/48"},
		{Username: "plain", Status: "enabled", ExpireTime: now.Add(24 * time.Hour)},
	}
	for _, user := range users {
		require.NoError(t, db.Create(user).Error)
	}

	// Online sessions: one per IPv6 attribute, one carrying all three, one with none.
	sessions := []*domain.RadiusOnline{
		{Username: "v6addr1", AcctSessionId: "s-addr", AcctStartTime: now.Add(-5 * time.Minute), FramedIpv6Address: "2001:db8:1::1"},
		{Username: "v6both", AcctSessionId: "s-prefix", AcctStartTime: now.Add(-5 * time.Minute), FramedIpv6Prefix: "2001:db8:2::/64"},
		{Username: "v6pd", AcctSessionId: "s-pd", AcctStartTime: now.Add(-5 * time.Minute), DelegatedIpv6Prefix: "2001:db8:3::/48"},
		{Username: "v6both", AcctSessionId: "s-all", AcctStartTime: now.Add(-5 * time.Minute), FramedIpv6Address: "2001:db8:2::1", FramedIpv6Prefix: "2001:db8:2::/64", DelegatedIpv6Prefix: "2001:db8:2:100::/56"},
		{Username: "plain", AcctSessionId: "s-none", AcctStartTime: now.Add(-5 * time.Minute)},
	}
	for _, session := range sessions {
		require.NoError(t, db.Create(session).Error)
	}

	e := setupTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/stats", nil)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)

	require.NoError(t, GetDashboardStats(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var response Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	dataBytes, err := json.Marshal(response.Data)
	require.NoError(t, err)
	var stats DashboardStats
	require.NoError(t, json.Unmarshal(dataBytes, &stats))

	ipv6 := stats.IPv6Stats
	// Online adoption dimension.
	assert.Equal(t, int64(2), ipv6.OnlineWithIPv6Address, "framed address: s-addr + s-all")
	assert.Equal(t, int64(2), ipv6.OnlineWithFramedPrefix, "framed prefix: s-prefix + s-all")
	assert.Equal(t, int64(2), ipv6.OnlineWithDelegatedPrefix, "delegated prefix: s-pd + s-all")
	assert.Equal(t, int64(4), ipv6.OnlineWithIPv6, "any IPv6 attr: all sessions except s-none")
	assert.Equal(t, int64(5), stats.OnlineUsers)
	assert.InDelta(t, 80.0, ipv6.AdoptionRate, 0.01, "4 of 5 online sessions carry IPv6")
	// User provisioning dimension.
	assert.Equal(t, int64(2), ipv6.UsersWithStaticAddress, "static address: v6addr1 + v6both")
	assert.Equal(t, int64(2), ipv6.UsersWithDelegatedPrefix, "delegated prefix: v6both + v6pd")
}
