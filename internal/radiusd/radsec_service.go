package radiusd

import (
	"context"
	"net"

	"go.uber.org/zap"
	"layeh.com/radius"
)

type RadsecService struct {
	AuthService *AuthService
	AcctService *AcctService
}

func (s *RadsecService) RADIUSSecret(ctx context.Context, remoteAddr net.Addr) ([]byte, error) {
	return []byte("radsec"), nil
}

func NewRadsecService(authService *AuthService, acctService *AcctService) *RadsecService {
	return &RadsecService{AuthService: authService, AcctService: acctService}
}

func (s *RadsecService) ServeRADIUS(w radius.ResponseWriter, r *radius.Request) {
	switch r.Code {
	case radius.CodeAccessRequest:
		s.AuthService.ServeRADIUS(w, r)
	case radius.CodeAccountingRequest:
		s.AcctService.ServeRADIUS(w, r)
	default:
		zap.L().Info("radius radsec message",
			zap.String("namespace", "radius"),
			zap.Int("code", int(r.Code)),
		)
	}
}
