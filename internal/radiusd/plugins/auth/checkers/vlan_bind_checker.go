package checkers

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
)

// VlanBindChecker VLAN 绑定检查器
type VlanBindChecker struct{}

func (c *VlanBindChecker) Name() string {
	return "vlan_bind"
}

func (c *VlanBindChecker) Order() int {
	return 21 // 在 MAC 绑定之后
}

func (c *VlanBindChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User

	// 不检查 VLAN 绑定
	if user.BindVlan == 0 {
		return nil
	}

	// 获取厂商请求中的 VLAN ID
	vendorReq, ok := authCtx.VendorRequest.(*vendorparsers.VendorRequest)
	if !ok || vendorReq == nil {
		return nil
	}

	reqvid1 := int(vendorReq.Vlanid1)
	reqvid2 := int(vendorReq.Vlanid2)

	// 检查 VLAN ID 1
	if user.Vlanid1 != 0 && reqvid1 != 0 && user.Vlanid1 != reqvid1 {
		return errors.NewVlanBindError()
	}

	// 检查 VLAN ID 2
	if user.Vlanid2 != 0 && reqvid2 != 0 && user.Vlanid2 != reqvid2 {
		return errors.NewVlanBindError()
	}

	return nil
}
