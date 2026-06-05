package radiusd

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"go.uber.org/zap"
	"layeh.com/radius"
)

// ParseVlanIds parses standard VLAN ID values from a NAS-Port-Id string. It
// delegates to vendorparsers.ParseVlanIDs, the single shared implementation
// used by the vendor parsers.
func ParseVlanIds(nasportid string) (int64, int64) {
	return vendorparsers.ParseVlanIDs(nasportid)
}

// ParseVendor uses the plugin system to parse vendor-specific attributes
func (s *RadiusService) ParseVendor(r *radius.Request, vendorCode string) *VendorRequest {
	// Retrieve the corresponding VendorParser from the vendor registry
	parser, ok := vendors.GetParser(vendorCode)
	if !ok {
		zap.L().Warn("vendor parser not found, using default parser",
			zap.String("namespace", "radius"),
			zap.String("vendor_code", vendorCode),
		)
		// e.g., if not found, try the default parser
		parser, ok = vendors.GetParser("default")
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
