package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/alcatel"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

func TestAlcatelParser_VendorCode(t *testing.T) {
	parser := &AlcatelParser{}
	assert.Equal(t, vendors.CodeAlcatel, parser.VendorCode())
}

func TestAlcatelParser_VendorName(t *testing.T) {
	parser := &AlcatelParser{}
	assert.Equal(t, "Alcatel", parser.VendorName())
}

func TestAlcatelParser_Parse(t *testing.T) {
	parser := &AlcatelParser{}

	tests := []struct {
		name           string
		alcatelMAC     string
		callingStation string
		nasPortID      string
		expectedMAC    string
		expectedVlan1  int64
		expectedVlan2  int64
	}{
		{
			name:          "use alcatel mac and parse vlan from nas-port-id",
			alcatelMAC:    "001122334455",
			nasPortID:     "3/0/1:2814.727",
			expectedMAC:   "00:11:22:33:44:55",
			expectedVlan1: 2814,
			expectedVlan2: 727,
		},
		{
			name:           "fallback to calling-station-id",
			callingStation: "aa-bb-cc-dd-ee-ff",
			expectedMAC:    "aa:bb:cc:dd:ee:ff",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:           "prefer alcatel mac over calling-station-id",
			alcatelMAC:     "11-22-33-44-55-66",
			callingStation: "00-00-00-00-00-00",
			expectedMAC:    "11:22:33:44:55:66",
			expectedVlan1:  0,
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
			if tt.alcatelMAC != "" {
				err := alcatel.AATUserMACAddress_SetString(packet, tt.alcatelMAC)
				require.NoError(t, err)
			}
			if tt.callingStation != "" {
				err := rfc2865.CallingStationID_SetString(packet, tt.callingStation)
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
