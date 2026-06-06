package pfSense

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// vsaVendorCount returns how many Vendor-Specific attributes on p carry the
// given IANA enterprise number. The generated codec packs each VSA under its
// own RFC 2865 attribute-26 envelope, so this counts only this dictionary's
// attributes and ignores foreign vendors.
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

// TestPfSenseBandwidthMaxUpIntegerVSARoundTrip pins the wire contract for the
// pfSense VSA dictionary. No accept enhancer ships pfSense VSAs today, so this
// round-trip is the only guard that the 32-bit Bandwidth-Max-Up attribute
// encodes under a pfSense (13644) Vendor-Specific attribute and decodes the
// same integer value back.
func TestPfSenseBandwidthMaxUpIntegerVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	const value PfSenseBandwidthMaxUp = 2048
	require.NoError(t, PfSenseBandwidthMaxUp_Set(p, value))

	assert.Equal(t, 1, vsaVendorCount(t, p, _PfSense_VendorID))
	assert.Equal(t, value, PfSenseBandwidthMaxUp_Get(p))

	PfSenseBandwidthMaxUp_Del(p)
	assert.Equal(t, PfSenseBandwidthMaxUp(0), PfSenseBandwidthMaxUp_Get(p))
	assert.Equal(t, 0, vsaVendorCount(t, p, _PfSense_VendorID), "the pfSense VSA should be gone after delete")
}

// TestPfSenseVSAVendorIsolation ensures the pfSense accessors ignore a foreign
// vendor's attribute carried in the same packet.
func TestPfSenseVSAVendorIsolation(t *testing.T) {
	p := &radius.Packet{}

	foreign, err := radius.NewVendorSpecific(14988, radius.Attribute{1, 4, 0xAA, 0xBB})
	require.NoError(t, err)
	p.Add(rfc2865.VendorSpecific_Type, foreign)

	assert.Equal(t, 0, vsaVendorCount(t, p, _PfSense_VendorID))
	assert.Equal(t, PfSenseBandwidthMaxUp(0), PfSenseBandwidthMaxUp_Get(p), "a foreign VSA must not be read as a pfSense attribute")
}
