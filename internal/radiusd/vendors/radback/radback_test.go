package radback

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// vsaVendorCount returns how many Vendor-Specific attributes on p carry the
// given IANA enterprise number. The codec packs each VSA under its own
// RFC 2865 attribute-26 envelope, so this counts only this dictionary's
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

// TestRedbackMacAddrStringVSARoundTrip pins the wire contract for the pruned
// Redback VSA dictionary: Mac-Addr (145) encodes under a Redback (2352)
// Vendor-Specific attribute and decodes the same string back.
func TestRedbackMacAddrStringVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	const value = "00:11:22:33:44:55"
	require.NoError(t, MacAddr_SetString(p, value))

	assert.Equal(t, 1, vsaVendorCount(t, p, _Redback_VendorID))
	assert.Equal(t, value, MacAddr_GetString(p))

	MacAddr_Del(p)
	assert.Equal(t, "", MacAddr_GetString(p))
	assert.Equal(t, 0, vsaVendorCount(t, p, _Redback_VendorID), "the Redback VSA should be gone after delete")
}

// TestRedbackBindDot1qVlanTagIDIntegerVSARoundTrip pins the wire contract for
// Bind-Dot1q-Vlan-Tag-Id (54): it encodes as a Redback integer VSA and decodes
// the same value back.
func TestRedbackBindDot1qVlanTagIDIntegerVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	require.NoError(t, BindDot1qVlanTagID_Set(p, 100))

	assert.Equal(t, 1, vsaVendorCount(t, p, _Redback_VendorID))
	assert.Equal(t, BindDot1qVlanTagID(100), BindDot1qVlanTagID_Get(p))

	value, err := BindDot1qVlanTagID_Lookup(p)
	require.NoError(t, err)
	assert.Equal(t, BindDot1qVlanTagID(100), value)
}

// TestRedbackVSAVendorIsolation ensures the Redback accessors ignore a foreign
// vendor's attribute carried in the same packet.
func TestRedbackVSAVendorIsolation(t *testing.T) {
	p := &radius.Packet{}

	foreign, err := radius.NewVendorSpecific(14988, radius.Attribute{1, 4, 0xAA, 0xBB})
	require.NoError(t, err)
	p.Add(rfc2865.VendorSpecific_Type, foreign)

	assert.Equal(t, 0, vsaVendorCount(t, p, _Redback_VendorID))
	assert.Equal(t, "", MacAddr_GetString(p), "a foreign VSA must not be read as a Redback attribute")
	assert.Equal(t, BindDot1qVlanTagID(0), BindDot1qVlanTagID_Get(p))

	_, lookupErr := BindDot1qVlanTagID_Lookup(p)
	assert.ErrorIs(t, lookupErr, radius.ErrNoAttribute)
}
