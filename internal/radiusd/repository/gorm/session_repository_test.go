package gorm

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

func newSessionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.RadiusOnline{}))
	return db
}

func countOnline(t *testing.T, db *gorm.DB, sessionId string) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&domain.RadiusOnline{}).
		Where("acct_session_id = ?", sessionId).Count(&count).Error)
	return count
}

// TestSessionRepository_Create_Idempotent reproduces the duplicate
// Accounting-Start retransmission scenario from issue #302: inserting the same
// Acct-Session-Id twice must keep exactly one online row, and the second call
// must report created=false so the caller can skip duplicate accounting.
func TestSessionRepository_Create_Idempotent(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: "sess-1"})
	require.NoError(t, err)
	require.True(t, created, "first insert should create a row")

	created, err = repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: "sess-1"})
	require.NoError(t, err)
	require.False(t, created, "retransmitted Accounting-Start must not create a duplicate")

	require.Equal(t, int64(1), countOnline(t, db, "sess-1"))
}

// TestSessionRepository_Create_DistinctSessions verifies the unique index does
// not collapse legitimately different sessions.
func TestSessionRepository_Create_DistinctSessions(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	for _, sid := range []string{"a", "b", "c"} {
		created, err := repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: sid})
		require.NoError(t, err)
		require.True(t, created)
	}

	var total int64
	require.NoError(t, db.Model(&domain.RadiusOnline{}).Count(&total).Error)
	require.Equal(t, int64(3), total)
}

// TestSessionRepository_Create_RecreateAfterDelete verifies a session id can be
// reused once the previous session is stopped (online row deleted).
func TestSessionRepository_Create_RecreateAfterDelete(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: "sess-1"})
	require.NoError(t, err)
	require.True(t, created)

	require.NoError(t, repo.Delete(ctx, "sess-1"))

	created, err = repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: "sess-1"})
	require.NoError(t, err)
	require.True(t, created, "session id should be reusable after the online row is deleted")
	require.Equal(t, int64(1), countOnline(t, db, "sess-1"))
}
