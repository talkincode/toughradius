package radiusd

import (
	"context"
	"strings"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	eap "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type AuthService struct {
	*RadiusService
	eapHelper *EAPAuthHelper
}

func NewAuthService(radiusService *RadiusService) *AuthService {
	return &AuthService{
		RadiusService: radiusService,
		eapHelper:     NewEAPAuthHelper(),
	}
}

func (s *AuthService) ServeRADIUS(w radius.ResponseWriter, r *radius.Request) {
	defer func() {
		if ret := recover(); ret != nil {
			switch ret.(type) {
			case error:
				err := ret.(error)
				zap.L().Error("radius auth error",
					zap.Error(err),
					zap.String("namespace", "radius"),
					zap.String("metrics", app.MetricsRadiusAuthDrop),
				)
				s.SendReject(w, r, err)
			case AuthError:
				err := ret.(AuthError)
				zap.L().Error("radius auth error",
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
		zap.S().Info(FmtRequest(r))
	}

	var eapMethod = s.GetEapMethod()
	var isEap bool
	if _, err := eap.ParseEAPMessage(r.Packet); err == nil {
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
		s.handleAuthError("validate_username", r, nil, nil, nil, false, callingStationID, ip,
			NewAuthError(app.MetricsRadiusRejectNotExists, "username is empty of client mac"))
	}

	nas, err := s.GetNas(ip, identifier)
	s.handleAuthError("load_nas", r, nil, nil, nil, false, username, ip, err)

	//  setup new packet secret
	r.Secret = []byte(nas.Secret)
	r.Packet.Secret = []byte(nas.Secret)

	if !isEap {
		s.handleAuthError("auth_rate_limit", r, nil, nas, nil, false, username, ip, s.CheckAuthRateLimit(username))
	}

	response := r.Response(radius.CodeAccessAccept)
	vendorReq := s.ParseVendor(r, nas.VendorCode)

	// ----------------------------------------------------------------------------------------------------
	// Fetch validate user
	isMacAuth := vendorReq.MacAddr == username
	user, err := s.GetValidUser(username, isMacAuth)
	s.handleAuthError("load_user", r, nil, nas, nil, isMacAuth, username, ip, err)

	// Note: Policy checks（online count、MACBind、VLANBind）now handled by plugin system
	// in AuthenticateUserWithPlugins() executed

	vendorReqForPlugin := &vendorparsers.VendorRequest{
		MacAddr: vendorReq.MacAddr,
		Vlanid1: vendorReq.Vlanid1,
		Vlanid2: vendorReq.Vlanid2,
	}

	sendAccept := func(isEapFlow bool) {
		s.ApplyAcceptEnhancers(user, nas, vendorReqForPlugin, response)

		if isEapFlow && s.eapHelper != nil {
			if err := s.eapHelper.SendEAPSuccess(w, r, response, nas.Secret); err != nil {
				zap.L().Error("send eap success failed",
					zap.String("namespace", "radius"),
					zap.Error(err),
				)
			}
			s.eapHelper.CleanupState(r)
		} else {
			s.SendAccept(w, r, response)
		}

		s.UpdateBind(user, vendorReq)
		s.UpdateUserLastOnline(user.Username)
		zap.L().Info("radius auth sucess",
			zap.String("namespace", "radius"),
			zap.String("username", username),
			zap.String("nasip", ip),
			zap.Bool("is_eap", isEapFlow),
			zap.String("result", "success"),
			zap.String("metrics", app.MetricsRadiusAccept),
		)
	}

	ctx := context.Background()

	if isEap && s.eapHelper != nil {
		handled, success, eapErr := s.eapHelper.HandleEAPAuthentication(
			w, r, user, nas, vendorReqForPlugin, response, eapMethod,
		)

		if eapErr != nil {
			zap.L().Warn("eap handling failed",
				zap.String("namespace", "radius"),
				zap.Error(eapErr),
			)
			_ = s.eapHelper.SendEAPFailure(w, r, nas.Secret, eapErr)
			s.eapHelper.CleanupState(r)
			return
		}

		if handled {
			if success {
				err = s.AuthenticateUserWithPlugins(ctx, r, response, user, nas, vendorReqForPlugin, isMacAuth, SkipPasswordValidation())
				if err != nil {
					_ = s.eapHelper.SendEAPFailure(w, r, nas.Secret, err)
					s.eapHelper.CleanupState(r)
					return
				}
				sendAccept(true)
			}
			return
		}
	}

	err = s.AuthenticateUserWithPlugins(ctx, r, response, user, nas, vendorReqForPlugin, isMacAuth)
	s.handleAuthError("plugin_auth", r, user, nas, vendorReqForPlugin, isMacAuth, username, ip, err)

	sendAccept(false)

	// s.CheckRequestSecret(r.Packet, []byte(nas.Secret))
}

func (s *AuthService) SendAccept(w radius.ResponseWriter, r *radius.Request, resp *radius.Packet) {
	defer func() {
		if ret := recover(); ret != nil {
			err2, ok := ret.(error)
			if ok {
				zap.L().Error("radius write accept error",
					zap.String("namespace", "radius"),
					zap.String("metrics", app.MetricsRadiusAuthDrop),
					zap.Error(err2),
				)
			}
		}
	}()

	common.Must(w.Write(resp))

	if s.eapHelper != nil {
		s.eapHelper.CleanupState(r)
	}

	if app.GConfig().Radiusd.Debug {
		zap.S().Debug(FmtResponse(resp, r.RemoteAddr))
	}

}

func (s *AuthService) SendReject(w radius.ResponseWriter, r *radius.Request, err error) {
	defer func() {
		if ret := recover(); ret != nil {
			err2, ok := ret.(error)
			if ok {
				zap.L().Error("radius write reject response error",
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

	if s.eapHelper != nil {
		s.eapHelper.CleanupState(r)
	}

	// debug message
	if app.GConfig().Radiusd.Debug {
		zap.S().Info(FmtResponse(resp, r.RemoteAddr))
	}
}

func (s *AuthService) handleAuthError(
	stage string,
	r *radius.Request,
	user interface{},
	nas *domain.NetNas,
	vendorReq *vendorparsers.VendorRequest,
	isMacAuth bool,
	username string,
	nasip string,
	err error,
) {
	if err == nil {
		return
	}

	var radiusUser *domain.RadiusUser
	if u, ok := user.(*domain.RadiusUser); ok {
		radiusUser = u
	}

	metadata := map[string]interface{}{
		"stage": stage,
	}
	if username != "" {
		metadata["username"] = username
	}
	if nasip != "" {
		metadata["nas_ip"] = nasip
	}

	authCtx := &auth.AuthContext{
		Request:       r,
		User:          radiusUser,
		Nas:           nas,
		VendorRequest: vendorReq,
		IsMacAuth:     isMacAuth,
		Metadata:      metadata,
	}

	for _, guard := range registry.GetAuthGuards() {
		if guardErr := guard.OnError(context.Background(), authCtx, stage, err); guardErr != nil {
			panic(guardErr)
		}
	}

	panic(err)
}
