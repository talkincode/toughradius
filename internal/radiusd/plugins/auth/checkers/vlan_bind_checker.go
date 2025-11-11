package checkers

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
)

// VlanBindChecker enforces VLAN binding
type VlanBindChecker struct{}

func (c *VlanBindChecker) Name() string {
	return "vlan_bind"
}

func (c *VlanBindChecker) Order() int {
	return 21 // Execute after MAC bind
}

func (c *VlanBindChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User

		// Skip VLAN bind check
	if user.BindVlan == 0 {
		return nil
	}

	// Get VLAN ID from the vendor request
	vendorReq, ok := authCtx.VendorRequest.(*vendorparsers.VendorRequest)
	if !ok || vendorReq == nil {
		return nil
	}

	reqvid1 := int(vendorReq.Vlanid1)
	reqvid2 := int(vendorReq.Vlanid2)

	// Check VLAN ID 1
	if user.Vlanid1 != 0 && reqvid1 != 0 && user.Vlanid1 != reqvid1 {
		return errors.NewVlanBindError()
	}

	// Check VLAN ID 2
	if user.Vlanid2 != 0 && reqvid2 != 0 && user.Vlanid2 != reqvid2 {
		return errors.NewVlanBindError()
	}

	return nil
}
