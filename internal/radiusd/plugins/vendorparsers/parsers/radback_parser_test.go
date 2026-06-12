package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/radback"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

func TestRadbackParser_VendorCode(t *testing.T) {
	parser := &RadbackParser{}
	assert.Equal(t, vendors.CodeRadback, parser.VendorCode())
}

func TestRadbackParser_VendorName(t *testing.T) {
	parser := &RadbackParser{}
	assert.Equal(t, "Radback", parser.VendorName())
}

func TestRadbackParser_Parse(t *testing.T) {
	parser := &RadbackParser{}

	tests := []struct {
		name           string
		radbackMAC     string
		callingStation string
		vlanID         uint32
		nasPortID      string
		expectedMAC    string
		expectedVlan1  int64
		expectedVlan2  int64
	}{
		{
			name:          "use radback mac and vlan",
			radbackMAC:    "001122334455",
			vlanID:        777,
			expectedMAC:   "00:11:22:33:44:55",
			expectedVlan1: 777,
			expectedVlan2: 0,
		},
		{
			name:           "fallback to calling station id",
			callingStation: "aa-bb-cc-dd-ee-ff",
			expectedMAC:    "aa:bb:cc:dd:ee:ff",
			expectedVlan1:  0,
			expectedVlan2:  0,
		},
		{
			name:          "fallback vlan parse from nas-port-id",
			radbackMAC:    "00-11-22-33-44-55",
			nasPortID:     "3/0/1:2814.727",
			expectedMAC:   "00:11:22:33:44:55",
			expectedVlan1: 2814,
			expectedVlan2: 727,
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
			if tt.radbackMAC != "" {
				err := radback.MacAddr_SetString(packet, tt.radbackMAC)
				require.NoError(t, err)
			}
			if tt.callingStation != "" {
				err := rfc2865.CallingStationID_SetString(packet, tt.callingStation)
				require.NoError(t, err)
			}
			if tt.vlanID > 0 {
				err := radback.BindDot1qVlanTagID_Set(packet, radback.BindDot1qVlanTagID(tt.vlanID))
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

func TestNormalizeMACAddress(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{name: "empty", in: "", out: ""},
		{name: "with dashes", in: "AA-BB-CC-DD-EE-FF", out: "AA:BB:CC:DD:EE:FF"},
		{name: "with colons", in: "AA:BB:CC:DD:EE:FF", out: "AA:BB:CC:DD:EE:FF"},
		{name: "twelve hex chars", in: "aabbccddeeff", out: "aa:bb:cc:dd:ee:ff"},
		{name: "other format", in: "abc", out: "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, normalizeMACAddress(tt.in))
		})
	}
}
