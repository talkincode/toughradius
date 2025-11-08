package parsers

import (
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// ZTEParser 中兴厂商属性解析器
type ZTEParser struct{}

func (p *ZTEParser) VendorCode() string {
	return "3902"
}

func (p *ZTEParser) VendorName() string {
	return "ZTE"
}

func (p *ZTEParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
	vr := &vendorparsers.VendorRequest{}

	// 解析 MAC 地址 - ZTE 设备的 MAC 地址格式为 12 位连续字符
	macval := rfc2865.CallingStationID_GetString(r.Packet)
	if macval != "" {
		if len(macval) >= 12 {
			// 将 12 位连续字符转换为标准格式
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

	// VLAN 解析
	nasportid := rfc2869.NASPortID_GetString(r.Packet)
	if nasportid == "" {
		vr.Vlanid1 = 0
		vr.Vlanid2 = 0
	}

	return vr, nil
}
