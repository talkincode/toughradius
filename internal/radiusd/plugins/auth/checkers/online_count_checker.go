package checkers

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
)

// OnlineCountChecker enforces online count limits
type OnlineCountChecker struct {
	sessionRepo repository.SessionRepository
}

// NewOnlineCountChecker creates an online count checker
func NewOnlineCountChecker(sessionRepo repository.SessionRepository) *OnlineCountChecker {
	return &OnlineCountChecker{sessionRepo: sessionRepo}
}

func (c *OnlineCountChecker) Name() string {
	return "online_count"
}

func (c *OnlineCountChecker) Order() int {
	return 30 // Execute after the bind check
}

func (c *OnlineCountChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User

	// Get profile cache from metadata
	var profileCache interface{}
	if authCtx.Metadata != nil {
		profileCache = authCtx.Metadata["profile_cache"]
	}

	// An activeNum of 0 indicates no limit
	activeNum := user.GetActiveNum(profileCache)
	if activeNum == 0 {
		return nil
	}

	count, err := c.sessionRepo.CountByUsername(ctx, user.Username)
	if err != nil {
		return err
	}

	if count >= activeNum {
		return errors.NewOnlineLimitError("user online count exceeded")
	}

	return nil
}
