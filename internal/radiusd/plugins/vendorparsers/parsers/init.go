package parsers

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
)

// init 自动注册所有厂商解析器
func init() {
	registry.RegisterVendorParser(&DefaultParser{})
	registry.RegisterVendorParser(&HuaweiParser{})
	registry.RegisterVendorParser(&H3CParser{})
	registry.RegisterVendorParser(&ZTEParser{})
}
