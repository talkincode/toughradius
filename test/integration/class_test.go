//go:build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// TestRadiusClassInAccessAccept checks that a configured RFC 2865 Class
// (Type 25) is copied into Access-Accept as-is, and omitted when unset.
func TestRadiusClassInAccessAccept(t *testing.T) {
	const secret = "it-radius-secret"
	suffix := uniqueSuffix()
	nasIP := "10.200.0.21"
	nasID := "it-class-nas-" + suffix

	nas := &domain.NetNas{
		ID:         common.UUIDint64(),
		Identifier: nasID,
		Ipaddr:     nasIP,
		Secret:     secret,
		VendorCode: "0",
		Status:     common.ENABLED,
	}
	require.NoError(t, h.appCtx.DB().Create(nas).Error)

	profileID := seedProfile(t, "it-class-profile-"+suffix)
	username := "it-class-" + suffix
	const password = "radius-Pw-123"
	const wantClass = "OU=vpn-users"
	user := &domain.RadiusUser{
		ID:          common.UUIDint64(),
		ProfileId:   profileID,
		Username:    username,
		Password:    password,
		Status:      common.ENABLED,
		RadiusClass: wantClass,
		ExpireTime:  time.Now().AddDate(1, 0, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(user).Error)

	plainUser := "it-class-empty-" + suffix
	empty := &domain.RadiusUser{
		ID:         common.UUIDint64(),
		ProfileId:  profileID,
		Username:   plainUser,
		Password:   password,
		Status:     common.ENABLED,
		ExpireTime: time.Now().AddDate(1, 0, 0),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(empty).Error)

	serverAddr := fmt.Sprintf("127.0.0.1:%d", h.cfg.Radiusd.AuthPort)

	t.Run("accept includes configured class", func(t *testing.T) {
		resp := exchange(t, serverAddr, secret, username, password, nasID, nasIP)
		assert.Equalf(t, radius.CodeAccessAccept, resp.Code, "expected Access-Accept, got %v", resp.Code)
		assert.Equal(t, wantClass, rfc2865.Class_GetString(resp))
		h.radiusSvc.ReleaseAuthRateLimit(username)
	})

	t.Run("accept omits empty class", func(t *testing.T) {
		resp := exchange(t, serverAddr, secret, plainUser, password, nasID, nasIP)
		assert.Equalf(t, radius.CodeAccessAccept, resp.Code, "expected Access-Accept, got %v", resp.Code)
		assert.Empty(t, rfc2865.Class_GetString(resp))
		h.radiusSvc.ReleaseAuthRateLimit(plainUser)
	})
}
