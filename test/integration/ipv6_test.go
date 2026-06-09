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
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc3162"
	"layeh.com/radius/rfc4818"
	"layeh.com/radius/rfc6911"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// TestRadiusIPv6ProvisioningEndToEnd drives the full IPv6 provisioning chain
// against the live RADIUS auth and accounting servers (backed by PostgreSQL):
//
//  1. A user is provisioned with a static IPv6 host address, a SLAAC prefix
//     pool, a statically delegated DHCPv6-PD prefix, and a delegated prefix
//     pool.
//  2. A PAP Access-Request is authenticated and the resulting Access-Accept is
//     asserted to carry every IPv6 attribute the enhancer is responsible for:
//     Framed-IPv6-Prefix (RFC 3162), Framed-IPv6-Address (RFC 6911 §2.1),
//     Framed-IPv6-Pool (RFC 3162), Delegated-IPv6-Prefix (RFC 4818 #123, M3.6)
//     and Delegated-IPv6-Prefix-Pool (RFC 6911 #171, M3.6).
//  3. The attributes the NAS would echo are replayed in an Accounting-Start, and
//     the persisted RadiusOnline and RadiusAccounting rows are asserted to carry
//     the identical IPv6 values, proving field consistency across the
//     auth -> accounting -> database path.
//
// It is intentionally serial (no t.Parallel) because the RADIUS plugin registry
// and rate limiter are process-global shared state.
func TestRadiusIPv6ProvisioningEndToEnd(t *testing.T) {
	const secret = "it-ipv6-secret"
	suffix := uniqueSuffix()
	nasIP := "10.203.0.1"
	nasID := "it-ipv6-nas-" + suffix

	require.NoError(t, h.appCtx.DB().Create(&domain.NetNas{
		ID:         common.UUIDint64(),
		Identifier: nasID,
		Ipaddr:     nasIP,
		Secret:     secret,
		VendorCode: "0",
		Status:     common.ENABLED,
	}).Error)

	// Provision values. The user keeps its own pool values (static link mode) so
	// issuance does not depend on profile inheritance, which the inheritance test
	// below exercises separately.
	const (
		staticHost    = "2001:db8:a::5"
		wantPrefix    = "2001:db8:a::5/128"
		wantDelegated = "2001:db8:1234::/48"
	)
	slaacPool := "it-slaac-" + suffix
	pdPool := "it-pd-" + suffix

	profileID := seedProfile(t, "it-ipv6-profile-"+suffix)
	username := "it-ipv6-" + suffix
	const password = "ipv6-Pw-123"
	require.NoError(t, h.appCtx.DB().Create(&domain.RadiusUser{
		ID:                      common.UUIDint64(),
		ProfileId:               profileID,
		Username:                username,
		Password:                password,
		Status:                  common.ENABLED,
		ProfileLinkMode:         domain.ProfileLinkModeStatic,
		IpV6Addr:                staticHost,
		IPv6PrefixPool:          slaacPool,
		DelegatedIpv6Prefix:     wantDelegated,
		DelegatedIpv6PrefixPool: pdPool,
		ExpireTime:              time.Now().AddDate(1, 0, 0),
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}).Error)

	authAddr := fmt.Sprintf("127.0.0.1:%d", h.cfg.Radiusd.AuthPort)
	acctAddr := fmt.Sprintf("127.0.0.1:%d", h.cfg.Radiusd.AcctPort)

	// Step 1+2: authenticate and assert the Access-Accept carries every IPv6
	// attribute with the provisioned values.
	resp := exchange(t, authAddr, secret, username, password, nasID, nasIP)
	h.radiusSvc.ReleaseAuthRateLimit(username)
	require.Equalf(t, radius.CodeAccessAccept, resp.Code, "expected Access-Accept, got %v", resp.Code)

	framedPrefix := rfc3162.FramedIPv6Prefix_Get(resp)
	require.NotNil(t, framedPrefix, "Access-Accept must carry Framed-IPv6-Prefix")
	assert.Equal(t, wantPrefix, framedPrefix.String())

	framedAddr := rfc6911.FramedIPv6Address_Get(resp)
	require.NotEmpty(t, framedAddr, "Access-Accept must carry Framed-IPv6-Address")
	assert.Equal(t, staticHost, framedAddr.String())

	assert.Equal(t, slaacPool, rfc3162.FramedIPv6Pool_GetString(resp),
		"Access-Accept must carry Framed-IPv6-Pool")

	delegated := rfc4818.DelegatedIPv6Prefix_Get(resp)
	require.NotNil(t, delegated, "Access-Accept must carry Delegated-IPv6-Prefix")
	assert.Equal(t, wantDelegated, delegated.String())

	assert.Equal(t, pdPool, rfc6911.DelegatedIPv6PrefixPool_GetString(resp),
		"Access-Accept must carry Delegated-IPv6-Prefix-Pool")

	// Step 3: replay the attributes the NAS echoes in an Accounting-Start and
	// assert the persisted records carry the identical IPv6 values.
	sessionID := "it-ipv6-sess-" + suffix
	acctResp := acctStartIPv6(t, acctAddr, secret, username, nasID, nasIP, sessionID,
		framedPrefix, framedAddr, delegated)
	require.Equalf(t, radius.CodeAccountingResponse, acctResp.Code,
		"expected Accounting-Response, got %v", acctResp.Code)

	online := waitForOnline(t, sessionID)
	assert.Equal(t, wantPrefix, online.FramedIpv6Prefix, "online Framed-IPv6-Prefix consistency")
	assert.Equal(t, staticHost, online.FramedIpv6Address, "online Framed-IPv6-Address consistency")
	assert.Equal(t, wantDelegated, online.DelegatedIpv6Prefix, "online Delegated-IPv6-Prefix consistency")

	var acct domain.RadiusAccounting
	require.NoError(t, h.appCtx.DB().Where("acct_session_id = ?", sessionID).First(&acct).Error)
	assert.Equal(t, wantPrefix, acct.FramedIpv6Prefix, "accounting Framed-IPv6-Prefix consistency")
	assert.Equal(t, staticHost, acct.FramedIpv6Address, "accounting Framed-IPv6-Address consistency")
	assert.Equal(t, wantDelegated, acct.DelegatedIpv6Prefix, "accounting Delegated-IPv6-Prefix consistency")
}

// TestRadiusIPv6PoolInheritanceEndToEnd proves that a user with empty pool
// fields in dynamic link mode inherits both the Framed-IPv6-Pool and the
// Delegated-IPv6-Prefix-Pool from its linked profile, resolved live through the
// running auth server's profile cache (RFC 6911 §2.4 keeps the two pools
// distinct).
func TestRadiusIPv6PoolInheritanceEndToEnd(t *testing.T) {
	const secret = "it-ipv6-inherit-secret"
	suffix := uniqueSuffix()
	nasIP := "10.203.0.2"
	nasID := "it-ipv6-inh-nas-" + suffix

	require.NoError(t, h.appCtx.DB().Create(&domain.NetNas{
		ID:         common.UUIDint64(),
		Identifier: nasID,
		Ipaddr:     nasIP,
		Secret:     secret,
		VendorCode: "0",
		Status:     common.ENABLED,
	}).Error)

	profileSlaacPool := "it-prof-slaac-" + suffix
	profilePDPool := "it-prof-pd-" + suffix
	require.NoError(t, h.appCtx.DB().Create(&domain.RadiusProfile{
		ID:                      common.UUIDint64(),
		Name:                    "it-ipv6-inh-profile-" + suffix,
		Status:                  common.ENABLED,
		ActiveNum:               1,
		UpRate:                  1000,
		DownRate:                2000,
		IPv6PrefixPool:          profileSlaacPool,
		DelegatedIpv6PrefixPool: profilePDPool,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}).Error)
	var profile domain.RadiusProfile
	require.NoError(t, h.appCtx.DB().Where("name = ?", "it-ipv6-inh-profile-"+suffix).First(&profile).Error)

	username := "it-ipv6-inh-" + suffix
	const password = "ipv6-inh-Pw-123"
	require.NoError(t, h.appCtx.DB().Create(&domain.RadiusUser{
		ID:              common.UUIDint64(),
		ProfileId:       profile.ID,
		Username:        username,
		Password:        password,
		Status:          common.ENABLED,
		ProfileLinkMode: domain.ProfileLinkModeDynamic,
		IpV6Addr:        "2001:db8:b::9",
		ExpireTime:      time.Now().AddDate(1, 0, 0),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}).Error)

	authAddr := fmt.Sprintf("127.0.0.1:%d", h.cfg.Radiusd.AuthPort)
	resp := exchange(t, authAddr, secret, username, password, nasID, nasIP)
	h.radiusSvc.ReleaseAuthRateLimit(username)
	require.Equalf(t, radius.CodeAccessAccept, resp.Code, "expected Access-Accept, got %v", resp.Code)

	assert.Equal(t, profileSlaacPool, rfc3162.FramedIPv6Pool_GetString(resp),
		"Framed-IPv6-Pool must be inherited from the profile")
	assert.Equal(t, profilePDPool, rfc6911.DelegatedIPv6PrefixPool_GetString(resp),
		"Delegated-IPv6-Prefix-Pool must be inherited from the profile")
}

// acctStartIPv6 sends a single Accounting-Start carrying the supplied IPv6
// attributes (as a NAS would echo them from an Access-Accept) and returns the
// Accounting-Response. radius.Exchange signs the request authenticator with the
// shared secret, which the server validates before persisting the session.
func acctStartIPv6(t *testing.T, serverAddr, secret, username, nasID, nasIP, sessionID string,
	framedPrefix *net.IPNet, framedAddr net.IP, delegated *net.IPNet) *radius.Packet {
	t.Helper()
	packet := radius.New(radius.CodeAccountingRequest, []byte(secret))
	_ = rfc2866.AcctStatusType_Set(packet, rfc2866.AcctStatusType_Value_Start)
	_ = rfc2866.AcctSessionID_SetString(packet, sessionID)
	_ = rfc2865.UserName_SetString(packet, username)
	_ = rfc2865.NASIdentifier_SetString(packet, nasID)
	_ = rfc2865.NASIPAddress_Set(packet, net.ParseIP(nasIP))
	if framedPrefix != nil {
		_ = rfc3162.FramedIPv6Prefix_Set(packet, framedPrefix)
	}
	if len(framedAddr) > 0 {
		_ = rfc6911.FramedIPv6Address_Set(packet, framedAddr)
	}
	if delegated != nil {
		_ = rfc4818.DelegatedIPv6Prefix_Set(packet, delegated)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := radius.Exchange(ctx, packet, serverAddr)
	require.NoError(t, err)
	return resp
}

// waitForOnline polls for the RadiusOnline row created by the asynchronous
// accounting pipeline, failing the test if it does not appear within the
// deadline instead of racing a fixed sleep.
func waitForOnline(t *testing.T, sessionID string) domain.RadiusOnline {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var online domain.RadiusOnline
	for time.Now().Before(deadline) {
		err := h.appCtx.DB().Where("acct_session_id = ?", sessionID).First(&online).Error
		if err == nil {
			return online
		}
		time.Sleep(100 * time.Millisecond)
	}
	require.FailNowf(t, "online session not persisted", "no RadiusOnline row for acct_session_id=%s", sessionID)
	return online
}
