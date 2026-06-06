package aruba

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

// TestArubaUserRoleStringVSARoundTrip pins the wire contract for the Aruba VSA
// dictionary. No accept enhancer ships Aruba VSAs today, so this round-trip is
// the only guard that Aruba-User-Role encodes under an Aruba (14823)
// Vendor-Specific attribute and decodes the same string back.
func TestArubaUserRoleStringVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	const value = "authenticated-guest"
	require.NoError(t, ArubaUserRole_SetString(p, value))

	assert.Equal(t, 1, vsaVendorCount(t, p, _Aruba_VendorID))
	assert.Equal(t, value, ArubaUserRole_GetString(p))

	ArubaUserRole_Del(p)
	assert.Equal(t, "", ArubaUserRole_GetString(p))
	assert.Equal(t, 0, vsaVendorCount(t, p, _Aruba_VendorID), "the Aruba VSA should be gone after delete")
}

// TestArubaVSAVendorIsolation ensures the Aruba accessors ignore a foreign
// vendor's attribute carried in the same packet.
func TestArubaVSAVendorIsolation(t *testing.T) {
	p := &radius.Packet{}

	foreign, err := radius.NewVendorSpecific(14988, radius.Attribute{1, 4, 0xAA, 0xBB})
	require.NoError(t, err)
	p.Add(rfc2865.VendorSpecific_Type, foreign)

	assert.Equal(t, 0, vsaVendorCount(t, p, _Aruba_VendorID))
	assert.Equal(t, "", ArubaUserRole_GetString(p), "a foreign VSA must not be read as an Aruba attribute")
}
