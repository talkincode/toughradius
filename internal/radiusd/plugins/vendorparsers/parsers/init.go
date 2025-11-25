package parsers

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
)

// init automatically registers all vendor parsers
func init() {
	// Register with the new centralized registry
	vendors.Register(&vendors.VendorInfo{
		Code:        "default",
		Name:        "Standard",
		Description: "Standard RADIUS attributes",
		Parser:      &DefaultParser{},
	})

	vendors.Register(&vendors.VendorInfo{
		Code:        vendors.CodeHuawei,
		Name:        "Huawei",
		Description: "Huawei RADIUS attributes",
		Parser:      &HuaweiParser{},
	})

	vendors.Register(&vendors.VendorInfo{
		Code:        vendors.CodeH3C,
		Name:        "H3C",
		Description: "H3C RADIUS attributes",
		Parser:      &H3CParser{},
	})

	vendors.Register(&vendors.VendorInfo{
		Code:        vendors.CodeZTE,
		Name:        "ZTE",
		Description: "ZTE RADIUS attributes",
		Parser:      &ZTEParser{},
	})
}
