package toughradius

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"github.com/talkincode/toughradius/v8/toughradius/vendors/h3c"
	"github.com/talkincode/toughradius/v8/toughradius/vendors/radback"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

var (
	vlanStdRegexp1 = regexp.MustCompile(`\w?\s?\d+/\d+/\d+:(\d+)(\.(\d+))?\s?`)
	vlanStdRegexp2 = regexp.MustCompile(`vlanid=(\d+);(vlanid2=?(\d+);)?`)
)

// ParseVlanIds 解析标准 VLANID 值
func ParseVlanIds(nasportid string) (int64, int64) {
	var vlanid1 int64 = 0
	var vlanid2 int64 = 0
	attrs := vlanStdRegexp1.FindStringSubmatch(nasportid)
	if attrs == nil {
		attrs = vlanStdRegexp2.FindStringSubmatch(nasportid)
	}

	if attrs != nil {
		vlanid1, _ = strconv.ParseInt(attrs[1], 10, 64)
		if attrs[2] != "" {
			vlanid2, _ = strconv.ParseInt(attrs[3], 10, 64)
		}
	}
	return vlanid1, vlanid2
}

// ParseVendor 解析厂商私有属性
func (s *RadiusService) ParseVendor(r *radius.Request, vendorCode string) *VendorRequest {
	switch vendorCode {
	case VendorH3c:
		return parseVendorH3c(r)
	case VendorRadback:
		return parseVendorRadback(r)
	case VendorZte:
		return parseVendorZte(r)
	default:
		return parseVendorDefault(r)
	}
}

// 解析标准属性
func parseVendorDefault(r *radius.Request) *VendorRequest {
	var attrs = new(VendorRequest)
	// 解析 MAC 地址
	macval := rfc2865.CallingStationID_GetString(r.Packet)
	if macval != "" {
		attrs.MacAddr = strings.ReplaceAll(macval, "-", ":")
	} else {
		log.WarnDetail("rfc2865.CallingStationID is empty", zap.String("namespace", "radius"))
	}
	nasportid := rfc2869.NASPortID_GetString(r.Packet)
	if nasportid == "" {
		attrs.Vlanid1 = 0
		attrs.Vlanid2 = 0
	}
	return attrs
}

// 解析 H3C 属性
func parseVendorH3c(r *radius.Request) *VendorRequest {
	var attrs = new(VendorRequest)
	// 解析 MAC 地址
	ipha := h3c.H3CIPHostAddr_GetString(r.Packet)
	if ipha != "" {
		iphalen := len(ipha)
		if len(ipha) > 17 {
			attrs.MacAddr = ipha[iphalen-17:]
		} else {
			attrs.MacAddr = ipha
		}
	} else {
		log.WarnDetail("h3c.H3CIPHostAddr is empty", zap.String("namespace", "radius"))
		macval := rfc2865.CallingStationID_GetString(r.Packet)
		if macval != "" {
			attrs.MacAddr = strings.ReplaceAll(macval, "-", ":")
		} else {
			log.WarnDetail("rfc2865.CallingStationID is empty", zap.String("namespace", "radius"))
		}
	}

	nasportid := rfc2869.NASPortID_GetString(r.Packet)
	if nasportid == "" {
		attrs.Vlanid1 = 0
		attrs.Vlanid2 = 0
	}
	return attrs
}

// 解析 ZTE 属性
func parseVendorZte(r *radius.Request) *VendorRequest {
	var attrs = new(VendorRequest)
	// 解析 MAC 地址
	macval := rfc2865.CallingStationID_GetString(r.Packet)
	if macval != "" {
		if len(macval) > 12 {
			attrs.MacAddr = fmt.Sprintf("%s:%s:%s:%s:%s:%s", macval[0:2], macval[2:4], macval[4:6], macval[6:8], macval[8:10], macval[10:12])
		} else {
			log.WarnDetail("rfc2865.CallingStationID length < 12", zap.String("namespace", "radius"))
		}
	} else {
		log.WarnDetail("rfc2865.CallingStationID is empty", zap.String("namespace", "radius"))
	}
	nasportid := rfc2869.NASPortID_GetString(r.Packet)
	if nasportid == "" {
		attrs.Vlanid1 = 0
		attrs.Vlanid2 = 0
	}
	return attrs
}

// 解析标准属性
func parseVendorRadback(r *radius.Request) *VendorRequest {
	var attrs = new(VendorRequest)
	// 解析 MAC 地址
	macval := radback.MacAddr_GetString(r.Packet)
	if macval != "" {
		attrs.MacAddr = strings.ReplaceAll(macval, "-", ":")
	} else {
		log.WarnDetail("rfc2865.CallingStationID is empty", zap.String("namespace", "radius"))
	}
	nasportid := rfc2869.NASPortID_GetString(r.Packet)
	if nasportid == "" {
		attrs.Vlanid1 = 0
		attrs.Vlanid2 = 0
	}
	return attrs
}
