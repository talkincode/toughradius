package enhancers

import (
	"context"
	"math"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/zte"
)

type ZTEAcceptEnhancer struct{}

func NewZTEAcceptEnhancer() *ZTEAcceptEnhancer {
	return &ZTEAcceptEnhancer{}
}

func (e *ZTEAcceptEnhancer) Name() string {
	return "accept-zte"
}

func (e *ZTEAcceptEnhancer) Enhance(ctx context.Context, authCtx *auth.AuthContext) error {
	if authCtx == nil || authCtx.Response == nil || authCtx.User == nil {
		return nil
	}
	if !matchVendor(authCtx, vendors.CodeZTE) {
		return nil
	}

	user := authCtx.User
	resp := authCtx.Response

	// Get profile cache from metadata
	var profileCache interface{}
	if authCtx.Metadata != nil {
		profileCache = authCtx.Metadata["profile_cache"]
	}

	// Use getter methods for bandwidth rates
	upRate := user.GetUpRate(profileCache)
	downRate := user.GetDownRate(profileCache)

	up := clampInt64(int64(upRate)*1024, math.MaxInt32)
	down := clampInt64(int64(downRate)*1024, math.MaxInt32)

	_ = zte.ZTERateCtrlSCRUp_Set(resp, zte.ZTERateCtrlSCRUp(up))       //nolint:errcheck
	_ = zte.ZTERateCtrlSCRDown_Set(resp, zte.ZTERateCtrlSCRDown(down)) //nolint:errcheck
	return nil
}
