package unix

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

// TestUnixFTPHomeStringVSARoundTrip pins the wire contract for the Unix VSA
// dictionary. No accept enhancer ships Unix VSAs today, so this round-trip is
// the only guard that Unix-FTP-Home encodes under a Unix (4) Vendor-Specific
// attribute and decodes the same string back.
func TestUnixFTPHomeStringVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	const value = "/home/ftp/user1"
	require.NoError(t, UnixFTPHome_SetString(p, value))

	assert.Equal(t, 1, vsaVendorCount(t, p, _Unix_VendorID))
	assert.Equal(t, value, UnixFTPHome_GetString(p))

	UnixFTPHome_Del(p)
	assert.Equal(t, "", UnixFTPHome_GetString(p))
	assert.Equal(t, 0, vsaVendorCount(t, p, _Unix_VendorID), "the Unix VSA should be gone after delete")
}

// TestUnixVSAVendorIsolation ensures the Unix accessors ignore a foreign
// vendor's attribute carried in the same packet.
func TestUnixVSAVendorIsolation(t *testing.T) {
	p := &radius.Packet{}

	foreign, err := radius.NewVendorSpecific(14988, radius.Attribute{1, 4, 0xAA, 0xBB})
	require.NoError(t, err)
	p.Add(rfc2865.VendorSpecific_Type, foreign)

	assert.Equal(t, 0, vsaVendorCount(t, p, _Unix_VendorID))
	assert.Equal(t, "", UnixFTPHome_GetString(p), "a foreign VSA must not be read as a Unix attribute")
}
