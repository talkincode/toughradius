package toughradius

import (
	"fmt"
	"math"
	"net"
	"time"

	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/models"
	"github.com/talkincode/toughradius/v8/toughradius/vendors/h3c"
	"github.com/talkincode/toughradius/v8/toughradius/vendors/huawei"
	"github.com/talkincode/toughradius/v8/toughradius/vendors/ikuai"
	"github.com/talkincode/toughradius/v8/toughradius/vendors/mikrotik"
	"github.com/talkincode/toughradius/v8/toughradius/vendors/zte"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2868"
	"layeh.com/radius/rfc2869"
)

// AcceptAcceptConfig 用户属性策略下发配置
func (s *AuthService) AcceptAcceptConfig(user *models.RadiusUser, vendorCode string, radAccept *radius.Packet) {
	configDefaultAccept(s, user, radAccept)
	switch vendorCode {
	case VendorHuawei:
		configHuaweiAccept(user, radAccept)
	case VendorH3c:
		configH3cAccept(user, radAccept)
	case VendorZte:
		configZteAccept(user, radAccept)
	case VendorMikrotik:
		configMikroTikAccept(user, radAccept)
	case VendorIkuai:
		configIkuaiAccept(user, radAccept)
	}
}

// 设置标准 RADIUS 属性
func configDefaultAccept(s *AuthService, user *models.RadiusUser, radAccept *radius.Packet) {
	var timeout = int64(user.ExpireTime.Sub(time.Now()).Seconds())
	if timeout > math.MaxInt32 {
		timeout = math.MaxInt32
	}
	var interimTimes = s.GetIntConfig(app.ConfigRadiusAcctInterimInterval, 120)
	rfc2865.SessionTimeout_Set(radAccept, rfc2865.SessionTimeout(timeout))
	rfc2869.AcctInterimInterval_Set(radAccept, rfc2869.AcctInterimInterval(interimTimes))

	if common.IsNotEmptyAndNA(user.AddrPool) {
		rfc2869.FramedPool_SetString(radAccept, user.AddrPool)
	}

	if common.IsNotEmptyAndNA(user.IpAddr) {
		rfc2865.FramedIPAddress_Set(radAccept, net.ParseIP(user.IpAddr))
	}

	// Configure VLAN assignment using tunnel attributes
	// Use primary VLAN ID if configured
	if user.Vlanid1 > 0 {
		configVlanTunnelAttributes(radAccept, user.Vlanid1)
	}
	// Use secondary VLAN ID as fallback if primary is not set
	if user.Vlanid1 == 0 && user.Vlanid2 > 0 {
		configVlanTunnelAttributes(radAccept, user.Vlanid2)
	}
}

func configMikroTikAccept(user *models.RadiusUser, radAccept *radius.Packet) {
	mikrotik.MikrotikRateLimit_SetString(radAccept, fmt.Sprintf("%dk/%dk", user.UpRate, user.DownRate))
}

func configIkuaiAccept(user *models.RadiusUser, radAccept *radius.Packet) {
	var up = int64(user.UpRate) * 1024 * 8
	var down = int64(user.DownRate) * 1024 * 8
	if up > math.MaxInt32 {
		up = math.MaxInt32
	}
	if down > math.MaxInt32 {
		down = math.MaxInt32
	}

	ikuai.RPUpstreamSpeedLimit_Set(radAccept, ikuai.RPUpstreamSpeedLimit(up))
	ikuai.RPDownstreamSpeedLimit_Set(radAccept, ikuai.RPDownstreamSpeedLimit(down))
}

func configHuaweiAccept(user *models.RadiusUser, radAccept *radius.Packet) {
	var up = int64(user.UpRate) * 1024
	var down = int64(user.DownRate) * 1024
	var upPeak = up * 4
	var downPeak = down * 4
	if up > math.MaxInt32 {
		up = math.MaxInt32
	}
	if upPeak > math.MaxInt32 {
		upPeak = math.MaxInt32
	}
	if down > math.MaxInt32 {
		down = math.MaxInt32
	}
	if downPeak > math.MaxInt32 {
		downPeak = math.MaxInt32
	}
	huawei.HuaweiInputAverageRate_Set(radAccept, huawei.HuaweiInputAverageRate(up))
	huawei.HuaweiInputPeakRate_Set(radAccept, huawei.HuaweiInputPeakRate(upPeak))
	huawei.HuaweiOutputAverageRate_Set(radAccept, huawei.HuaweiOutputAverageRate(down))
	huawei.HuaweiOutputPeakRate_Set(radAccept, huawei.HuaweiOutputPeakRate(downPeak))

}

func configH3cAccept(user *models.RadiusUser, radAccept *radius.Packet) {
	var up = int64(user.UpRate) * 1024
	var down = int64(user.DownRate) * 1024
	var upPeak = up * 4
	var downPeak = down * 4
	if up > math.MaxInt32 {
		up = math.MaxInt32
	}
	if upPeak > math.MaxInt32 {
		upPeak = math.MaxInt32
	}
	if down > math.MaxInt32 {
		down = math.MaxInt32
	}
	if downPeak > math.MaxInt32 {
		downPeak = math.MaxInt32
	}
	h3c.H3CInputAverageRate_Set(radAccept, h3c.H3CInputAverageRate(up))
	h3c.H3CInputPeakRate_Set(radAccept, h3c.H3CInputPeakRate(upPeak))
	h3c.H3COutputAverageRate_Set(radAccept, h3c.H3COutputAverageRate(down))
	h3c.H3COutputPeakRate_Set(radAccept, h3c.H3COutputPeakRate(downPeak))
}

func configZteAccept(user *models.RadiusUser, radAccept *radius.Packet) {
	var up = int64(user.UpRate) * 1024
	var down = int64(user.DownRate) * 1024
	if up > math.MaxInt32 {
		up = math.MaxInt32
	}
	if down > math.MaxInt32 {
		down = math.MaxInt32
	}
	zte.ZTERateCtrlSCRUp_Set(radAccept, zte.ZTERateCtrlSCRUp(up))
	zte.ZTERateCtrlSCRDown_Set(radAccept, zte.ZTERateCtrlSCRDown(down))
}

// configVlanTunnelAttributes 配置VLAN隧道属性
// 使用RFC 2868隧道属性为交换机下发VLAN ID
func configVlanTunnelAttributes(radAccept *radius.Packet, vlanId int) {
	// Tag 0 is typically used for single tunnel
	tag := byte(0)
	
	// Tunnel-Type = VLAN (13) - RFC 2868 standard value for VLAN
	rfc2868.TunnelType_Set(radAccept, tag, rfc2868.TunnelType(13))
	
	// Tunnel-Medium-Type = IEEE-802 (6) for Ethernet
	rfc2868.TunnelMediumType_Set(radAccept, tag, rfc2868.TunnelMediumType(6))
	
	// Tunnel-Private-Group-ID = VLAN ID (as string)
	rfc2868.TunnelPrivateGroupID_SetString(radAccept, tag, fmt.Sprintf("%d", vlanId))
}
