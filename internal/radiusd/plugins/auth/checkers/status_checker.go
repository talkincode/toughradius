package checkers

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// StatusChecker verifies user status
type StatusChecker struct{}

func (c *StatusChecker) Name() string {
	return "status"
}

func (c *StatusChecker) Order() int {
	return 5 // Execute early
}

func (c *StatusChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User

	if user.Status == common.DISABLED {
		return errors.NewUserDisabledError()
	}

	return nil
}
