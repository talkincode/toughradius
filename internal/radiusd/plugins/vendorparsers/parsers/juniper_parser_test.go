package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/juniper"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

func TestJuniperParser_VendorCode(t *testing.T) {
	parser := &JuniperParser{}
	assert.Equal(t, vendors.CodeJuniper, parser.VendorCode())
}

func TestJuniperParser_VendorName(t *testing.T) {
	parser := &JuniperParser{}
	assert.Equal(t, "Juniper", parser.VendorName())
}

func TestJuniperParser_Parse(t *testing.T) {
	parser := &JuniperParser{}

	tests := []struct {
		name           string
		callingStation string
		voipVLAN       string
		nasPortID      string
		expectedMAC    string
		expectedVlan1  int64
		expectedVlan2  int64
	}{
		{
			name:           "use juniper voip vlan and normalize mac",
			callingStation: "001122334455",
			voipVLAN:       "120",
			expectedMAC:    "00:11:22:33:44:55",
			expectedVlan1:  120,
			expectedVlan2:  0,
		},
		{
			name:           "fallback vlan parse from nas-port-id when voip vlan is missing",
			callingStation: "aa-bb-cc-dd-ee-ff",
			nasPortID:      "3/0/1:2814.727",
			expectedMAC:    "aa:bb:cc:dd:ee:ff",
			expectedVlan1:  2814,
			expectedVlan2:  727,
		},
		{
			name:           "fallback vlan parse from nas-port-id when voip vlan is malformed",
			callingStation: "11-22-33-44-55-66",
			voipVLAN:       "voice-200",
			nasPortID:      "3/0/1:400.12",
			expectedMAC:    "11:22:33:44:55:66",
			expectedVlan1:  400,
			expectedVlan2:  12,
		},
		{
			name:           "prefer juniper voip vlan over nas-port-id",
			callingStation: "11-22-33-44-55-66",
			voipVLAN:       "100",
			nasPortID:      "3/0/1:2814.727",
			expectedMAC:    "11:22:33:44:55:66",
			expectedVlan1:  100,
			expectedVlan2:  0,
		},
		{
			name:          "no attributes",
			expectedMAC:   "",
			expectedVlan1: 0,
			expectedVlan2: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := radius.New(radius.CodeAccessRequest, []byte("secret"))

			if tt.callingStation != "" {
				err := rfc2865.CallingStationID_SetString(packet, tt.callingStation)
				require.NoError(t, err)
			}
			if tt.voipVLAN != "" {
				err := juniper.JuniperVoIPVlan_SetString(packet, tt.voipVLAN)
				require.NoError(t, err)
			}
			if tt.nasPortID != "" {
				err := rfc2869.NASPortID_SetString(packet, tt.nasPortID)
				require.NoError(t, err)
			}

			req := &radius.Request{Packet: packet}
			vr, err := parser.Parse(req)
			require.NoError(t, err)
			require.NotNil(t, vr)
			assert.Equal(t, tt.expectedMAC, vr.MacAddr)
			assert.Equal(t, tt.expectedVlan1, vr.Vlanid1)
			assert.Equal(t, tt.expectedVlan2, vr.Vlanid2)
		})
	}
}
