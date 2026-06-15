package parsers

import (
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/radback"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// RadbackParser handles Radback vendor attributes.
//
// Note: dictionary support is not parse support. Radback VSAs only affect
// VendorRequest when this parser is registered and selected by NAS vendor code.
type RadbackParser struct{}

func (p *RadbackParser) VendorCode() string {
	return vendors.CodeRadback
}

func (p *RadbackParser) VendorName() string {
	return "Radback"
}

func (p *RadbackParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
	vr := &vendorparsers.VendorRequest{}

	// Radback request-side MAC is carried in VSA Mac-Addr (type 145). If absent,
	// keep compatibility with the default parser and read Calling-Station-Id.
	mac := strings.TrimSpace(radback.MacAddr_GetString(r.Packet))
	if mac == "" {
		mac = strings.TrimSpace(rfc2865.CallingStationID_GetString(r.Packet))
	}
	vr.MacAddr = normalizeMACAddress(mac)

	// Bind-Dot1q-Vlan-Tag-Id (type 54) is the Radback request-side VLAN source.
	// Fallback to shared NAS-Port-Id parsing when the VSA is not present.
	vr.Vlanid1 = int64(radback.BindDot1qVlanTagID_Get(r.Packet))
	if vr.Vlanid1 == 0 {
		nasPortID := rfc2869.NASPortID_GetString(r.Packet)
		vr.Vlanid1, vr.Vlanid2 = vendorparsers.ParseVlanIDs(nasPortID)
	}

	return vr, nil
}

func normalizeMACAddress(raw string) string {
	mac := strings.ReplaceAll(raw, "-", ":")
	if strings.Contains(mac, ":") || len(mac) != 12 {
		return mac
	}

	return strings.Join([]string{
		mac[0:2],
		mac[2:4],
		mac[4:6],
		mac[6:8],
		mac[8:10],
		mac[10:12],
	}, ":")
}
