package checkers

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
)

// OnlineCountChecker 在线数检查器
type OnlineCountChecker struct {
	sessionRepo repository.SessionRepository
}

// NewOnlineCountChecker 创建在线数检查器
func NewOnlineCountChecker(sessionRepo repository.SessionRepository) *OnlineCountChecker {
	return &OnlineCountChecker{sessionRepo: sessionRepo}
}

func (c *OnlineCountChecker) Name() string {
	return "online_count"
}

func (c *OnlineCountChecker) Order() int {
	return 30 // 在绑定检查之后
}

func (c *OnlineCountChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User

	// activeNum为0表示不限制
	if user.ActiveNum == 0 {
		return nil
	}

	count, err := c.sessionRepo.CountByUsername(ctx, user.Username)
	if err != nil {
		return err
	}

	if count >= user.ActiveNum {
		return errors.NewOnlineLimitError("user online count exceeded")
	}

	return nil
}
