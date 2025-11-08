package enhancers

import (
	"context"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
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
	if !matchVendor(authCtx, vendorMikrotik) {
		return nil
	}

	user := authCtx.User
	resp := authCtx.Response
	mikrotik.MikrotikRateLimit_SetString(resp, fmt.Sprintf("%dk/%dk", user.UpRate, user.DownRate))
	return nil
}
