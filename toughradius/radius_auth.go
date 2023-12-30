package toughradius

import (
	"fmt"
	"strings"

	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
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

	var eapMethod = s.GetEapMethod()
	var isEap = false
	eapmsg, err := parseEAPMessage(r.Packet)
	if err == nil {
		isEap = true
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

	vpe, err := s.GetNas(ip, identifier)
	s.CheckRadAuthError(username, ip, err)

	//  setup new packet secret
	r.Secret = []byte(vpe.Secret)
	r.Packet.Secret = []byte(vpe.Secret)

	if !isEap {
		s.CheckRadAuthError(username, ip, s.CheckAuthRateLimit(username))
	}

	if isEap && eapmsg.Code == EAPCodeResponse && eapmsg.Type == EAPTypeIdentity {
		switch eapMethod {
		case EapMd5Method:
			// 发送EAP-Request/MD5-Challenge消息
			err = s.sendEapMD5ChallengeRequest(w, r, vpe.Secret)
			if err != nil {
				s.CheckRadAuthError(username, ip, fmt.Errorf("eap: send eap request error: %s", err))
			}
			return
		case EapMschapv2Method:
			// 发送EAP-Request/MSCHAPv2-Challenge消息
			err = s.sendEapMsChapV2Request(w, r, vpe.Secret)
			if err != nil {
				s.CheckRadAuthError(username, ip, fmt.Errorf("eap: send eap request error: %s", err))
			}
			return

		}
	}

	response := r.Response(radius.CodeAccessAccept)
	vendorReq := s.ParseVendor(r, vpe.VendorCode)

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

	sendAccept := func() {
		if isEap {
			eapSuccess := NewEAPSuccess(r.Identifier)
			// 设置EAP-Message属性
			rfc2869.EAPMessage_Set(response, eapSuccess.Serialize())
			rfc2869.MessageAuthenticator_Set(response, make([]byte, 16))
			authenticator := generateMessageAuthenticator(response, vpe.Secret)
			// 设置Message-Authenticator属性
			rfc2869.MessageAuthenticator_Set(response, authenticator)
		}
		s.AcceptAcceptConfig(user, vpe.VendorCode, response)
		s.SendAccept(w, r, response)
		s.UpdateBind(user, vendorReq)
		log.Info2("radius auth sucess",
			zap.String("namespace", "radius"),
			zap.String("username", username),
			zap.String("nasip", ip),
			zap.String("result", "success"),
			zap.String("metrics", app.MetricsRadiusAccept),
		)
	}

	if isEap && eapmsg.Code == EAPCodeResponse {
		switch eapmsg.Type {
		case EAPTypeMD5Challenge:
			stateid := rfc2865.State_GetString(r.Packet)
			eapState, err := s.GetEapState(stateid)
			if err != nil {
				s.SendEapFailureReject(w, r, vpe.Secret, fmt.Errorf("eap: get eap state error"))
				return
			}
			localpwd, err := s.GetLocalPassword(user, isMacAuth)
			if err != nil {
				s.SendEapFailureReject(w, r, vpe.Secret, fmt.Errorf("eap: get local password error: %s", err))
				return
			}
			if !s.verifyEapMD5Response(eapmsg.Identifier, localpwd, eapState.Challenge, eapmsg.Data) {
				s.SendEapFailureReject(w, r, vpe.Secret, fmt.Errorf("eap: verify eap md5 response error"))
				return
			}
			eapState.Success = true
			sendAccept()

		case EAPTypeMSCHAPv2:
			opcode, err := parseEAPMSCHAPv2OpCode(r.Packet)
			if err != nil {
				s.SendEapFailureReject(w, r, vpe.Secret, fmt.Errorf("eap: parse eap mschapv2 opcode error"))
				return
			}

			switch opcode {
			case MSCHAPv2Response:
				stateid := rfc2865.State_GetString(r.Packet)
				eapState, err := s.GetEapState(stateid)
				if err != nil {
					s.SendEapFailureReject(w, r, vpe.Secret, fmt.Errorf("eap: get eap state error"))
					return
				}
				localpwd, err := s.GetLocalPassword(user, isMacAuth)
				if err != nil {
					s.SendEapFailureReject(w, r, vpe.Secret, fmt.Errorf("eap: get local password error: %s", err))
					return
				}
				eapMv2Message, err := ParseEAPMSCHAPv2Response(r.Packet)
				if err != nil {
					s.SendEapFailureReject(w, r, vpe.Secret, fmt.Errorf("eap: parse eap mschapv2 response error"))
					return
				}
				err = s.CheckMsChapV2Password(
					username, localpwd, eapState.Challenge,
					eapMv2Message.Identifier,
					eapMv2Message.PeerChallenge[:],
					eapMv2Message.Response[:],
					response,
				)
				if err != nil {
					s.SendEapFailureReject(w, r, vpe.Secret, fmt.Errorf("eap: verify mschapv2 response error"))
					return
				}
				eapState.Success = true
				sendAccept()
			case MSCHAPv2Success:
				sendAccept()
			}
		}

	} else {
		localpwd, err := s.GetLocalPassword(user, isMacAuth)
		if err != nil {
			s.CheckRadAuthError(username, ip, fmt.Errorf("user local password error: %s", err))
		}
		s.CheckRadAuthError(username, ip, s.CheckPassword(r, user.Username, localpwd, response, isMacAuth))
		sendAccept()
	}

	// s.CheckRequestSecret(r.Packet, []byte(vpe.Secret))
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

	state := rfc2865.State_GetString(r.Packet)
	if state != "" {
		s.DeleteEapState(state)
	}

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

	_ = w.Write(resp)

	state := rfc2865.State_GetString(r.Packet)
	if state != "" {
		s.DeleteEapState(state)
	}

	// debug message
	if app.GConfig().Radiusd.Debug {
		log.Info(FmtResponse(resp, r.RemoteAddr))
	}
}

func (s *AuthService) SendEapFailureReject(w radius.ResponseWriter, r *radius.Request, secret string, err error) {
	defer func() {
		if ret := recover(); ret != nil {
			err2, ok := ret.(error)
			if ok {
				log.Error2("radius write eap reject response error",
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

	eapFailure := NewEAPFailure(r.Identifier)
	rfc2869.EAPMessage_Set(resp, eapFailure.Serialize())
	rfc2869.MessageAuthenticator_Set(resp, make([]byte, 16))
	authenticator := generateMessageAuthenticator(resp, secret)
	rfc2869.MessageAuthenticator_Set(resp, authenticator)

	_ = w.Write(resp)

	state := rfc2865.State_GetString(r.Packet)
	if state != "" {
		s.DeleteEapState(state)
	}

	// debug message
	if app.GConfig().Radiusd.Debug {
		log.Info(FmtResponse(resp, r.RemoteAddr))
	}
}
