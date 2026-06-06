package microsoft

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// vsaVendorCount returns how many Vendor-Specific attributes on p carry the
// given IANA enterprise number. The MSCHAPv2 validator/handler read and write
// these as Microsoft (311) VSAs, so the codec must encode them under attribute
// 26 with the correct vendor ID.
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

// TestMicrosoftMSCHAPOctetVSARoundTrip pins the wire contract for the octet
// MSCHAP attributes the validator reads off requests: encode under a Microsoft
// (311) VSA and decode the same bytes back. These carry the challenge/response
// the validator depends on, so a codec regression would break MSCHAPv2 auth.
func TestMicrosoftMSCHAPOctetVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	challenge := []byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17}
	response := bytes.Repeat([]byte{0xAB}, 50)

	require.NoError(t, MSCHAPChallenge_Set(p, challenge))
	require.NoError(t, MSCHAP2Response_Set(p, response))

	assert.Equal(t, 2, vsaVendorCount(t, p, _Microsoft_VendorID))
	assert.Equal(t, challenge, MSCHAPChallenge_Get(p))
	assert.Equal(t, response, MSCHAP2Response_Get(p))

	MSCHAPChallenge_Del(p)
	assert.Nil(t, MSCHAPChallenge_Get(p))
	assert.Equal(t, 1, vsaVendorCount(t, p, _Microsoft_VendorID), "only the MSCHAP2-Response VSA should remain")
}
