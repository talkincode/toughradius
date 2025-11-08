package radiusd

import (
	"github.com/talkincode/toughradius/v9/internal/app"
)

func (s *AuthService) CheckOnlineCount(username string, activeNUm int) error {
	if activeNUm != 0 {
		onlineCount := s.GetRadiusOnlineCount(username)
		if onlineCount >= activeNUm {
			return NewAuthError(app.MetricsRadiusRejectLimit, "user active num over limit")
		}
	}
	return nil
}
