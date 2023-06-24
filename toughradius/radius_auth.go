package toughradius

import (
	"strings"

	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"github.com/talkincode/toughradius/v8/models"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type AuthService struct {
	*RadiusService
}

func NewAuthService(radiusService *RadiusService) *AuthService {
	return &AuthService{RadiusService: radiusService}
}

func (s *AuthService) ServeRADIUS(w radius.ResponseWriter, r *radius.Request) {
	defer func() {
		if ret := recover(); ret != nil {
			switch ret.(type) {
			case error:
				err := ret.(error)
				log.Error2("radius auth error",
					zap.Error(err),
					zap.String("namespace", "radius"),
					zap.String("metrics", app.MetricsRadiusAuthDrop),
				)
				s.SendReject(w, r, err)
			case AuthError:
				err := ret.(AuthError)
				log.Error2("radius auth error",
					zap.String("namespace", "radius"),
					zap.String("metrics", err.Type),
					zap.Error(err.Err),
				)
				s.SendReject(w, r, err.Err)
			}
		}
	}()

	if r == nil {
		return
	}

	if app.GConfig().Radiusd.Debug {
		log.Info(FmtRequest(r))
	}

	// nas access check
	raddrstr := r.RemoteAddr.String()
	ip := raddrstr[:strings.Index(raddrstr, ":")]
	var identifier = rfc2865.NASIdentifier_GetString(r.Packet)
	username := rfc2865.UserName_GetString(r.Packet)
	callingStationID := rfc2865.CallingStationID_GetString(r.Packet)

	// Username empty  check
	if username == "" {
		s.CheckRadAuthError(callingStationID, ip, NewAuthError(app.MetricsRadiusRejectNotExists, "username is empty of client mac"))
	}

	s.CheckRadAuthError(username, ip, s.CheckAuthRateLimit(username))

	vpe, err := s.GetNas(ip, identifier)
	s.CheckRadAuthError(username, ip, err)

	//  setup new packet secret
	r.Secret = []byte(vpe.Secret)
	r.Packet.Secret = []byte(vpe.Secret)

	// s.CheckRequestSecret(r.Packet, []byte(vpe.Secret))

	response := r.Response(radius.CodeAccessAccept)
	vendorReq := s.ParseVendor(r, vpe.VendorCode)

	// ----------------------------------------------------------------------------------------------------
	// Ldap auth
	if vpe.LdapId != 0 {
		var lnode *models.NetLdapServer
		lnode, err = s.GetLdapServer(vpe.LdapId)
		s.CheckRadAuthError(username, ip, err)
		var userProfile *LdapRadisProfile
		userProfile, err = s.LdapUserAuth(w, r, username, lnode, response, vendorReq)
		s.CheckRadAuthError(username, ip, err)
		s.LdapAcceptAcceptConfig(userProfile, vpe.VendorCode, response)
		s.SendAccept(w, r, response)

		log.Info2("radius ldap auth sucess",
			zap.String("namespace", "radius"),
			zap.String("username", username),
			zap.String("nasip", ip),
			zap.String("result", "success"),
			zap.String("metrics", app.MetricsRadiusAccept),
		)

		return
	}

	// ----------------------------------------------------------------------------------------------------
	// Fetch validate user
	isMacAuth := vendorReq.MacAddr == username
	user, err := s.GetValidUser(username, isMacAuth)
	s.CheckRadAuthError(username, ip, err)

	if !isMacAuth {
		// check subscribe active num
		s.CheckRadAuthError(username, ip, s.CheckOnlineCount(username, user.ActiveNum))

		// Username Mac bind check
		s.CheckRadAuthError(username, ip, s.CheckMacBind(user, vendorReq))

		// Username vlanid check
		s.CheckRadAuthError(username, ip, s.CheckVlanBind(user, vendorReq))
	}

	// Password check
	// if mschapv2 auth, will set accept attribute
	localpwd, err := s.GetLocalPassword(user, isMacAuth)
	s.CheckRadAuthError(username, ip, err)
	s.CheckRadAuthError(username, ip, s.CheckPassword(r, user.Username, localpwd, response, isMacAuth))

	// setup accept
	s.AcceptAcceptConfig(user, vpe.VendorCode, response)

	// send accept
	s.SendAccept(w, r, response)

	// update subscribe vlan and mac
	s.UpdateBind(user, vendorReq)

	log.Info2("radius auth sucess",
		zap.String("namespace", "radius"),
		zap.String("username", username),
		zap.String("nasip", ip),
		zap.String("result", "success"),
		zap.String("metrics", app.MetricsRadiusAccept),
	)
}

func (s *AuthService) SendAccept(w radius.ResponseWriter, r *radius.Request, resp *radius.Packet) {
	defer func() {
		if ret := recover(); ret != nil {
			err2, ok := ret.(error)
			if ok {
				log.Error2("radius write accept error",
					zap.String("namespace", "radius"),
					zap.String("metrics", app.MetricsRadiusAuthDrop),
					zap.Error(err2),
				)
			}
		}
	}()
	common.Must(w.Write(resp))

	if app.GConfig().Radiusd.Debug {
		log.Debug(FmtResponse(resp, r.RemoteAddr))
	}

}

func (s *AuthService) SendReject(w radius.ResponseWriter, r *radius.Request, err error) {
	defer func() {
		if ret := recover(); ret != nil {
			err2, ok := ret.(error)
			if ok {
				log.Error2("radius write reject response error",
					zap.String("namespace", "radius"),
					zap.String("metrics", app.MetricsRadiusAuthDrop),
					zap.Error(err2),
				)
			}
		}
	}()

	var code = radius.CodeAccessReject
	var resp = r.Response(code)
	if err != nil {
		msg := err.Error()
		if len(msg) > 253 {
			msg = msg[:253]
		}
		_ = rfc2865.ReplyMessage_SetString(resp, msg)
	}

	common.Must(w.Write(resp))

	// debug message
	if app.GConfig().Radiusd.Debug {
		log.Info(FmtResponse(resp, r.RemoteAddr))
	}
}
