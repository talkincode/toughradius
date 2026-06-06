package app

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// TestDedupOnlineSessions verifies the pre-migration cleanup removes duplicate
// radius_online rows (left over from before the unique index existed) so that
// AutoMigrate can subsequently create the unique index on acct_session_id.
func TestDedupOnlineSessions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Simulate a legacy table without the unique index, holding duplicates.
	require.NoError(t, db.Exec(
		`CREATE TABLE radius_online (id INTEGER PRIMARY KEY, acct_session_id TEXT, username TEXT)`).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO radius_online (id, acct_session_id, username) VALUES
			(1,'sess-1','u'),(2,'sess-1','u'),(3,'sess-2','u'),(4,'sess-2','u'),(5,'sess-3','u')`).Error)

	a := &Application{gormDB: db}
	a.dedupOnlineSessions()

	var total int64
	require.NoError(t, db.Raw(`SELECT COUNT(1) FROM radius_online`).Scan(&total).Error)
	require.Equal(t, int64(3), total, "one row should remain per acct_session_id")

	// After dedup, AutoMigrate must be able to add the unique index cleanly.
	require.NoError(t, db.AutoMigrate(&domain.RadiusOnline{}))

	// And the unique index must now reject a duplicate acct_session_id.
	err = db.Exec(
		`INSERT INTO radius_online (id, acct_session_id, username) VALUES (99,'sess-1','u')`).Error
	require.Error(t, err, "unique index should reject duplicate acct_session_id after migration")
}

// TestDedupOnlineSessions_NoTable verifies the cleanup is a no-op (no panic,
// no error) when the table does not exist yet, e.g. on a fresh install.
func TestDedupOnlineSessions_NoTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	a := &Application{gormDB: db}
	require.NotPanics(t, a.dedupOnlineSessions)
}

// TestUpgradePath_LegacyNonUniqueIndex reproduces the upgrade scenario from the
// PR review: a deployment that already has the legacy non-unique index
// (idx_radius_online_acct_session_id) plus duplicate rows. After dedup + legacy
// index drop, AutoMigrate must create the unique index so ON CONFLICT works.
func TestUpgradePath_LegacyNonUniqueIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Legacy schema: non-unique index named like GORM's gorm:"index" default.
	require.NoError(t, db.Exec(
		`CREATE TABLE radius_online (id INTEGER PRIMARY KEY, acct_session_id TEXT, username TEXT)`).Error)
	require.NoError(t, db.Exec(
		`CREATE INDEX idx_radius_online_acct_session_id ON radius_online (acct_session_id)`).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO radius_online (id, acct_session_id, username) VALUES
			(1,'sess-1','u'),(2,'sess-1','u'),(3,'sess-2','u')`).Error)

	a := &Application{gormDB: db}
	a.dedupOnlineSessions()
	a.dropLegacyOnlineSessionIndex()

	require.False(t, db.Migrator().HasIndex(&domain.RadiusOnline{}, "idx_radius_online_acct_session_id"),
		"legacy non-unique index should be dropped")
	require.NoError(t, db.AutoMigrate(&domain.RadiusOnline{}))
	require.True(t, db.Migrator().HasIndex(&domain.RadiusOnline{}, "udx_radius_online_acct_session_id"),
		"unique index should be created after migration")

	err = db.Exec(
		`INSERT INTO radius_online (id, acct_session_id, username) VALUES (99,'sess-1','u')`).Error
	require.Error(t, err, "unique index must reject duplicate acct_session_id on the upgrade path")
}
