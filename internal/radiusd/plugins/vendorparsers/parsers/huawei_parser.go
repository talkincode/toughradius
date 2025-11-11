package parsers

import (
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// HuaweiParser handles Huawei vendor attributes
type HuaweiParser struct{}

func (p *HuaweiParser) VendorCode() string {
	return "2011"
}

func (p *HuaweiParser) VendorName() string {
	return "Huawei"
}

func (p *HuaweiParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
	vr := &vendorparsers.VendorRequest{}

	// Parse MAC addresses; Huawei devices prefer CallingStationID
	macval := rfc2865.CallingStationID_GetString(r.Packet)
	if macval != "" {
		vr.MacAddr = strings.ReplaceAll(macval, "-", ":")
	}

	// Huawei devices parse VLANs from NAS-Port-Id or other vendor-specific attributes
	// Keep it simple here by using default values
	vr.Vlanid1 = 0
	vr.Vlanid2 = 0

	return vr, nil
}
