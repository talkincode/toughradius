package parsers

import (
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// HuaweiParser handles Huawei vendor attributes.
//
// Note: registering a vendor (or shipping its attribute dictionary) does not by
// itself parse anything — dictionary support is not parse support. A field on
// VendorRequest is only populated when this parser explicitly extracts it.
type HuaweiParser struct{}

func (p *HuaweiParser) VendorCode() string {
	return vendors.CodeHuawei
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

	// Huawei encodes VLAN IDs in the NAS-Port-Id attribute. Parse them with the
	// shared standard parser instead of leaving them stubbed at zero.
	nasPortID := rfc2869.NASPortID_GetString(r.Packet)
	vr.Vlanid1, vr.Vlanid2 = vendorparsers.ParseVlanIDs(nasPortID)

	return vr, nil
}
