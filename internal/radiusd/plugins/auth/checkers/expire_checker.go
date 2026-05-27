package checkers

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
)

// ExpireChecker checks whether the account is expired
type ExpireChecker struct{}

func (c *ExpireChecker) Name() string {
	return "expire"
}

func (c *ExpireChecker) Order() int {
	return 10 // Execute first
}

func (c *ExpireChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User

	if user.ExpireTime.Before(time.Now()) {
		return errors.NewUserExpiredError()
	}

	return nil
}
