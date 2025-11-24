# Vendor Registry 使用示例

`internal/radiusd/vendors` 提供了一个线程安全的注册表，用于集中管理所有厂商的解析器与响应构建器。下面的示例展示了如何在自定义插件或测试中注册新的厂商，并从 Registry 中检索相关元数据。

```go
package customvendor

import (
    "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
    "github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
)

type FooParser struct{}

func (p *FooParser) VendorCode() string { return "65000" }
func (p *FooParser) VendorName() string { return "FooVendor" }
func (p *FooParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
    // ... populate vendor request ...
    return &vendorparsers.VendorRequest{}, nil
}

func init() {
    _ = vendors.Register(&vendors.VendorInfo{
        Code:   "65000",
        Name:   "FooVendor",
        Parser: &FooParser{},
    })
}
```

```go
info, ok := vendors.Get("65000")
if !ok {
    panic("vendor not registered")
}
parser := info.Parser
request, err := parser.Parse(radiusRequest)
```

> 最佳实践：优先重用 `vendors.CodeXXX` 常量（位于 `internal/radiusd/vendors/codes.go`），避免在代码中散落硬编码 Vendor ID。
