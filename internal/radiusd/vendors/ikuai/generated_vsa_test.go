package ikuai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// vsaVendorCount returns how many Vendor-Specific attributes on p carry the
// given IANA enterprise number. The accept enhancer ships iKuai speed limits as
// VSAs, so the codec must encode them under attribute 26 with vendor 10055.
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

// TestIKuaiSpeedVSARoundTrip pins the wire contract for the iKuai speed-limit
// attributes the accept enhancer ships: encode under iKuai (10055) VSAs and
// decode the same values back.
func TestIKuaiSpeedVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	require.NoError(t, RPUpstreamSpeedLimit_Set(p, RPUpstreamSpeedLimit(100)))
	require.NoError(t, RPDownstreamSpeedLimit_Set(p, RPDownstreamSpeedLimit(200)))

	assert.Equal(t, 2, vsaVendorCount(t, p, _IKuai_VendorID))
	assert.Equal(t, RPUpstreamSpeedLimit(100), RPUpstreamSpeedLimit_Get(p))
	assert.Equal(t, RPDownstreamSpeedLimit(200), RPDownstreamSpeedLimit_Get(p))

	gets, err := RPUpstreamSpeedLimit_Gets(p)
	require.NoError(t, err)
	assert.Equal(t, []RPUpstreamSpeedLimit{100}, gets)

	RPDownstreamSpeedLimit_Del(p)
	assert.Equal(t, RPDownstreamSpeedLimit(0), RPDownstreamSpeedLimit_Get(p))
	assert.Equal(t, 1, vsaVendorCount(t, p, _IKuai_VendorID), "only the upstream VSA should remain")
}
