package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestDefaultParser_VendorCode(t *testing.T) {
	parser := &DefaultParser{}
	assert.Equal(t, "default", parser.VendorCode())
}

func TestDefaultParser_VendorName(t *testing.T) {
	parser := &DefaultParser{}
	assert.Equal(t, "Standard", parser.VendorName())
}

func TestDefaultParser_Parse(t *testing.T) {
	parser := &DefaultParser{}

	tests := []struct {
		name           string
		callingStation string
		expectedMac    string
	}{
		{
			name:           "mac with colons",
			callingStation: "00:11:22:33:44:55",
			expectedMac:    "00:11:22:33:44:55",
		},
		{
			name:           "mac with dashes",
			callingStation: "00-11-22-33-44-55",
			expectedMac:    "00:11:22:33:44:55",
		},
		{
			name:           "empty calling station",
			callingStation: "",
			expectedMac:    "",
		},
		{
			name:           "mixed format",
			callingStation: "00-11-22:33-44:55",
			expectedMac:    "00:11:22:33:44:55",
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
			assert.Equal(t, int64(0), vr.Vlanid1)
			assert.Equal(t, int64(0), vr.Vlanid2)
		})
	}
}
