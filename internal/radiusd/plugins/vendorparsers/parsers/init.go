package parsers

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
)

// init automatically registers all vendor parsers.
//
// Only the vendors listed here have an actual attribute parser. Many more
// vendors ship an attribute dictionary under internal/radiusd/vendors/<vendor>
// (constants and VSA definitions), but a dictionary only describes attributes —
// it does not extract them. In other words, dictionary support is not parse
// support: requests from a vendor without a parser registered below fall back
// to the DefaultParser, which reads only the standard RADIUS attributes.
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

	_ = vendors.Register(&vendors.VendorInfo{ //nolint:errcheck
		Code:        vendors.CodeRadback,
		Name:        "Radback",
		Description: "Radback RADIUS attributes",
		Parser:      &RadbackParser{},
	})

	_ = vendors.Register(&vendors.VendorInfo{ //nolint:errcheck
		Code:        vendors.CodeAlcatel,
		Name:        "Alcatel",
		Description: "Alcatel RADIUS attributes",
		Parser:      &AlcatelParser{},
	})
}
