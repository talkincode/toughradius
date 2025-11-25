package radiusd

import (
	"regexp"
	"strconv"

	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"go.uber.org/zap"
	"layeh.com/radius"
)

var (
	vlanStdRegexp1 = regexp.MustCompile(`\w?\s?\d+/\d+/\d+:(\d+)(\.(\d+))?\s?`)
	vlanStdRegexp2 = regexp.MustCompile(`vlanid=(\d+);(vlanid2=?(\d+);)?`)
)

// ParseVlanIds parses standard VLAN ID values
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

// ParseVendor uses the plugin system to parse vendor-specific attributes
func (s *RadiusService) ParseVendor(r *radius.Request, vendorCode string) *VendorRequest {
	// Retrieve the corresponding VendorParser from the registry
	parser, ok := registry.GetVendorParser(vendorCode)
	if !ok {
		zap.L().Warn("vendor parser not found, using default parser",
			zap.String("namespace", "radius"),
			zap.String("vendor_code", vendorCode),
		)
		// e.g., if not found, try the default parser
		parser, ok = registry.GetVendorParser("default")
		if !ok {
			// e.g., if even the default parser is missing, return an empty result
			zap.L().Error("default vendor parser not found",
				zap.String("namespace", "radius"),
			)
			return &VendorRequest{}
		}
	}

	// Use the plugin to parse
	vendorReq, err := parser.Parse(r)
	if err != nil {
		zap.L().Error("vendor parser error",
			zap.String("namespace", "radius"),
			zap.String("vendor_code", vendorCode),
			zap.Error(err),
		)
		return &VendorRequest{}
	}

	// Convert toVendorRequest
	return &VendorRequest{
		MacAddr: vendorReq.MacAddr,
		Vlanid1: vendorReq.Vlanid1,
		Vlanid2: vendorReq.Vlanid2,
	}
}
