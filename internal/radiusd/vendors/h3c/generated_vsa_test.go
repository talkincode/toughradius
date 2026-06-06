package h3c

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// vsaVendorCount returns how many Vendor-Specific attributes on p carry the
// given IANA enterprise number. The accept enhancer ships H3C rate limits as
// VSAs, so the codec must encode them under attribute 26 with vendor 25506.
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

// TestH3CRateVSARoundTrip pins the wire contract for the H3C bandwidth
// attributes the accept enhancer ships: encode under H3C (25506) VSAs and
// decode the same values back.
func TestH3CRateVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	require.NoError(t, H3CInputAverageRate_Set(p, H3CInputAverageRate(1024)))
	require.NoError(t, H3COutputAverageRate_Set(p, H3COutputAverageRate(2048)))

	assert.Equal(t, 2, vsaVendorCount(t, p, _H3C_VendorID))
	assert.Equal(t, H3CInputAverageRate(1024), H3CInputAverageRate_Get(p))
	assert.Equal(t, H3COutputAverageRate(2048), H3COutputAverageRate_Get(p))

	gets, err := H3CInputAverageRate_Gets(p)
	require.NoError(t, err)
	assert.Equal(t, []H3CInputAverageRate{1024}, gets)

	H3CInputAverageRate_Del(p)
	assert.Equal(t, H3CInputAverageRate(0), H3CInputAverageRate_Get(p))
	assert.Equal(t, 1, vsaVendorCount(t, p, _H3C_VendorID), "only the output-rate VSA should remain")
}
