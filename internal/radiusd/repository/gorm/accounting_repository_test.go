package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

func newAccountingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.RadiusAccounting{}))
	return db
}

func TestAccountingRepository_Create_AssignsID(t *testing.T) {
	db := newAccountingTestDB(t)
	repo := NewGormAccountingRepository(db)
	ctx := context.Background()

	acct := &domain.RadiusAccounting{Username: "u1", AcctSessionId: "sess-1"}
	require.NoError(t, repo.Create(ctx, acct))
	require.NotZero(t, acct.ID, "Create should assign a primary key when ID is zero")

	var stored domain.RadiusAccounting
	require.NoError(t, db.Where("acct_session_id = ?", "sess-1").First(&stored).Error)
	require.Equal(t, acct.ID, stored.ID)
}

func TestAccountingRepository_Create_PreservesProvidedID(t *testing.T) {
	db := newAccountingTestDB(t)
	repo := NewGormAccountingRepository(db)
	ctx := context.Background()

	acct := &domain.RadiusAccounting{ID: 42, Username: "u1", AcctSessionId: "sess-1"}
	require.NoError(t, repo.Create(ctx, acct))
	require.Equal(t, int64(42), acct.ID)
}

func TestAccountingRepository_UpdateStop_Success(t *testing.T) {
	db := newAccountingTestDB(t)
	repo := NewGormAccountingRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &domain.RadiusAccounting{
		Username:      "u1",
		AcctSessionId: "sess-1",
		AcctStartTime: time.Now().Add(-time.Hour),
	}))

	err := repo.UpdateStop(ctx, "sess-1", &domain.RadiusAccounting{
		AcctInputTotal:    1000,
		AcctOutputTotal:   2000,
		AcctInputPackets:  10,
		AcctOutputPackets: 20,
		AcctSessionTime:   3600,
	})
	require.NoError(t, err)

	var stored domain.RadiusAccounting
	require.NoError(t, db.Where("acct_session_id = ?", "sess-1").First(&stored).Error)
	require.Equal(t, int64(1000), stored.AcctInputTotal)
	require.Equal(t, int64(2000), stored.AcctOutputTotal)
	require.Equal(t, 10, stored.AcctInputPackets)
	require.Equal(t, 20, stored.AcctOutputPackets)
	require.Equal(t, 3600, stored.AcctSessionTime)
	require.False(t, stored.AcctStopTime.IsZero(), "stop time should be set")
}

func TestAccountingRepository_UpdateStop_NoMatchingSession(t *testing.T) {
	db := newAccountingTestDB(t)
	repo := NewGormAccountingRepository(db)
	ctx := context.Background()

	err := repo.UpdateStop(ctx, "missing", &domain.RadiusAccounting{})
	require.Error(t, err, "stopping an unknown session must return an error")
}

// TestAccountingRepository_Create_AllowsDuplicateSessionId documents that the
// accounting (history) table intentionally has no unique index on
// acct_session_id (unlike radius_online), so a session id may legitimately
// recur across the retained history.
func TestAccountingRepository_Create_AllowsDuplicateSessionId(t *testing.T) {
	db := newAccountingTestDB(t)
	repo := NewGormAccountingRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &domain.RadiusAccounting{Username: "u1", AcctSessionId: "sess-1"}))
	require.NoError(t, repo.Create(ctx, &domain.RadiusAccounting{Username: "u1", AcctSessionId: "sess-1"}))

	var count int64
	require.NoError(t, db.Model(&domain.RadiusAccounting{}).
		Where("acct_session_id = ?", "sess-1").Count(&count).Error)
	require.Equal(t, int64(2), count)
}
