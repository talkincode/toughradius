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

func newUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.RadiusUser{}))
	return db
}

func seedUser(t *testing.T, db *gorm.DB, u *domain.RadiusUser) {
	t.Helper()
	if u.ID == 0 {
		u.ID = 1
	}
	require.NoError(t, db.Create(u).Error)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := newUserTestDB(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()
	seedUser(t, db, &domain.RadiusUser{Username: "alice", Password: "pw", Status: "enabled"})

	got, err := repo.GetByUsername(ctx, "alice")
	require.NoError(t, err)
	require.Equal(t, "alice", got.Username)
	require.Equal(t, "pw", got.Password)

	_, err = repo.GetByUsername(ctx, "nobody")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepository_GetByMacAddr(t *testing.T) {
	db := newUserTestDB(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()
	seedUser(t, db, &domain.RadiusUser{Username: "bob", MacAddr: "AA:BB:CC:DD:EE:FF"})

	got, err := repo.GetByMacAddr(ctx, "AA:BB:CC:DD:EE:FF")
	require.NoError(t, err)
	require.Equal(t, "bob", got.Username)

	_, err = repo.GetByMacAddr(ctx, "00:00:00:00:00:00")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepository_UpdateMacAddr(t *testing.T) {
	db := newUserTestDB(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()
	seedUser(t, db, &domain.RadiusUser{Username: "carol"})

	require.NoError(t, repo.UpdateMacAddr(ctx, "carol", "11:22:33:44:55:66"))

	got, err := repo.GetByUsername(ctx, "carol")
	require.NoError(t, err)
	require.Equal(t, "11:22:33:44:55:66", got.MacAddr)
}

func TestUserRepository_UpdateVlanId(t *testing.T) {
	db := newUserTestDB(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()
	seedUser(t, db, &domain.RadiusUser{Username: "dave"})

	require.NoError(t, repo.UpdateVlanId(ctx, "dave", 100, 200))

	got, err := repo.GetByUsername(ctx, "dave")
	require.NoError(t, err)
	require.Equal(t, 100, got.Vlanid1)
	require.Equal(t, 200, got.Vlanid2)
}

func TestUserRepository_UpdateLastOnline(t *testing.T) {
	db := newUserTestDB(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()
	seedUser(t, db, &domain.RadiusUser{Username: "erin"})

	before := time.Now().Add(-time.Minute)
	require.NoError(t, repo.UpdateLastOnline(ctx, "erin"))

	got, err := repo.GetByUsername(ctx, "erin")
	require.NoError(t, err)
	require.True(t, got.LastOnline.After(before), "last_online should be updated to now")
}

func TestUserRepository_UpdateField(t *testing.T) {
	db := newUserTestDB(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()
	seedUser(t, db, &domain.RadiusUser{Username: "frank", Status: "enabled"})

	require.NoError(t, repo.UpdateField(ctx, "frank", "status", "disabled"))

	got, err := repo.GetByUsername(ctx, "frank")
	require.NoError(t, err)
	require.Equal(t, "disabled", got.Status)
}
