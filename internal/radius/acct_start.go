package toughradius

import (
	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"layeh.com/radius"
)

func (s *AcctService) DoAcctStart(r *radius.Request, vr *VendorRequest, username string, vpe *domain.NetVpe, nasrip string) {
	online := GetNetRadiusOnlineFromRequest(r, vr, vpe, nasrip)
	err := s.AddRadiusOnline(online)
	if err != nil {
		zap.L().Error("add radius online error",
			zap.String("namespace", "radius"),
			zap.String("username", username),
			zap.Error(err),
		)
	}

	if err = s.AddRadiusAccounting(online, true); err != nil {
		zap.L().Error("add radius accounting error",
			zap.String("namespace", "radius"),
			zap.String("username", username),
			zap.Error(err),
		)
	}
}
