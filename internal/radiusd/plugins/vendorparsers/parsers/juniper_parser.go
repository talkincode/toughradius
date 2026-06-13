package parsers

import (
	"strconv"
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/juniper"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// JuniperParser parses Juniper request attributes.
//
// Note: dictionary support is not parse support. Juniper VSAs only affect
// VendorRequest when this parser is registered and selected by NAS vendor code.
type JuniperParser struct{}

func (p *JuniperParser) VendorCode() string {
	return vendors.CodeJuniper
}

func (p *JuniperParser) VendorName() string {
	return "Juniper"
}

func (p *JuniperParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
	vr := &vendorparsers.VendorRequest{}

	// Juniper request-side MAC typically remains in Calling-Station-Id.
	mac := strings.TrimSpace(rfc2865.CallingStationID_GetString(r.Packet))
	vr.MacAddr = normalizeMACAddress(mac)

	// Juniper request-side VLAN can be carried in Juniper-VoIP-Vlan (type 49).
	// If absent or malformed, keep compatibility with shared NAS-Port-Id parsing.
	vlan := strings.TrimSpace(juniper.JuniperVoIPVlan_GetString(r.Packet))
	if vlan != "" {
		if id, err := strconv.ParseInt(vlan, 10, 64); err == nil && id > 0 {
			vr.Vlanid1 = id
			return vr, nil
		}
	}

	nasPortID := rfc2869.NASPortID_GetString(r.Packet)
	vr.Vlanid1, vr.Vlanid2 = vendorparsers.ParseVlanIDs(nasPortID)

	return vr, nil
}
