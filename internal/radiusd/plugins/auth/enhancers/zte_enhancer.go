package enhancers

import (
	"context"
	"math"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
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
	if !matchVendor(authCtx, vendorZTE) {
		return nil
	}

	user := authCtx.User
	resp := authCtx.Response

	up := clampInt64(int64(user.UpRate)*1024, math.MaxInt32)
	down := clampInt64(int64(user.DownRate)*1024, math.MaxInt32)

	zte.ZTERateCtrlSCRUp_Set(resp, zte.ZTERateCtrlSCRUp(up))
	zte.ZTERateCtrlSCRDown_Set(resp, zte.ZTERateCtrlSCRDown(down))
	return nil
}
