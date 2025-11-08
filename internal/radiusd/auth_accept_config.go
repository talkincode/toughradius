package radiusd

import (
	"fmt"
	"math"
	"net"
	"time"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/h3c"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/huawei"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/ikuai"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/mikrotik"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/zte"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// AcceptAcceptConfig 用户属性策略下发配置
func (s *AuthService) AcceptAcceptConfig(user *domain.RadiusUser, vendorCode string, radAccept *radius.Packet) {
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
func configDefaultAccept(s *AuthService, user *domain.RadiusUser, radAccept *radius.Packet) {
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
}

func configMikroTikAccept(user *domain.RadiusUser, radAccept *radius.Packet) {
	mikrotik.MikrotikRateLimit_SetString(radAccept, fmt.Sprintf("%dk/%dk", user.UpRate, user.DownRate))
}

func configIkuaiAccept(user *domain.RadiusUser, radAccept *radius.Packet) {
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

func configHuaweiAccept(user *domain.RadiusUser, radAccept *radius.Packet) {
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

func configH3cAccept(user *domain.RadiusUser, radAccept *radius.Packet) {
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

func configZteAccept(user *domain.RadiusUser, radAccept *radius.Packet) {
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
