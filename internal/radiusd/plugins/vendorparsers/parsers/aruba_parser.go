package parsers

import (
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/aruba"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// ArubaParser parses Aruba request attributes.
//
// Note: dictionary support is not parse support. Aruba VSAs only affect
// VendorRequest when this parser is registered and selected by NAS vendor code.
type ArubaParser struct{}

func (p *ArubaParser) VendorCode() string {
	return vendors.CodeAruba
}

func (p *ArubaParser) VendorName() string {
	return "Aruba"
}

func (p *ArubaParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
	vr := &vendorparsers.VendorRequest{}

	// Aruba request-side MAC usually remains in standard Calling-Station-Id.
	mac := strings.TrimSpace(rfc2865.CallingStationID_GetString(r.Packet))
	vr.MacAddr = normalizeMACAddress(mac)

	// Aruba request-side VLAN uses Aruba-User-Vlan (type 2). When absent, keep
	// compatibility with shared NAS-Port-Id parsing.
	vr.Vlanid1 = int64(aruba.ArubaUserVlan_Get(r.Packet))
	if vr.Vlanid1 == 0 {
		nasPortID := rfc2869.NASPortID_GetString(r.Packet)
		vr.Vlanid1, vr.Vlanid2 = vendorparsers.ParseVlanIDs(nasPortID)
	}

	return vr, nil
}
