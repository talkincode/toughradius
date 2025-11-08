package checkers

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
)

// ExpireChecker 过期检查器
type ExpireChecker struct{}

func (c *ExpireChecker) Name() string {
	return "expire"
}

func (c *ExpireChecker) Order() int {
	return 10 // 最先执行
}

func (c *ExpireChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User

	if user.ExpireTime.Before(time.Now()) {
		return errors.NewUserExpiredError()
	}

	return nil
}
