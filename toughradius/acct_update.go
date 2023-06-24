package toughradius

import (
	"time"

	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"github.com/talkincode/toughradius/v8/models"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2866"
)

func (s *AcctService) DoAcctUpdateBefore(r *radius.Request, vr *VendorRequest, user *models.RadiusUser, vpe *models.NetVpe, nasrip string) {
	// 用户状态变更为停用后触发下线
	if user.Status == common.DISABLED {
		s.DoAcctDisconnect(r, vpe, user.Username, nasrip)
	}

	// 用户过期后触发下线
	if user.ExpireTime.Before(time.Now()) {
		go s.DoAcctDisconnect(r, vpe, user.Username, nasrip)
	}

	s.DoAcctUpdate(r, vr, user.Username, vpe, nasrip)
}

func (s *AcctService) DoAcctUpdate(r *radius.Request, vr *VendorRequest, username string, vpe *models.NetVpe, nasrip string) {

	online := GetNetRadiusOnlineFromRequest(r, vr, vpe, nasrip)
	// 如果在线用户不存在立即新增
	exists := s.ExistRadiusOnline(rfc2866.AcctSessionID_GetString(r.Packet))
	if !exists {
		err := s.AddRadiusOnline(online)
		if err != nil {
			log.Error2("add radius online error",
				zap.String("namespace", "radius"),
				zap.String("username", username),
				zap.Error(err),
			)
		}
	}

	// 更新在线信息
	err := s.UpdateRadiusOnlineData(online)
	if err != nil {
		log.Error2("update radius online error",
			zap.String("namespace", "radius"),
			zap.String("username", username),
			zap.Error(err),
		)
	}

}
