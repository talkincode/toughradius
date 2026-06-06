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
//
// The lengths match the real on-wire contract enforced by the MSCHAP validator
// (internal/radiusd/plugins/auth/validators/mschap_validator.go): MS-CHAP-
// Challenge is exactly 16 bytes and MS-CHAP2-Response is exactly 50 bytes.
func TestMicrosoftMSCHAPOctetVSARoundTrip(t *testing.T) {
	p := &radius.Packet{}

	challenge := bytes.Repeat([]byte{0x5A}, 16)
	response := bytes.Repeat([]byte{0xAB}, 50)

	require.NoError(t, MSCHAPChallenge_Set(p, challenge))
	require.NoError(t, MSCHAP2Response_Set(p, response))

	assert.Equal(t, 2, vsaVendorCount(t, p, _Microsoft_VendorID))

	gotChallenge := MSCHAPChallenge_Get(p)
	gotResponse := MSCHAP2Response_Get(p)
	assert.Equal(t, challenge, gotChallenge)
	assert.Equal(t, response, gotResponse)

	// Guard the exact lengths the validator requires; a codec that truncated or
	// padded the octet value would break MSCHAPv2 authentication.
	assert.Len(t, gotChallenge, 16)
	assert.Len(t, gotResponse, 50)

	MSCHAPChallenge_Del(p)
	assert.Nil(t, MSCHAPChallenge_Get(p))
	assert.Equal(t, 1, vsaVendorCount(t, p, _Microsoft_VendorID), "only the MSCHAP2-Response VSA should remain")
}
