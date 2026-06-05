package radiusd

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/domain"
	repogorm "github.com/talkincode/toughradius/v9/internal/radiusd/repository/gorm"
)

func newUpdateBindTestService(t *testing.T) (*AuthService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.RadiusUser{}))
	svc := &AuthService{RadiusService: &RadiusService{
		UserRepo: repogorm.NewGormUserRepository(db),
	}}
	return svc, db
}

// TestUpdateBindPersistsBothVlanIds verifies that UpdateBind writes the request
// VLAN IDs to the matching columns (vlanid1 -> vlanid1, vlanid2 -> vlanid2).
func TestUpdateBindPersistsBothVlanIds(t *testing.T) {
	svc, db := newUpdateBindTestService(t)

	user := &domain.RadiusUser{Username: "u1", Vlanid1: 10, Vlanid2: 20, MacAddr: "aa:bb"}
	require.NoError(t, db.Create(user).Error)

	svc.UpdateBind(user, &VendorRequest{Vlanid1: 100, Vlanid2: 200, MacAddr: "cc:dd"})

	var got domain.RadiusUser
	require.NoError(t, db.Where("username = ?", "u1").First(&got).Error)
	require.Equal(t, 100, got.Vlanid1)
	require.Equal(t, 200, got.Vlanid2)
	require.Equal(t, "cc:dd", got.MacAddr)
}

// TestUpdateBindVlanid1ChangeKeepsVlanid2 is the regression for the original bug
// where a vlanid1 change was written into vlanid2 (and vlanid1 was never saved).
func TestUpdateBindVlanid1ChangeKeepsVlanid2(t *testing.T) {
	svc, db := newUpdateBindTestService(t)

	user := &domain.RadiusUser{Username: "u2", Vlanid1: 10, Vlanid2: 20}
	require.NoError(t, db.Create(user).Error)

	// Only vlanid1 differs from the stored value.
	svc.UpdateBind(user, &VendorRequest{Vlanid1: 99, Vlanid2: 20})

	var got domain.RadiusUser
	require.NoError(t, db.Where("username = ?", "u2").First(&got).Error)
	require.Equal(t, 99, got.Vlanid1, "vlanid1 must be persisted to vlanid1")
	require.Equal(t, 20, got.Vlanid2, "vlanid2 must be preserved")
}
