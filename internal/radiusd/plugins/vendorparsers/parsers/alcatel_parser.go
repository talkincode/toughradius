package parsers

import (
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/alcatel"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// AlcatelParser parses Alcatel request attributes.
//
// Note: dictionary support is not parse support. Alcatel VSAs only affect
// VendorRequest when this parser is registered and selected by NAS vendor code.
type AlcatelParser struct{}

func (p *AlcatelParser) VendorCode() string {
	return vendors.CodeAlcatel
}

func (p *AlcatelParser) VendorName() string {
	return "Alcatel"
}

func (p *AlcatelParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
	vr := &vendorparsers.VendorRequest{}

	// Alcatel request-side MAC is carried in AAT-User-MAC-Address (type 132). If
	// absent, keep compatibility with the default parser behavior.
	mac := strings.TrimSpace(alcatel.AATUserMACAddress_GetString(r.Packet))
	if mac == "" {
		mac = strings.TrimSpace(rfc2865.CallingStationID_GetString(r.Packet))
	}
	vr.MacAddr = normalizeMACAddress(mac)

	// Alcatel request-side VLAN falls back to the shared NAS-Port-Id parser.
	nasPortID := rfc2869.NASPortID_GetString(r.Packet)
	vr.Vlanid1, vr.Vlanid2 = vendorparsers.ParseVlanIDs(nasPortID)

	return vr, nil
}
