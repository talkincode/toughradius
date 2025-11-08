package parsers

import (
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/h3c"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// H3CParser H3C 厂商属性解析器
type H3CParser struct{}

func (p *H3CParser) VendorCode() string {
	return "25506"
}

func (p *H3CParser) VendorName() string {
	return "H3C"
}

func (p *H3CParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
	vr := &vendorparsers.VendorRequest{}

	// 解析 MAC 地址 - H3C 使用 H3C-IP-Host-Addr
	ipha := h3c.H3CIPHostAddr_GetString(r.Packet)
	if ipha != "" {
		iphalen := len(ipha)
		if iphalen > 17 {
			vr.MacAddr = ipha[iphalen-17:]
		} else {
			vr.MacAddr = ipha
		}
	} else {
		// 备用方案：使用标准 CallingStationID
		macval := rfc2865.CallingStationID_GetString(r.Packet)
		if macval != "" {
			vr.MacAddr = strings.ReplaceAll(macval, "-", ":")
		} else {
			zap.L().Warn("h3c.H3CIPHostAddr and CallingStationID are empty", zap.String("namespace", "radius"))
		}
	}

	// H3C 的 VLAN 解析
	nasportid := rfc2869.NASPortID_GetString(r.Packet)
	if nasportid == "" {
		vr.Vlanid1 = 0
		vr.Vlanid2 = 0
	}

	return vr, nil
}
