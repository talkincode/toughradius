package handlers

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc3162"
	"layeh.com/radius/rfc4818"
	"layeh.com/radius/rfc6911"
)

// TestStartHandler_Handle_ParsesIPv6Attributes verifies that the Accounting-Start
// handler decodes the standard IPv6 attributes into the online session and
// accounting records: Framed-IPv6-Address (RFC 6911), Framed-IPv6-Prefix
// (RFC 3162), and Delegated-IPv6-Prefix (RFC 4818).
func TestStartHandler_Handle_ParsesIPv6Attributes(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	acctRepo := newMockAccountingRepo()
	handler := NewStartHandler(sessionRepo, acctRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_Start))
	pkt := ctx.Request.Packet

	require.NoError(t, rfc6911.FramedIPv6Address_Set(pkt, net.ParseIP("2001:db8::1")))
	_, framedPrefix, err := net.ParseCIDR("2001:db8:1::/64")
	require.NoError(t, err)
	require.NoError(t, rfc3162.FramedIPv6Prefix_Set(pkt, framedPrefix))
	_, delegatedPrefix, err := net.ParseCIDR("2001:db8:dead::/48")
	require.NoError(t, err)
	require.NoError(t, rfc4818.DelegatedIPv6Prefix_Set(pkt, delegatedPrefix))

	require.NoError(t, handler.Handle(ctx))

	online := sessionRepo.sessions["test-session-123"]
	require.NotNil(t, online)
	assert.Equal(t, "2001:db8::1", online.FramedIpv6Address)
	assert.Equal(t, "2001:db8:1::/64", online.FramedIpv6Prefix)
	assert.Equal(t, "2001:db8:dead::/48", online.DelegatedIpv6Prefix)

	acct := acctRepo.records["test-session-123"]
	require.NotNil(t, acct)
	assert.Equal(t, "2001:db8::1", acct.FramedIpv6Address)
	assert.Equal(t, "2001:db8:1::/64", acct.FramedIpv6Prefix)
	assert.Equal(t, "2001:db8:dead::/48", acct.DelegatedIpv6Prefix)
}

// TestStartHandler_Handle_AbsentIPv6AttributesStoreNA verifies that when a NAS
// omits the IPv6 attributes, the persisted fields fall back to the not-available
// sentinel instead of the literal "<nil>" that a nil net.IP / *net.IPNet would
// otherwise stringify to.
func TestStartHandler_Handle_AbsentIPv6AttributesStoreNA(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	acctRepo := newMockAccountingRepo()
	handler := NewStartHandler(sessionRepo, acctRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_Start))
	require.NoError(t, handler.Handle(ctx))

	online := sessionRepo.sessions["test-session-123"]
	require.NotNil(t, online)
	assert.Equal(t, common.NA, online.FramedIpv6Address)
	assert.Equal(t, common.NA, online.FramedIpv6Prefix)
	assert.Equal(t, common.NA, online.DelegatedIpv6Prefix)
}

// TestIPv6AddrOrNA exercises the host-address normalization helper directly.
func TestIPv6AddrOrNA(t *testing.T) {
	assert.Equal(t, common.NA, ipv6AddrOrNA(nil))
	assert.Equal(t, common.NA, ipv6AddrOrNA(net.IPv6unspecified))
	assert.Equal(t, "2001:db8::1", ipv6AddrOrNA(net.ParseIP("2001:db8::1")))
}

// TestIPv6PrefixOrNA exercises the prefix normalization helper directly.
func TestIPv6PrefixOrNA(t *testing.T) {
	assert.Equal(t, common.NA, ipv6PrefixOrNA(nil))
	_, prefix, err := net.ParseCIDR("2001:db8::/32")
	require.NoError(t, err)
	assert.Equal(t, "2001:db8::/32", ipv6PrefixOrNA(prefix))
}
