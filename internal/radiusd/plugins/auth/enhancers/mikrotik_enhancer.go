package enhancers

import (
	"context"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/mikrotik"
)

type MikrotikAcceptEnhancer struct{}

func NewMikrotikAcceptEnhancer() *MikrotikAcceptEnhancer {
	return &MikrotikAcceptEnhancer{}
}

func (e *MikrotikAcceptEnhancer) Name() string {
	return "accept-mikrotik"
}

func (e *MikrotikAcceptEnhancer) Enhance(ctx context.Context, authCtx *auth.AuthContext) error {
	if authCtx == nil || authCtx.Response == nil || authCtx.User == nil {
		return nil
	}
	if !matchVendor(authCtx, vendors.CodeMikrotik) {
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

	_ = mikrotik.MikrotikRateLimit_SetString(resp, fmt.Sprintf("%dk/%dk", upRate, downRate)) //nolint:errcheck
	return nil
}
