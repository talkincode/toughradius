package parsers

import (
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// ZTEParser parses ZTE vendor attributes
type ZTEParser struct{}

func (p *ZTEParser) VendorCode() string {
	return "3902"
}

func (p *ZTEParser) VendorName() string {
	return "ZTE"
}

func (p *ZTEParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
	vr := &vendorparsers.VendorRequest{}

		// Parse MAC addresses; ZTE devices provide 12-digit strings
	macval := rfc2865.CallingStationID_GetString(r.Packet)
	if macval != "" {
		if len(macval) >= 12 {
			// Convert the 12-digit string to the standard format
			vr.MacAddr = fmt.Sprintf("%s:%s:%s:%s:%s:%s",
				macval[0:2], macval[2:4], macval[4:6],
				macval[6:8], macval[8:10], macval[10:12])
		} else {
			zap.L().Warn("ZTE CallingStationID length < 12",
				zap.String("namespace", "radius"),
				zap.String("mac", macval))
		}
	} else {
		zap.L().Warn("ZTE CallingStationID is empty", zap.String("namespace", "radius"))
	}

	// VLAN Parse
	nasportid := rfc2869.NASPortID_GetString(r.Packet)
	if nasportid == "" {
		vr.Vlanid1 = 0
		vr.Vlanid2 = 0
	}

	return vr, nil
}
