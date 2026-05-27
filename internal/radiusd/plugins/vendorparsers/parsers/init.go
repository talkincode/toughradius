package parsers

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
)

// init automatically registers all vendor parsers
func init() {
	// Register with the new centralized registry
	_ = vendors.Register(&vendors.VendorInfo{ //nolint:errcheck
		Code:        "default",
		Name:        "Standard",
		Description: "Standard RADIUS attributes",
		Parser:      &DefaultParser{},
	})

	_ = vendors.Register(&vendors.VendorInfo{ //nolint:errcheck
		Code:        vendors.CodeHuawei,
		Name:        "Huawei",
		Description: "Huawei RADIUS attributes",
		Parser:      &HuaweiParser{},
	})

	_ = vendors.Register(&vendors.VendorInfo{ //nolint:errcheck
		Code:        vendors.CodeH3C,
		Name:        "H3C",
		Description: "H3C RADIUS attributes",
		Parser:      &H3CParser{},
	})

	_ = vendors.Register(&vendors.VendorInfo{ //nolint:errcheck
		Code:        vendors.CodeZTE,
		Name:        "ZTE",
		Description: "ZTE RADIUS attributes",
		Parser:      &ZTEParser{},
	})
}
