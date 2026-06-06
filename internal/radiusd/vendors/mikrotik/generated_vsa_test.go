package mikrotik

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// vsaVendorCount returns how many Vendor-Specific attributes on p carry the
// given IANA enterprise number. The accept enhancer ships Mikrotik-Rate-Limit
// as a VSA, so the codec must encode it under attribute 26 with vendor 14988.
func vsaVendorCount(t *testing.T, p *radius.Packet, wantVendorID uint32) int {
	t.Helper()
	var n int
	for _, avp := range p.Attributes {
		if avp.Type != rfc2865.VendorSpecific_Type {
			continue
		}
		vendorID, _, err := radius.VendorSpecific(avp.Attribute)
		require.NoError(t, err)
		if vendorID == wantVendorID {
			n++
		}
	}
	return n
}

// TestMikrotikRateLimitStringVSARoundTrip pins the wire contract for the string
// Mikrotik-Rate-Limit attribute the accept enhancer ships (e.g. "1000k/2000k"):
// encode under a Mikrotik (14988) VSA and decode the same string back.
func TestMikrotikRateLimitStringVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	require.NoError(t, MikrotikRateLimit_SetString(p, "1000k/2000k"))

	assert.Equal(t, 1, vsaVendorCount(t, p, _Mikrotik_VendorID))
	assert.Equal(t, "1000k/2000k", MikrotikRateLimit_GetString(p))
}

// TestMikrotikIntegerVSARoundTrip exercises the integer value codec and the
// delete path on a Mikrotik VSA.
func TestMikrotikIntegerVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	require.NoError(t, MikrotikRecvLimit_Set(p, MikrotikRecvLimit(8192)))
	assert.Equal(t, 1, vsaVendorCount(t, p, _Mikrotik_VendorID))
	assert.Equal(t, MikrotikRecvLimit(8192), MikrotikRecvLimit_Get(p))

	gets, err := MikrotikRecvLimit_Gets(p)
	require.NoError(t, err)
	assert.Equal(t, []MikrotikRecvLimit{8192}, gets)

	MikrotikRecvLimit_Del(p)
	assert.Equal(t, 0, vsaVendorCount(t, p, _Mikrotik_VendorID))
	assert.Equal(t, MikrotikRecvLimit(0), MikrotikRecvLimit_Get(p))
}
