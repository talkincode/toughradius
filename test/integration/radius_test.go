//go:build integration

package integration

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// TestRadiusPAPAuthentication drives the running RADIUS auth server (backed by
// PostgreSQL) with a real PAP Access-Request and asserts Accept/Reject outcomes.
// It is intentionally serial (no t.Parallel) because the RADIUS plugin registry
// and rate limiter are process-global shared state.
func TestRadiusPAPAuthentication(t *testing.T) {
	const secret = "it-radius-secret"
	suffix := uniqueSuffix()
	nasIP := "10.200.0.1"
	nasID := "it-nas-" + suffix

	nas := &domain.NetNas{
		ID:         common.UUIDint64(),
		Identifier: nasID,
		Ipaddr:     nasIP,
		Secret:     secret,
		VendorCode: "0",
		Status:     common.ENABLED,
	}
	require.NoError(t, h.appCtx.DB().Create(nas).Error)

	profileID := seedProfile(t, "it-radius-profile-"+suffix)
	username := "it-radius-" + suffix
	const password = "radius-Pw-123"
	user := &domain.RadiusUser{
		ID:         common.UUIDint64(),
		ProfileId:  profileID,
		Username:   username,
		Password:   password,
		Status:     common.ENABLED,
		ExpireTime: time.Now().AddDate(1, 0, 0),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(user).Error)

	serverAddr := fmt.Sprintf("127.0.0.1:%d", h.cfg.Radiusd.AuthPort)

	t.Run("accept valid credentials", func(t *testing.T) {
		resp := exchange(t, serverAddr, secret, username, password, nasID, nasIP)
		assert.Equalf(t, radius.CodeAccessAccept, resp.Code, "expected Access-Accept, got %v", resp.Code)
		h.radiusSvc.ReleaseAuthRateLimit(username)
	})

	t.Run("reject wrong password", func(t *testing.T) {
		resp := exchange(t, serverAddr, secret, username, "wrong-password", nasID, nasIP)
		assert.Equalf(t, radius.CodeAccessReject, resp.Code, "expected Access-Reject, got %v", resp.Code)
		h.radiusSvc.ReleaseAuthRateLimit(username)
	})
}

// exchange sends a single PAP Access-Request with a bounded timeout so a stuck
// server fails the test fast instead of hanging.
func exchange(t *testing.T, serverAddr, secret, username, password, nasID, nasIP string) *radius.Packet {
	t.Helper()
	packet := radius.New(radius.CodeAccessRequest, []byte(secret))
	_ = rfc2865.UserName_SetString(packet, username)
	_ = rfc2865.UserPassword_SetString(packet, password)
	_ = rfc2865.NASIdentifier_SetString(packet, nasID)
	_ = rfc2865.NASIPAddress_Set(packet, net.ParseIP(nasIP))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := radius.Exchange(ctx, packet, serverAddr)
	require.NoError(t, err)
	return resp
}
