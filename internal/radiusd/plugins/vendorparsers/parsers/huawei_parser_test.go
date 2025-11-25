package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestHuaweiParser_VendorCode(t *testing.T) {
	parser := &HuaweiParser{}
	assert.Equal(t, vendors.CodeHuawei, parser.VendorCode())
}

func TestHuaweiParser_VendorName(t *testing.T) {
	parser := &HuaweiParser{}
	assert.Equal(t, "Huawei", parser.VendorName())
}

func TestHuaweiParser_Parse(t *testing.T) {
	parser := &HuaweiParser{}

	tests := []struct {
		name           string
		callingStation string
		expectedMac    string
		expectedVlan1  int64
		expectedVlan2  int64
	}{
		{
			name:           "mac with colons",
			callingStation: "00:11:22:33:44:55",
			expectedMac:    "00:11:22:33:44:55",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "mac with dashes",
			callingStation: "00-11-22-33-44-55",
			expectedMac:    "00:11:22:33:44:55",
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
			name:           "mac with mixed format",
			callingStation: "aa-bb-cc:dd-ee:ff",
			expectedMac:    "aa:bb:cc:dd:ee:ff",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "uppercase mac",
			callingStation: "AA-BB-CC-DD-EE-FF",
			expectedMac:    "AA:BB:CC:DD:EE:FF",
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

func TestHuaweiParser_Parse_NoAttributes(t *testing.T) {
	parser := &HuaweiParser{}

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	req := &radius.Request{Packet: packet}

	vr, err := parser.Parse(req)
	require.NoError(t, err)
	require.NotNil(t, vr)

	// When no attributes are present, should return an empty VendorRequest
	assert.Equal(t, "", vr.MacAddr)
	assert.Equal(t, int64(0), vr.Vlanid1)
	assert.Equal(t, int64(0), vr.Vlanid2)
}
