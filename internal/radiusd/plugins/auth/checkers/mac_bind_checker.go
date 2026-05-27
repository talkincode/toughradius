package checkers

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// MacBindChecker enforces MAC binding
type MacBindChecker struct{}

func (c *MacBindChecker) Name() string {
	return "mac_bind"
}

func (c *MacBindChecker) Order() int {
	return 20 // Execute after status and expiration checks
}

func (c *MacBindChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User

	// Get profile cache from metadata
	var profileCache interface{}
	if authCtx.Metadata != nil {
		profileCache = authCtx.Metadata["profile_cache"]
	}

	// Skip MAC bind check
	if user.GetBindMac(profileCache) == 0 {
		return nil
	}

	// Get MAC addresses from the vendor request
	vendorReq, ok := authCtx.VendorRequest.(*vendorparsers.VendorRequest)
	if !ok || vendorReq == nil {
		return nil
	}

	// e.g., if both the user MAC and request MAC are present, ensure they match
	if common.IsNotEmptyAndNA(user.MacAddr) && vendorReq.MacAddr != "" && user.MacAddr != vendorReq.MacAddr {
		return errors.NewMacBindError()
	}

	return nil
}
