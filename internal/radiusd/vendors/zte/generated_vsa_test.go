package zte

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// vsaVendorCount returns how many Vendor-Specific attributes on p carry the
// given IANA enterprise number. The accept enhancer ships ZTE rate limits as
// VSAs, so the codec must encode them under attribute 26 with vendor 3902.
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

// TestZTERateVSARoundTrip pins the wire contract for the ZTE rate-control
// attributes the accept enhancer ships: encode under ZTE (3902) VSAs and decode
// the same values back.
func TestZTERateVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	require.NoError(t, ZTERateCtrlSCRUp_Set(p, ZTERateCtrlSCRUp(512)))
	require.NoError(t, ZTERateCtrlSCRDown_Set(p, ZTERateCtrlSCRDown(4096)))

	assert.Equal(t, 2, vsaVendorCount(t, p, _ZTE_VendorID))
	assert.Equal(t, ZTERateCtrlSCRUp(512), ZTERateCtrlSCRUp_Get(p))
	assert.Equal(t, ZTERateCtrlSCRDown(4096), ZTERateCtrlSCRDown_Get(p))

	gets, err := ZTERateCtrlSCRDown_Gets(p)
	require.NoError(t, err)
	assert.Equal(t, []ZTERateCtrlSCRDown{4096}, gets)

	ZTERateCtrlSCRUp_Del(p)
	assert.Equal(t, ZTERateCtrlSCRUp(0), ZTERateCtrlSCRUp_Get(p))
	assert.Equal(t, 1, vsaVendorCount(t, p, _ZTE_VendorID), "only the down-rate VSA should remain")
}
