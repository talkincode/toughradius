package parsers

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
)

// init automatically registers all vendor parsers
func init() {
	registry.RegisterVendorParser(&DefaultParser{})
	registry.RegisterVendorParser(&HuaweiParser{})
	registry.RegisterVendorParser(&H3CParser{})
	registry.RegisterVendorParser(&ZTEParser{})
}
