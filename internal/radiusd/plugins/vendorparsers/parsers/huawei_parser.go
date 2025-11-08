package parsers

import (
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// HuaweiParser 华为厂商属性解析器
type HuaweiParser struct{}

func (p *HuaweiParser) VendorCode() string {
	return "2011"
}

func (p *HuaweiParser) VendorName() string {
	return "Huawei"
}

func (p *HuaweiParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
	vr := &vendorparsers.VendorRequest{}

	// 解析 MAC 地址 - 华为设备优先使用 CallingStationID
	macval := rfc2865.CallingStationID_GetString(r.Packet)
	if macval != "" {
		vr.MacAddr = strings.ReplaceAll(macval, "-", ":")
	}

	// 华为设备的 VLAN 解析可以从 NAS-Port-Id 或其他私有属性获取
	// 这里保持简单，使用默认值
	vr.Vlanid1 = 0
	vr.Vlanid2 = 0

	return vr, nil
}
