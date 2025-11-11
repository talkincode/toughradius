package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/h3c"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestH3CParser_VendorCode(t *testing.T) {
	parser := &H3CParser{}
	assert.Equal(t, "25506", parser.VendorCode())
}

func TestH3CParser_VendorName(t *testing.T) {
	parser := &H3CParser{}
	assert.Equal(t, "H3C", parser.VendorName())
}

func TestH3CParser_Parse(t *testing.T) {
	parser := &H3CParser{}

	tests := []struct {
		name           string
		ipHostAddr     string
		callingStation string
		expectedMac    string
		expectedVlan1  int64
		expectedVlan2  int64
	}{
		{
			name:           "valid H3C-IP-Host-Addr with MAC",
			ipHostAddr:     "192.168.1.100\x0000:11:22:33:44:55",
			callingStation: "",
			expectedMac:    "00:11:22:33:44:55",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "H3C-IP-Host-Addr too short, fallback to CallingStationID",
			ipHostAddr:     "192.168.1.100",
			callingStation: "aa:bb:cc:dd:ee:ff",
			expectedMac:    "192.168.1.100", // Length shorter than 17, use IP as MAC
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "H3C-IP-Host-Addr empty, fallback to CallingStationID",
			ipHostAddr:     "",
			callingStation: "11:22:33:44:55:66",
			expectedMac:    "11:22:33:44:55:66",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "both empty",
			ipHostAddr:     "",
			callingStation: "",
			expectedMac:    "",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "H3C-IP-Host-Addr exactly 17 chars",
			ipHostAddr:     "AA:BB:CC:DD:EE:FF",
			callingStation: "00:00:00:00:00:00",
			expectedMac:    "AA:BB:CC:DD:EE:FF",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "H3C-IP-Host-Addr with prefix, extract last 17",
			ipHostAddr:     "some_prefix_AA:BB:CC:DD:EE:FF",
			callingStation: "00:00:00:00:00:00",
			expectedMac:    "AA:BB:CC:DD:EE:FF",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "CallingStationID with dashes",
			ipHostAddr:     "",
			callingStation: "aa-bb-cc-dd-ee-ff",
			expectedMac:    "aa:bb:cc:dd:ee:ff",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := radius.New(radius.CodeAccessRequest, []byte("secret"))

			if tt.ipHostAddr != "" {
				// H3C-IP-Host-Addr (Vendor 25506, Type 255)
				err := h3c.H3CIPHostAddr_AddString(packet, tt.ipHostAddr)
				require.NoError(t, err)
			}

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

func TestH3CParser_Parse_NoAttributes(t *testing.T) {
	parser := &H3CParser{}

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	req := &radius.Request{Packet: packet}

	vr, err := parser.Parse(req)
	require.NoError(t, err)
	require.NotNil(t, vr)

	assert.Equal(t, "", vr.MacAddr)
	assert.Equal(t, int64(0), vr.Vlanid1)
	assert.Equal(t, int64(0), vr.Vlanid2)
}
