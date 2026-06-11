//go:build integration

package integration

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// signAccessRequest adds a valid RFC 3579 Message-Authenticator to a client
// Access-Request, mirroring what a compliant NAS does before sending.
func signAccessRequest(t *testing.T, packet *radius.Packet, secret string) {
	t.Helper()
	require.NoError(t, rfc2869.MessageAuthenticator_Set(packet, make([]byte, md5.Size)))
	wire, err := packet.MarshalBinary()
	require.NoError(t, err)
	mac := hmac.New(md5.New, []byte(secret))
	mac.Write(wire)
	require.NoError(t, rfc2869.MessageAuthenticator_Set(packet, mac.Sum(nil)))
}

// newAccessRequest builds a PAP Access-Request for the given identity.
func newAccessRequest(secret, username, password, nasID, nasIP string) *radius.Packet {
	packet := radius.New(radius.CodeAccessRequest, []byte(secret))
	_ = rfc2865.UserName_SetString(packet, username)
	_ = rfc2865.UserPassword_SetString(packet, password)
	_ = rfc2865.NASIdentifier_SetString(packet, nasID)
	_ = rfc2865.NASIPAddress_Set(packet, net.ParseIP(nasIP))
	return packet
}

// TestRadiusMessageAuthenticator exercises the CVE-2024-3596 (BlastRADIUS)
// hardening end-to-end: responses are signed in the default warn mode, and
// enforce mode silently discards Access-Request packets without a valid
// Message-Authenticator while still accepting compliant ones.
func TestRadiusMessageAuthenticator(t *testing.T) {
	const secret = "it-msgauth-secret"
	suffix := uniqueSuffix()
	nasIP := "10.201.0.1"
	nasID := "it-msgauth-nas-" + suffix

	nas := &domain.NetNas{
		ID:         common.UUIDint64(),
		Identifier: nasID,
		Ipaddr:     nasIP,
		Secret:     secret,
		VendorCode: "0",
		Status:     common.ENABLED,
	}
	require.NoError(t, h.appCtx.DB().Create(nas).Error)

	profileID := seedProfile(t, "it-msgauth-profile-"+suffix)
	username := "it-msgauth-" + suffix
	const password = "msgAuth-Pw-123"
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
	cfgMgr := h.appCtx.ConfigMgr()

	// Restore the default mode after the test so other suites are unaffected.
	original := cfgMgr.GetString("radius", "RequireMessageAuthenticator")
	t.Cleanup(func() { _ = cfgMgr.Set("radius", "RequireMessageAuthenticator", original) })

	t.Run("warn mode signs the accept response", func(t *testing.T) {
		require.NoError(t, cfgMgr.Set("radius", "RequireMessageAuthenticator", "warn"))

		packet := newAccessRequest(secret, username, password, nasID, nasIP)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		resp, err := radius.Exchange(ctx, packet, serverAddr)
		require.NoError(t, err)
		assert.Equal(t, radius.CodeAccessAccept, resp.Code)

		// The response must carry a Message-Authenticator that validates against
		// the shared secret using the original request authenticator.
		mac, err := rfc2869.MessageAuthenticator_Lookup(resp)
		require.NoError(t, err)
		require.Len(t, mac, md5.Size)
		assert.True(t, validResponseMessageAuthenticator(t, resp, packet.Authenticator[:], secret))

		h.radiusSvc.ReleaseAuthRateLimit(username)
	})

	t.Run("enforce accepts request with valid message-authenticator", func(t *testing.T) {
		require.NoError(t, cfgMgr.Set("radius", "RequireMessageAuthenticator", "enforce"))

		packet := newAccessRequest(secret, username, password, nasID, nasIP)
		signAccessRequest(t, packet, secret)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		resp, err := radius.Exchange(ctx, packet, serverAddr)
		require.NoError(t, err)
		assert.Equal(t, radius.CodeAccessAccept, resp.Code)

		h.radiusSvc.ReleaseAuthRateLimit(username)
	})

	t.Run("enforce discards request without message-authenticator", func(t *testing.T) {
		require.NoError(t, cfgMgr.Set("radius", "RequireMessageAuthenticator", "enforce"))

		packet := newAccessRequest(secret, username, password, nasID, nasIP)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := radius.Exchange(ctx, packet, serverAddr)
		// The server silently discards the packet (RFC 3579 §3.2), so the client
		// gets no reply and the bounded context deadline fires.
		require.Error(t, err)

		h.radiusSvc.ReleaseAuthRateLimit(username)
	})
}

// validResponseMessageAuthenticator recomputes the response Message-Authenticator
// using the request authenticator (RFC 2869 §5.14) and compares it to the value
// the server emitted.
func validResponseMessageAuthenticator(t *testing.T, resp *radius.Packet, requestAuth []byte, secret string) bool {
	t.Helper()
	received, err := rfc2869.MessageAuthenticator_Lookup(resp)
	require.NoError(t, err)

	verify := resp.Response(resp.Code)
	verify.Attributes = append(radius.Attributes(nil), resp.Attributes...)
	copy(verify.Authenticator[:], requestAuth)
	require.NoError(t, rfc2869.MessageAuthenticator_Set(verify, make([]byte, md5.Size)))
	wire, err := verify.MarshalBinary()
	require.NoError(t, err)
	mac := hmac.New(md5.New, []byte(secret))
	mac.Write(wire)
	return hmac.Equal(received, mac.Sum(nil))
}
