package enhancers

import (
	"context"
	"math"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/h3c"
)

type H3CAcceptEnhancer struct{}

func NewH3CAcceptEnhancer() *H3CAcceptEnhancer {
	return &H3CAcceptEnhancer{}
}

func (e *H3CAcceptEnhancer) Name() string {
	return "accept-h3c"
}

func (e *H3CAcceptEnhancer) Enhance(ctx context.Context, authCtx *auth.AuthContext) error {
	if authCtx == nil || authCtx.Response == nil || authCtx.User == nil {
		return nil
	}
	if !matchVendor(authCtx, vendors.CodeH3C) {
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
	upPeak := clampInt64(up*4, math.MaxInt32)
	downPeak := clampInt64(down*4, math.MaxInt32)

	_ = h3c.H3CInputAverageRate_Set(resp, h3c.H3CInputAverageRate(up))     //nolint:errcheck
	_ = h3c.H3CInputPeakRate_Set(resp, h3c.H3CInputPeakRate(upPeak))       //nolint:errcheck
	_ = h3c.H3COutputAverageRate_Set(resp, h3c.H3COutputAverageRate(down)) //nolint:errcheck
	_ = h3c.H3COutputPeakRate_Set(resp, h3c.H3COutputPeakRate(downPeak))   //nolint:errcheck
	return nil
}
