package juniper

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

// TestJuniperLocalUserNameStringVSARoundTrip pins the wire contract for the
// Juniper VSA dictionary. No accept enhancer ships Juniper VSAs today, so this
// round-trip is the only guard that Juniper-Local-User-Name encodes under a
// Juniper (2636) Vendor-Specific attribute and decodes the same string back.
func TestJuniperLocalUserNameStringVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	const value = "netops-admin"
	require.NoError(t, JuniperLocalUserName_SetString(p, value))

	assert.Equal(t, 1, vsaVendorCount(t, p, _Juniper_VendorID))
	assert.Equal(t, value, JuniperLocalUserName_GetString(p))

	JuniperLocalUserName_Del(p)
	assert.Equal(t, "", JuniperLocalUserName_GetString(p))
	assert.Equal(t, 0, vsaVendorCount(t, p, _Juniper_VendorID), "the Juniper VSA should be gone after delete")
}

// TestJuniperVSAVendorIsolation ensures the Juniper accessors ignore a foreign
// vendor's attribute carried in the same packet.
func TestJuniperVSAVendorIsolation(t *testing.T) {
	p := &radius.Packet{}

	foreign, err := radius.NewVendorSpecific(14988, radius.Attribute{1, 4, 0xAA, 0xBB})
	require.NoError(t, err)
	p.Add(rfc2865.VendorSpecific_Type, foreign)

	assert.Equal(t, 0, vsaVendorCount(t, p, _Juniper_VendorID))
	assert.Equal(t, "", JuniperLocalUserName_GetString(p), "a foreign VSA must not be read as a Juniper attribute")
}
