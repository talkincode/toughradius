package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/aruba"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

func TestArubaParser_VendorCode(t *testing.T) {
	parser := &ArubaParser{}
	assert.Equal(t, vendors.CodeAruba, parser.VendorCode())
}

func TestArubaParser_VendorName(t *testing.T) {
	parser := &ArubaParser{}
	assert.Equal(t, "Aruba", parser.VendorName())
}

func TestArubaParser_Parse(t *testing.T) {
	parser := &ArubaParser{}

	tests := []struct {
		name           string
		callingStation string
		arubaVLAN      uint32
		nasPortID      string
		expectedMAC    string
		expectedVlan1  int64
		expectedVlan2  int64
	}{
		{
			name:           "use aruba user vlan and normalize mac",
			callingStation: "001122334455",
			arubaVLAN:      120,
			expectedMAC:    "00:11:22:33:44:55",
			expectedVlan1:  120,
			expectedVlan2:  0,
		},
		{
			name:           "fallback vlan parse from nas-port-id",
			callingStation: "aa-bb-cc-dd-ee-ff",
			nasPortID:      "3/0/1:2814.727",
			expectedMAC:    "aa:bb:cc:dd:ee:ff",
			expectedVlan1:  2814,
			expectedVlan2:  727,
		},
		{
			name:           "prefer aruba user vlan over nas-port-id",
			callingStation: "11-22-33-44-55-66",
			arubaVLAN:      100,
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
			if tt.arubaVLAN > 0 {
				err := aruba.ArubaUserVlan_Set(packet, aruba.ArubaUserVlan(tt.arubaVLAN))
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
