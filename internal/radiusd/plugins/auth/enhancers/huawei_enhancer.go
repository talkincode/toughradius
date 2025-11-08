package enhancers

import (
	"context"
	"math"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/huawei"
)

type HuaweiAcceptEnhancer struct{}

func NewHuaweiAcceptEnhancer() *HuaweiAcceptEnhancer {
	return &HuaweiAcceptEnhancer{}
}

func (e *HuaweiAcceptEnhancer) Name() string {
	return "accept-huawei"
}

func (e *HuaweiAcceptEnhancer) Enhance(ctx context.Context, authCtx *auth.AuthContext) error {
	if authCtx == nil || authCtx.Response == nil || authCtx.User == nil {
		return nil
	}
	if !matchVendor(authCtx, vendorHuawei) {
		return nil
	}

	user := authCtx.User
	resp := authCtx.Response

	up := clampInt64(int64(user.UpRate)*1024, math.MaxInt32)
	down := clampInt64(int64(user.DownRate)*1024, math.MaxInt32)
	upPeak := clampInt64(up*4, math.MaxInt32)
	downPeak := clampInt64(down*4, math.MaxInt32)

	huawei.HuaweiInputAverageRate_Set(resp, huawei.HuaweiInputAverageRate(up))
	huawei.HuaweiInputPeakRate_Set(resp, huawei.HuaweiInputPeakRate(upPeak))
	huawei.HuaweiOutputAverageRate_Set(resp, huawei.HuaweiOutputAverageRate(down))
	huawei.HuaweiOutputPeakRate_Set(resp, huawei.HuaweiOutputPeakRate(downPeak))
	return nil
}
