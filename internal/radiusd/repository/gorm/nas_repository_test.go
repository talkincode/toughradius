package gorm

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

func newNasTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.NetNas{}))
	return db
}

func seedNas(t *testing.T, db *gorm.DB, nas *domain.NetNas) {
	t.Helper()
	if nas.ID == 0 {
		nas.ID = 1
	}
	require.NoError(t, db.Create(nas).Error)
}

func TestNasRepository_GetByIP(t *testing.T) {
	db := newNasTestDB(t)
	repo := NewGormNasRepository(db)
	ctx := context.Background()
	seedNas(t, db, &domain.NetNas{Name: "nas1", Ipaddr: "10.0.0.1", Identifier: "id-1", Secret: "s"})

	got, err := repo.GetByIP(ctx, "10.0.0.1")
	require.NoError(t, err)
	require.Equal(t, "nas1", got.Name)

	_, err = repo.GetByIP(ctx, "10.0.0.99")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestNasRepository_GetByIdentifier(t *testing.T) {
	db := newNasTestDB(t)
	repo := NewGormNasRepository(db)
	ctx := context.Background()
	seedNas(t, db, &domain.NetNas{Name: "nas1", Ipaddr: "10.0.0.1", Identifier: "id-1"})

	got, err := repo.GetByIdentifier(ctx, "id-1")
	require.NoError(t, err)
	require.Equal(t, "nas1", got.Name)

	_, err = repo.GetByIdentifier(ctx, "id-missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestNasRepository_GetByIPOrIdentifier(t *testing.T) {
	db := newNasTestDB(t)
	repo := NewGormNasRepository(db)
	ctx := context.Background()
	seedNas(t, db, &domain.NetNas{ID: 1, Name: "by-ip", Ipaddr: "10.0.0.1", Identifier: "id-1"})
	seedNas(t, db, &domain.NetNas{ID: 2, Name: "by-id", Ipaddr: "10.0.0.2", Identifier: "id-2"})

	// Matches on IP.
	got, err := repo.GetByIPOrIdentifier(ctx, "10.0.0.1", "no-such-id")
	require.NoError(t, err)
	require.Equal(t, "by-ip", got.Name)

	// Matches on identifier when IP does not match.
	got, err = repo.GetByIPOrIdentifier(ctx, "192.168.0.1", "id-2")
	require.NoError(t, err)
	require.Equal(t, "by-id", got.Name)

	// Neither matches.
	_, err = repo.GetByIPOrIdentifier(ctx, "192.168.0.1", "id-missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
