package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestZTEParser_VendorCode(t *testing.T) {
	parser := &ZTEParser{}
	assert.Equal(t, "3902", parser.VendorCode())
}

func TestZTEParser_VendorName(t *testing.T) {
	parser := &ZTEParser{}
	assert.Equal(t, "ZTE", parser.VendorName())
}

func TestZTEParser_Parse(t *testing.T) {
	parser := &ZTEParser{}

	tests := []struct {
		name           string
		callingStation string
		expectedMac    string
		expectedVlan1  int64
		expectedVlan2  int64
	}{
		{
			name:           "12-char continuous MAC",
			callingStation: "aabbccddeeff",
			expectedMac:    "aa:bb:cc:dd:ee:ff",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "12-char uppercase MAC",
			callingStation: "AABBCCDDEEFF",
			expectedMac:    "AA:BB:CC:DD:EE:FF",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "mixed case MAC - preserves case",
			callingStation: "AaBbCcDdEeFf",
			expectedMac:    "Aa:Bb:Cc:Dd:Ee:Ff", // Preserve the original casing
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "MAC already with colons - parses first 12 chars",
			callingStation: "00:11:22:33:44:55",
			expectedMac:    "00::1:1::22::3:3:", // Parse as a 12-digit string (includes colons)
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "MAC with dashes - parses first 12 chars",
			callingStation: "00-11-22-33-44-55",
			expectedMac:    "00:-1:1-:22:-3:3-", // Parse as a 12-digit string (includes hyphens)
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "empty calling station",
			callingStation: "",
			expectedMac:    "",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "short MAC (less than 12 chars) - warning logged",
			callingStation: "aabbccdd",
			expectedMac:    "", // Length insufficient; do not handle
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "13-char MAC - uses first 12 chars",
			callingStation: "aabbccddeeff1",
			expectedMac:    "aa:bb:cc:dd:ee:ff", // Only use the first 12 digits
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "numeric MAC",
			callingStation: "001122334455",
			expectedMac:    "00:11:22:33:44:55",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
			if tt.callingStation != "" {
				rfc2865.CallingStationID_SetString(packet, tt.callingStation)
			}

			req := &radius.Request{Packet: packet}

			vr, err := parser.Parse(req)
			require.NoError(t, err)
			require.NotNil(t, vr)

			assert.Equal(t, tt.expectedMac, vr.MacAddr)
			assert.Equal(t, tt.expectedVlan1, vr.Vlanid1)
			assert.Equal(t, tt.expectedVlan2, vr.Vlanid2)
		})
	}
}

func TestZTEParser_Parse_NoAttributes(t *testing.T) {
	parser := &ZTEParser{}

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	req := &radius.Request{Packet: packet}

	vr, err := parser.Parse(req)
	require.NoError(t, err)
	require.NotNil(t, vr)

	assert.Equal(t, "", vr.MacAddr)
	assert.Equal(t, int64(0), vr.Vlanid1)
	assert.Equal(t, int64(0), vr.Vlanid2)
}
