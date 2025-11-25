package vendorparsers

import (
	"layeh.com/radius"
)

// VendorRequest holds vendor-specific request data
type VendorRequest struct {
	MacAddr string
	Vlanid1 int64
	Vlanid2 int64
}

// VendorParser defines the vendor attribute parser interface
type VendorParser interface {
	// VendorCode returns the vendor ID
	VendorCode() string

	// VendorName returns the vendor name
	VendorName() string

	// Parse extracts vendor-specific attributes
	Parse(r *radius.Request) (*VendorRequest, error)
}

// VendorResponseBuilder defines vendor response builders
type VendorResponseBuilder interface {
	// VendorCode returns the vendor ID
	VendorCode() string

	// Build constructs vendor-specific response attributes
	Build(resp *radius.Packet, user interface{}, vlanId1, vlanId2 int) error
}
