package toughradius

import (
	"errors"
	"strings"

	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/zaplog/log"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

// 记账服务
type AcctService struct {
	*RadiusService
}

func NewAcctService(radiusService *RadiusService) *AcctService {
	return &AcctService{RadiusService: radiusService}
}

func (s *AcctService) ServeRADIUS(w radius.ResponseWriter, r *radius.Request) {
	defer func() {
		if ret := recover(); ret != nil {
			err, ok := ret.(error)
			if ok {
				log.Error2("radius accounting error",
					zap.Error(err),
					zap.String("namespace", "radius"),
					zap.String("metrics", app.MetricsRadiusAcctDrop),
				)
			}
		}
	}()

	if r == nil {
		return
	}

	if app.GConfig().Radiusd.Debug {
		log.Debug(FmtRequest(r))
	}

	// NAS 接入检查
	raddrstr := r.RemoteAddr.String()
	nasrip := raddrstr[:strings.Index(raddrstr, ":")]
	var identifier = rfc2865.NASIdentifier_GetString(r.Packet)
	vpe, err := s.GetNas(nasrip, identifier)
	common.Must(err)

	// 重新设置数据报文秘钥
	r.Secret = []byte(vpe.Secret)
	r.Packet.Secret = []byte(vpe.Secret)

	statusType := rfc2866.AcctStatusType_Get(r.Packet)

	// 用户名检查
	var username string
	if statusType != rfc2866.AcctStatusType_Value_AccountingOn &&
		statusType != rfc2866.AcctStatusType_Value_AccountingOff {
		username = rfc2865.UserName_GetString(r.Packet)
		if username == "" {
			common.Must(errors.New("username is empty"))
		}
	}

	defer s.ReleaseAuthRateLimit(username)

	// s.CheckRequestSecret(r.Packet, []byte(vpe.Secret))

	vendorReq := s.ParseVendor(r, vpe.VendorCode)

	// Ldap acct
	if vpe.LdapId != 0 {
		_, err := s.GetLdapServer(vpe.LdapId)
		common.Must(err)
		s.SendResponse(w, r)
		// check ldap auth
		common.Must(s.TaskPool.Submit(func() {
			s.LdapUserAcct(r, vendorReq, username, vpe, nasrip)
		}))

		return
	}

	s.SendResponse(w, r)

	log.Info2("radius accounting",
		zap.String("namespace", "radius"),
		zap.String("metrics", app.MetricsRadiusAccounting),
	)

	// async process accounting
	common.Must(s.TaskPool.Submit(func() {
		switch statusType {
		case rfc2866.AcctStatusType_Value_Start:
			log.Info2("radius accounting start",
				zap.String("namespace", "radius"),
				zap.String("metrics", app.MetricsRadiusOline),
			)
			user, err := s.GetUserForAcct(username)
			common.Must(err)
			s.DoAcctStart(r, vendorReq, user.Username, vpe, nasrip)
		case rfc2866.AcctStatusType_Value_InterimUpdate:
			user, err := s.GetUserForAcct(username)
			common.Must(err)
			s.DoAcctUpdateBefore(r, vendorReq, user, vpe, nasrip)
		case rfc2866.AcctStatusType_Value_Stop:
			log.Info2("radius accounting stop",
				zap.String("namespace", "radius"),
				zap.String("metrics", app.MetricsRadiusOffline),
			)
			user, err := s.GetUserForAcct(username)
			common.Must(err)
			s.DoAcctStop(r, vendorReq, user.Username, vpe, nasrip)
		case rfc2866.AcctStatusType_Value_AccountingOn:
			s.DoAcctNasOn(r)
		case rfc2866.AcctStatusType_Value_AccountingOff:
			s.DoAcctNasOff(r)
		}
	}))
}

func (s *AcctService) SendResponse(w radius.ResponseWriter, r *radius.Request) {
	resp := r.Response(radius.CodeAccountingResponse)
	err := w.Write(resp)
	if err != nil {
		log.Error2("radius accounting response error",
			zap.Error(err),
			zap.String("namespace", "radius"),
			zap.String("metrics", app.MetricsRadiusAcctDrop),
		)
		return
	}

	if app.GConfig().Radiusd.Debug {
		log.Debug(FmtResponse(resp, r.RemoteAddr))
	}

}
