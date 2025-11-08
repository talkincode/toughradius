package vendorparsers

import (
	"layeh.com/radius"
)

// VendorRequest 厂商请求数据
type VendorRequest struct {
	MacAddr string
	Vlanid1 int64
	Vlanid2 int64
}

// VendorParser 厂商属性解析器接口
type VendorParser interface {
	// VendorCode 返回厂商代码
	VendorCode() string

	// VendorName 返回厂商名称
	VendorName() string

	// Parse 解析厂商私有属性
	Parse(r *radius.Request) (*VendorRequest, error)
}

// VendorResponseBuilder 厂商响应构建器接口
type VendorResponseBuilder interface {
	// VendorCode 返回厂商代码
	VendorCode() string

	// Build 构建厂商特定的响应属性
	Build(resp *radius.Packet, user interface{}, vlanId1, vlanId2 int) error
}
