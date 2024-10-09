package toughradius

import (
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"github.com/talkincode/toughradius/v8/models"
	"go.uber.org/zap"
	"layeh.com/radius"
)

func (s *AcctService) DoAcctStop(r *radius.Request, vr *VendorRequest, username string, vpe *models.NetVpe, nasrip string) {
	online := GetNetRadiusOnlineFromRequest(r, vr, vpe, nasrip)
	if err := s.EndRadiusAccounting(online); err != nil {
		err := s.AddRadiusAccounting(online, false)
		if err != nil {
			log.ErrorDetail("add radius accounting error",
				zap.String("namespace", "radius"),
				zap.String("username", username),
				zap.Error(err),
			)
		}
	}

	if err := s.RemoveRadiusOnline(online.AcctSessionId); err != nil {
		log.ErrorDetail("remove radius online error",
			zap.String("namespace", "radius"),
			zap.String("username", username),
			zap.Error(err),
		)
	}

}
