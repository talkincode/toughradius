package huawei

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// countVSAForVendor decodes every Vendor-Specific attribute on p and returns how
// many carry wantVendorID. It is the on-wire view a NAS sees: the generated
// codec must pack vendor attributes under RFC 2865 attribute 26 with the
// vendor's IANA enterprise number, or the device silently ignores them.
func countVSAForVendor(t *testing.T, p *radius.Packet, wantVendorID uint32) int {
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

// TestHuaweiRateVSARoundTrip exercises the generated Huawei VSA codec end to end
// for the bandwidth attribute the accept enhancer actually ships to Huawei NAS
// devices: Set must encode it as a Huawei (2011) Vendor-Specific attribute, and
// Get/Gets/Lookup must decode the same value back.
func TestHuaweiRateVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	require.NoError(t, HuaweiInputAverageRate_Set(p, HuaweiInputAverageRate(2048)))

	require.Equal(t, 1, countVSAForVendor(t, p, _Huawei_VendorID),
		"value must be encoded as exactly one Huawei Vendor-Specific attribute")
	assert.Equal(t, HuaweiInputAverageRate(2048), HuaweiInputAverageRate_Get(p))

	got, err := HuaweiInputAverageRate_Lookup(p)
	require.NoError(t, err)
	assert.Equal(t, HuaweiInputAverageRate(2048), got)

	gets, err := HuaweiInputAverageRate_Gets(p)
	require.NoError(t, err)
	assert.Equal(t, []HuaweiInputAverageRate{2048}, gets)

	HuaweiInputAverageRate_Del(p)
	assert.Equal(t, 0, countVSAForVendor(t, p, _Huawei_VendorID))
	assert.Equal(t, HuaweiInputAverageRate(0), HuaweiInputAverageRate_Get(p))
	if _, err := HuaweiInputAverageRate_Lookup(p); err == nil {
		t.Fatal("Lookup after Del must return an error")
	}
}

// TestHuaweiStringAndIPv6VSARoundTrip covers the non-integer value codecs the
// Huawei enhancer relies on: a string domain name and an IPv6 framed address.
func TestHuaweiStringAndIPv6VSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	require.NoError(t, HuaweiDomainName_SetString(p, "corp.example.net"))
	assert.Equal(t, "corp.example.net", HuaweiDomainName_GetString(p))

	ip := net.ParseIP("2001:db8::1")
	require.NotNil(t, ip)
	require.NoError(t, HuaweiFramedIPv6Address_Set(p, ip))
	assert.True(t, HuaweiFramedIPv6Address_Get(p).Equal(ip))

	// Two distinct Huawei attributes must produce two Huawei VSAs.
	assert.Equal(t, 2, countVSAForVendor(t, p, _Huawei_VendorID))
}

// TestHuaweiVSAVendorIsolation verifies the codec ignores Vendor-Specific
// attributes belonging to other vendors: a foreign VSA mixed into the packet
// must not be mistaken for a Huawei attribute, and the Huawei value must still
// decode correctly. This guards the vendorID mismatch branch in the codec.
func TestHuaweiVSAVendorIsolation(t *testing.T) {
	p := &radius.Packet{}

	// Inject a foreign vendor's VSA (Mikrotik enterprise number 14988).
	foreign, err := radius.NewVendorSpecific(14988, radius.Attribute{1, 4, 0xAA, 0xBB})
	require.NoError(t, err)
	p.Add(rfc2865.VendorSpecific_Type, foreign)

	require.NoError(t, HuaweiOutputAverageRate_Set(p, HuaweiOutputAverageRate(4096)))

	assert.Equal(t, HuaweiOutputAverageRate(4096), HuaweiOutputAverageRate_Get(p))
	assert.Equal(t, 1, countVSAForVendor(t, p, _Huawei_VendorID),
		"foreign VSA must not be counted as Huawei")
	assert.Equal(t, 1, countVSAForVendor(t, p, 14988),
		"foreign VSA must be left intact")
}
