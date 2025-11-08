package checkers

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// MacBindChecker MAC 绑定检查器
type MacBindChecker struct{}

func (c *MacBindChecker) Name() string {
	return "mac_bind"
}

func (c *MacBindChecker) Order() int {
	return 20 // 在状态和过期之后
}

func (c *MacBindChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User

	// 不检查 MAC 绑定
	if user.BindMac == 0 {
		return nil
	}

	// 获取厂商请求中的 MAC 地址
	vendorReq, ok := authCtx.VendorRequest.(*vendorparsers.VendorRequest)
	if !ok || vendorReq == nil {
		return nil
	}

	// 如果用户 MAC 和请求 MAC 都有效，则检查是否匹配
	if common.IsNotEmptyAndNA(user.MacAddr) && vendorReq.MacAddr != "" && user.MacAddr != vendorReq.MacAddr {
		return errors.NewMacBindError()
	}

	return nil
}
