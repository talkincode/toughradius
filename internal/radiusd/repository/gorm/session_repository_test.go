package gorm

import (
	"context"
	"fmt"
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

func TestSessionRepository_Update(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: "sess-1"})
	require.NoError(t, err)

	err = repo.Update(ctx, &domain.RadiusOnline{
		AcctSessionId:     "sess-1",
		AcctInputTotal:    111,
		AcctOutputTotal:   222,
		AcctInputPackets:  3,
		AcctOutputPackets: 4,
		AcctSessionTime:   55,
	})
	require.NoError(t, err)

	got, err := repo.GetBySessionId(ctx, "sess-1")
	require.NoError(t, err)
	require.Equal(t, int64(111), got.AcctInputTotal)
	require.Equal(t, int64(222), got.AcctOutputTotal)
	require.Equal(t, 3, got.AcctInputPackets)
	require.Equal(t, 4, got.AcctOutputPackets)
	require.Equal(t, 55, got.AcctSessionTime)
}

func TestSessionRepository_GetBySessionId_NotFound(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	_, err := repo.GetBySessionId(ctx, "missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestSessionRepository_Exists(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	exists, err := repo.Exists(ctx, "sess-1")
	require.NoError(t, err)
	require.False(t, exists)

	_, err = repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: "sess-1"})
	require.NoError(t, err)

	exists, err = repo.Exists(ctx, "sess-1")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestSessionRepository_CountByUsername(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	for _, sid := range []string{"a", "b"} {
		_, err := repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: sid})
		require.NoError(t, err)
	}
	_, err := repo.Create(ctx, &domain.RadiusOnline{Username: "u2", AcctSessionId: "c"})
	require.NoError(t, err)

	count, err := repo.CountByUsername(ctx, "u1")
	require.NoError(t, err)
	require.Equal(t, 2, count)

	count, err = repo.CountByUsername(ctx, "u2")
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

// TestSessionRepository_CountByUsername_CacheInvalidation verifies the count
// cache is refreshed when a new session is created for the user (Create calls
// invalidate), so callers never read a stale count after a state change.
func TestSessionRepository_CountByUsername_CacheInvalidation(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: "a"})
	require.NoError(t, err)

	count, err := repo.CountByUsername(ctx, "u1") // populates the cache
	require.NoError(t, err)
	require.Equal(t, 1, count)

	_, err = repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: "b"})
	require.NoError(t, err)

	count, err = repo.CountByUsername(ctx, "u1")
	require.NoError(t, err)
	require.Equal(t, 2, count, "count must reflect the new session, not a stale cached value")
}

func TestSessionRepository_BatchDelete(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	ids := []string{}
	for _, sid := range []string{"a", "b", "c"} {
		s := &domain.RadiusOnline{Username: "u1", AcctSessionId: sid}
		_, err := repo.Create(ctx, s)
		require.NoError(t, err)
		ids = append(ids, fmt.Sprintf("%d", s.ID))
	}

	require.NoError(t, repo.BatchDelete(ctx, ids[:2]))

	var total int64
	require.NoError(t, db.Model(&domain.RadiusOnline{}).Count(&total).Error)
	require.Equal(t, int64(1), total)
}

func TestSessionRepository_BatchDeleteByNas(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: "a", NasAddr: "10.0.0.1", NasId: "nasA"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.RadiusOnline{Username: "u2", AcctSessionId: "b", NasAddr: "10.0.0.2", NasId: "nasB"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.RadiusOnline{Username: "u3", AcctSessionId: "c", NasAddr: "10.0.0.1", NasId: "nasA"})
	require.NoError(t, err)

	// Delete by NAS address removes both sessions on 10.0.0.1.
	require.NoError(t, repo.BatchDeleteByNas(ctx, "10.0.0.1", ""))

	var total int64
	require.NoError(t, db.Model(&domain.RadiusOnline{}).Count(&total).Error)
	require.Equal(t, int64(1), total)

	// Delete by NAS id removes the remaining session on nasB.
	require.NoError(t, repo.BatchDeleteByNas(ctx, "", "nasB"))
	require.NoError(t, db.Model(&domain.RadiusOnline{}).Count(&total).Error)
	require.Equal(t, int64(0), total)
}

// TestSessionRepository_BatchDeleteByNas_NoArgs verifies that passing neither a
// NAS address nor a NAS id is a safe no-op (does not delete everything).
func TestSessionRepository_BatchDeleteByNas_NoArgs(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, &domain.RadiusOnline{Username: "u1", AcctSessionId: "a", NasAddr: "10.0.0.1"})
	require.NoError(t, err)

	require.NoError(t, repo.BatchDeleteByNas(ctx, "", ""))

	var total int64
	require.NoError(t, db.Model(&domain.RadiusOnline{}).Count(&total).Error)
	require.Equal(t, int64(1), total, "no-arg BatchDeleteByNas must not delete any rows")
}

// TestSessionRepository_CountByUsername_Empty exercises the non-cached path used
// for an empty username (e.g. global online count style queries).
func TestSessionRepository_CountByUsername_Empty(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	count, err := repo.CountByUsername(ctx, "")
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

// TestSessionRepository_Delete_MissingSession verifies deleting an unknown
// session id is a no-op without error (lookupUsernameBySession finds nothing).
func TestSessionRepository_Delete_MissingSession(t *testing.T) {
	db := newSessionTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Delete(ctx, "does-not-exist"))
}
