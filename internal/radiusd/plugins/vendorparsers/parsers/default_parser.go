package parsers

import (
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// DefaultParser 默认厂商属性解析器
type DefaultParser struct{}

func (p *DefaultParser) VendorCode() string {
	return "default"
}

func (p *DefaultParser) VendorName() string {
	return "Standard"
}

func (p *DefaultParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
	vr := &vendorparsers.VendorRequest{}

	// 解析MAC地址
	macval := rfc2865.CallingStationID_GetString(r.Packet)
	if macval != "" {
		vr.MacAddr = strings.ReplaceAll(macval, "-", ":")
	}

	// 默认解析器不解析VLAN
	vr.Vlanid1 = 0
	vr.Vlanid2 = 0

	return vr, nil
}
