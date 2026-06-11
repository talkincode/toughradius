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

// TestSchedClearExpireData verifies the daily cleanup removes only genuinely
// stale rows: radius_online sessions that have missed several interim updates,
// and terminated radius_accounting rows whose AcctStopTime predates the
// retention window. Live sessions (recent online rows) and active accounting
// rows (zero AcctStopTime) must always survive.
func TestSchedClearExpireData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.RadiusOnline{}, &domain.RadiusAccounting{}))

	now := time.Now()

	// radius_online with the default 300s interim -> 900s stale window:
	//   dangling (20m, missed several interims) is deleted; live sessions kept.
	require.NoError(t, db.Create(&domain.RadiusOnline{
		Username: "dangling", AcctSessionId: "sess-dangling", LastUpdate: now.Add(-20 * time.Minute),
	}).Error)
	require.NoError(t, db.Create(&domain.RadiusOnline{
		Username: "live-recent", AcctSessionId: "sess-recent", LastUpdate: now.Add(-60 * time.Second),
	}).Error)
	require.NoError(t, db.Create(&domain.RadiusOnline{
		Username: "live-quiet", AcctSessionId: "sess-quiet", LastUpdate: now.Add(-10 * time.Minute),
	}).Error)

	// radius_accounting with a 30-day retention.
	require.NoError(t, db.Create(&domain.RadiusAccounting{
		Username: "old-stopped", AcctStopTime: now.AddDate(0, 0, -40),
	}).Error)
	require.NoError(t, db.Create(&domain.RadiusAccounting{
		Username: "recent-stopped", AcctStopTime: now.AddDate(0, 0, -5),
	}).Error)
	// Active session: row created at Accounting-Start with a zero AcctStopTime;
	// must NOT be purged or its billing history is lost permanently.
	require.NoError(t, db.Create(&domain.RadiusAccounting{
		Username: "active", AcctStartTime: now.AddDate(0, 0, -40),
	}).Error)

	cm := &ConfigManager{
		configs: map[string]string{"radius.AccountingHistoryDays": "30"},
		schemas: make(map[string]*ConfigSchema),
	}
	a := &Application{gormDB: db, configManager: cm}

	a.SchedClearExpireData()

	require.ElementsMatch(t, []string{"live-recent", "live-quiet"}, onlineUsernames(t, db),
		"only dangling online sessions should be removed; live sessions must remain")
	require.ElementsMatch(t, []string{"recent-stopped", "active"}, accountingUsernames(t, db),
		"only terminated records past retention should be removed; active session must remain")
}

// TestSchedClearExpireData_DisabledRetention verifies that AccountingHistoryDays=0
// disables accounting cleanup (matching the config schema's "0=disabled"), so even
// very old terminated rows and active rows are kept — while the independent
// radius_online cleanup still removes dangling sessions.
func TestSchedClearExpireData_DisabledRetention(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.RadiusOnline{}, &domain.RadiusAccounting{}))

	now := time.Now()
	require.NoError(t, db.Create(&domain.RadiusAccounting{
		Username: "ancient", AcctStopTime: now.AddDate(0, 0, -1000),
	}).Error)
	require.NoError(t, db.Create(&domain.RadiusAccounting{
		Username: "active", AcctStartTime: now.AddDate(0, 0, -1000),
	}).Error)
	require.NoError(t, db.Create(&domain.RadiusOnline{
		Username: "dangling", AcctSessionId: "sess-dangling", LastUpdate: now.Add(-20 * time.Minute),
	}).Error)

	cm := &ConfigManager{
		configs: map[string]string{"radius.AccountingHistoryDays": "0"},
		schemas: make(map[string]*ConfigSchema),
	}
	a := &Application{gormDB: db, configManager: cm}

	a.SchedClearExpireData()

	require.ElementsMatch(t, []string{"ancient", "active"}, accountingUsernames(t, db),
		"AccountingHistoryDays=0 must disable accounting cleanup entirely")

	var onlineCount int64
	require.NoError(t, db.Model(&domain.RadiusOnline{}).Count(&onlineCount).Error)
	require.Equal(t, int64(0), onlineCount, "online cleanup runs regardless of AccountingHistoryDays")
}

func onlineUsernames(t *testing.T, db *gorm.DB) []string {
	t.Helper()
	var rows []domain.RadiusOnline
	require.NoError(t, db.Find(&rows).Error)
	names := make([]string, 0, len(rows))
	for _, r := range rows {
		names = append(names, r.Username)
	}
	return names
}

func accountingUsernames(t *testing.T, db *gorm.DB) []string {
	t.Helper()
	var rows []domain.RadiusAccounting
	require.NoError(t, db.Find(&rows).Error)
	names := make([]string, 0, len(rows))
	for _, r := range rows {
		names = append(names, r.Username)
	}
	return names
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
