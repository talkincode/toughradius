package toughradius

import (
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2866"
)

func (s *AcctService) DoAcctUpdateBefore(r *radius.Request, vr *VendorRequest, user *domain.RadiusUser, nas *domain.NetNas, nasrip string) {
	// 用户状态变更为停用后触发下线
	if user.Status == common.DISABLED {
		s.DoAcctDisconnect(r, nas, user.Username, nasrip)
	}

	// 用户过期后触发下线
	if user.ExpireTime.Before(time.Now()) {
		go s.DoAcctDisconnect(r, nas, user.Username, nasrip)
	}

	s.DoAcctUpdate(r, vr, user.Username, nas, nasrip)
}

func (s *AcctService) DoAcctUpdate(r *radius.Request, vr *VendorRequest, username string, nas *domain.NetNas, nasrip string) {

	online := GetNetRadiusOnlineFromRequest(r, vr, nas, nasrip)
	// 如果在线用户不存在立即新增
	exists := s.ExistRadiusOnline(rfc2866.AcctSessionID_GetString(r.Packet))
	if !exists {
		err := s.AddRadiusOnline(online)
		if err != nil {
			zap.L().Error("add radius online error",
				zap.String("namespace", "radius"),
				zap.String("username", username),
				zap.Error(err),
			)
		}
	}

	// 更新在线信息
	err := s.UpdateRadiusOnlineData(online)
	if err != nil {
		zap.L().Error("update radius online error",
			zap.String("namespace", "radius"),
			zap.String("username", username),
			zap.Error(err),
		)
	}

}
