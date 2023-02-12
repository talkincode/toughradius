package toughradius

import (
	"context"
	"fmt"

	"github.com/talkincode/toughradius/common/zaplog/log"
	"github.com/talkincode/toughradius/models"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

func (s *AcctService) DoAcctNasOn(r *radius.Request) {
	s.BatchClearRadiusOnlineByNas(
		rfc2865.NASIPAddress_Get(r.Packet).String(),
		rfc2865.NASIdentifier_GetString(r.Packet),
	)
}

func (s *AcctService) DoAcctNasOff(r *radius.Request) {
	s.BatchClearRadiusOnlineByNas(
		rfc2865.NASIPAddress_Get(r.Packet).String(),
		rfc2865.NASIdentifier_GetString(r.Packet),
	)
}

func (s *AcctService) DoAcctDisconnect(r *radius.Request, vpe *models.NetVpe, username, nasrip string) {
	packet := radius.New(radius.CodeDisconnectRequest, []byte(vpe.Secret))
	sessionid := rfc2866.AcctSessionID_GetString(r.Packet)
	if sessionid == "" {
		return
	}
	_ = rfc2865.UserName_SetString(packet, username)
	_ = rfc2866.AcctSessionID_Set(packet, []byte(sessionid))
	response, err := radius.Exchange(context.Background(), packet, fmt.Sprintf("%s:%d", nasrip, vpe.CoaPort))
	if err != nil {
		log.Error2("radius disconnect error",
			zap.String("namespace", "radius"),
			zap.String("username", username),
			zap.Error(err),
		)
		return
	}
	log.Info2("radius disconnect done",
		zap.String("namespace", "radius"),
		zap.String("nasip", nasrip),
		zap.Int("coaport", vpe.CoaPort),
		zap.String("request", FmtPacket(packet)),
		zap.String("response", FmtPacket(response)),
	)
}
