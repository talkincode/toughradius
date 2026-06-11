package app

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// TestSchedClearExpireData verifies the daily cleanup removes only the stale
// rows: radius_online sessions whose LastUpdate is older than 300 seconds, and
// radius_accounting rows whose AcctStopTime predates the configured retention
// window; fresh/recent rows must survive.
func TestSchedClearExpireData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.RadiusOnline{}, &domain.RadiusAccounting{}))

	now := time.Now()

	// radius_online: one stale (>300s), one fresh (<300s).
	require.NoError(t, db.Create(&domain.RadiusOnline{
		Username: "stale", AcctSessionId: "sess-stale", LastUpdate: now.Add(-600 * time.Second),
	}).Error)
	require.NoError(t, db.Create(&domain.RadiusOnline{
		Username: "fresh", AcctSessionId: "sess-fresh", LastUpdate: now.Add(-60 * time.Second),
	}).Error)

	// radius_accounting with a 30-day retention: one old (40d), one recent (5d).
	require.NoError(t, db.Create(&domain.RadiusAccounting{
		Username: "old", AcctStopTime: now.AddDate(0, 0, -40),
	}).Error)
	require.NoError(t, db.Create(&domain.RadiusAccounting{
		Username: "recent", AcctStopTime: now.AddDate(0, 0, -5),
	}).Error)

	cm := &ConfigManager{
		configs: map[string]string{"radius.AccountingHistoryDays": "30"},
		schemas: make(map[string]*ConfigSchema),
	}
	a := &Application{gormDB: db, configManager: cm}

	a.SchedClearExpireData()

	var online []domain.RadiusOnline
	require.NoError(t, db.Find(&online).Error)
	require.Len(t, online, 1, "only the fresh online session should remain")
	require.Equal(t, "fresh", online[0].Username)

	var acct []domain.RadiusAccounting
	require.NoError(t, db.Find(&acct).Error)
	require.Len(t, acct, 1, "only the recent accounting record should remain")
	require.Equal(t, "recent", acct[0].Username)
}

// TestSchedClearExpireData_DisabledRetention verifies that AccountingHistoryDays=0
// disables accounting cleanup (matching the config schema's "0=disabled"), so even
// very old accounting rows are kept — while the independent radius_online cleanup
// still removes stale sessions.
func TestSchedClearExpireData_DisabledRetention(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.RadiusOnline{}, &domain.RadiusAccounting{}))

	now := time.Now()
	require.NoError(t, db.Create(&domain.RadiusAccounting{
		Username: "ancient", AcctStopTime: now.AddDate(0, 0, -1000),
	}).Error)
	require.NoError(t, db.Create(&domain.RadiusOnline{
		Username: "stale", AcctSessionId: "sess-stale", LastUpdate: now.Add(-600 * time.Second),
	}).Error)

	cm := &ConfigManager{
		configs: map[string]string{"radius.AccountingHistoryDays": "0"},
		schemas: make(map[string]*ConfigSchema),
	}
	a := &Application{gormDB: db, configManager: cm}

	a.SchedClearExpireData()

	var acctCount int64
	require.NoError(t, db.Model(&domain.RadiusAccounting{}).Count(&acctCount).Error)
	require.Equal(t, int64(1), acctCount, "AccountingHistoryDays=0 must disable accounting cleanup")

	var onlineCount int64
	require.NoError(t, db.Model(&domain.RadiusOnline{}).Count(&onlineCount).Error)
	require.Equal(t, int64(0), onlineCount, "online cleanup runs regardless of AccountingHistoryDays")
}

// TestInitJobRegistersCleanup is the regression guard for M6.4: SchedClearExpireData
// was defined but never registered with cron, so expired online/accounting data was
// never auto-purged. initJob must now schedule it alongside the monitor and
// operation-log jobs (3 cron entries total).
func TestInitJobRegistersCleanup(t *testing.T) {
	a := &Application{appConfig: &config.AppConfig{}}
	a.initJob()
	defer a.sched.Stop()

	require.Len(t, a.sched.Entries(), 3,
		"expected monitor + operation-log cleanup + expire-data cleanup cron entries")
}
