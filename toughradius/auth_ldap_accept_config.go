package toughradius

import (
	"fmt"
	"math"
	"net"
	"time"

	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/toughradius/vendors/cisco"
	"github.com/talkincode/toughradius/toughradius/vendors/h3c"
	"github.com/talkincode/toughradius/toughradius/vendors/huawei"
	"github.com/talkincode/toughradius/toughradius/vendors/ikuai"
	"github.com/talkincode/toughradius/toughradius/vendors/mikrotik"
	"github.com/talkincode/toughradius/toughradius/vendors/radback"
	"github.com/talkincode/toughradius/toughradius/vendors/zte"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

func (s *AuthService) LdapAcceptAcceptConfig(user *LdapRadisProfile, vendorCode string, radAccept *radius.Packet) {
	ldapConfigDefaultAccept(s, user, radAccept)
	switch vendorCode {
	case VendorHuawei:
		ldapConfigHuaweiAccept(user, radAccept)
	case VendorH3c:
		ldapConfigH3cAccept(user, radAccept)
	case VendorRadback:
		ldapConfigRadbackAccept(user, radAccept)
	case VendorZte:
		ldapConfigZteAccept(user, radAccept)
	case VendorCisco:
		ldapConfigCiscoAccept(user, radAccept)
	case VendorMikrotik:
		ldapConfigMikroTikAccept(user, radAccept)
	case VendorIkuai:
		ldapConfigIkuaiAccept(user, radAccept)
	}
}

// 设置标准 RADIUS 属性
func ldapConfigDefaultAccept(s *AuthService, user *LdapRadisProfile, radAccept *radius.Packet) {
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

func ldapConfigMikroTikAccept(user *LdapRadisProfile, radAccept *radius.Packet) {
	mikrotik.MikrotikRateLimit_SetString(radAccept, fmt.Sprintf("%dk/%dk", user.UpRate, user.DownRate))
}

func ldapConfigIkuaiAccept(user *LdapRadisProfile, radAccept *radius.Packet) {
	var up = int64(user.UpRate) * 1024 * 8
	var down = int64(user.DownRate) * 1024 * 8
	if up > math.MaxInt32 {
		up = math.MaxInt32
	}
	if down > math.MaxInt32 {
		down = math.MaxInt32
	}
	if up > 0 {
		ikuai.RPUpstreamSpeedLimit_Set(radAccept, ikuai.RPUpstreamSpeedLimit(up))
	}
	if down > 0 {
		ikuai.RPDownstreamSpeedLimit_Set(radAccept, ikuai.RPDownstreamSpeedLimit(down))
	}
}

func ldapConfigHuaweiAccept(user *LdapRadisProfile, radAccept *radius.Packet) {
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
	if up > 0 {
		huawei.HuaweiInputAverageRate_Set(radAccept, huawei.HuaweiInputAverageRate(up))
		huawei.HuaweiInputPeakRate_Set(radAccept, huawei.HuaweiInputPeakRate(upPeak))
	}

	if down > 0 {
		huawei.HuaweiOutputAverageRate_Set(radAccept, huawei.HuaweiOutputAverageRate(down))
		huawei.HuaweiOutputPeakRate_Set(radAccept, huawei.HuaweiOutputPeakRate(downPeak))
	}

	if common.IsNotEmptyAndNA(user.Domain) {
		huawei.HuaweiDomainName_SetString(radAccept, user.Domain)
	}
}

func ldapConfigH3cAccept(user *LdapRadisProfile, radAccept *radius.Packet) {
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

	if up > 0 {
		h3c.H3CInputAverageRate_Set(radAccept, h3c.H3CInputAverageRate(up))
		h3c.H3CInputPeakRate_Set(radAccept, h3c.H3CInputPeakRate(upPeak))
	}

	if down > 0 {
		h3c.H3COutputAverageRate_Set(radAccept, h3c.H3COutputAverageRate(down))
		h3c.H3COutputPeakRate_Set(radAccept, h3c.H3COutputPeakRate(downPeak))
	}

}

func ldapConfigRadbackAccept(user *LdapRadisProfile, radAccept *radius.Packet) {
	if common.IsNotEmptyAndNA(user.LimitPolicy) {
		radback.SubscriberProfileName_SetString(radAccept, user.LimitPolicy)
	}
	if common.IsNotEmptyAndNA(user.Domain) {
		radback.ContextName_SetString(radAccept, user.Domain)
	}
}

func ldapConfigZteAccept(user *LdapRadisProfile, radAccept *radius.Packet) {
	var up = int64(user.UpRate) * 1024
	var down = int64(user.DownRate) * 1024
	if up > math.MaxInt32 {
		up = math.MaxInt32
	}
	if down > math.MaxInt32 {
		down = math.MaxInt32
	}
	if up > 0 {
		zte.ZTERateCtrlSCRUp_Set(radAccept, zte.ZTERateCtrlSCRUp(up))
	}
	if down > 0 {
		zte.ZTERateCtrlSCRDown_Set(radAccept, zte.ZTERateCtrlSCRDown(down))
	}
	if common.IsNotEmptyAndNA(user.Domain) {
		zte.ZTEContextName_SetString(radAccept, user.Domain)
	}
}

func ldapConfigCiscoAccept(user *LdapRadisProfile, radAccept *radius.Packet) {
	if common.IsNotEmptyAndNA(user.UpLimitPolicy) {
		cisco.CiscoAVPair_Add(radAccept, []byte(fmt.Sprintf("sub-qos-policy-in=%s", user.UpLimitPolicy)))
	}
	if common.IsNotEmptyAndNA(user.DownLimitPolicy) {
		cisco.CiscoAVPair_Add(radAccept, []byte(fmt.Sprintf("sub-qos-policy-out=%s", user.DownLimitPolicy)))
	}
}
