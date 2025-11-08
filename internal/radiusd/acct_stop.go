package radiusd

import (
	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"layeh.com/radius"
)

func (s *AcctService) DoAcctStop(r *radius.Request, vr *VendorRequest, username string, nas *domain.NetNas, nasrip string) {
	online := GetNetRadiusOnlineFromRequest(r, vr, nas, nasrip)
	if err := s.EndRadiusAccounting(online); err != nil {
		err := s.AddRadiusAccounting(online, false)
		if err != nil {
			zap.L().Error("add radius accounting error",
				zap.String("namespace", "radius"),
				zap.String("username", username),
				zap.Error(err),
			)
		}
	}

	if err := s.RemoveRadiusOnline(online.AcctSessionId); err != nil {
		zap.L().Error("remove radius online error",
			zap.String("namespace", "radius"),
			zap.String("username", username),
			zap.Error(err),
		)
	}

}
