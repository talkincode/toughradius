package radiusd

import (
	stderrs "errors"
	"fmt"
	"net"
	"strings"

	"github.com/talkincode/toughradius/v9/internal/app"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	eap "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	StageRequestMetadata = "request_metadata"
	StageNasLookup       = "nas_lookup"
	StageRateLimit       = "auth_rate_limit"
	StageVendorParsing   = "vendor_parsing"
	StageLoadUser        = "load_user"
	StageEAPDispatch     = "eap_dispatch"
	StagePluginAuth      = "plugin_auth"
)

func (s *AuthService) registerDefaultStages() {
	stages := []AuthPipelineStage{
		newStage(StageRequestMetadata, s.stageRequestMetadata),
		newStage(StageNasLookup, s.stageNasLookup),
		newStage(StageRateLimit, s.stageRateLimit),
		newStage(StageVendorParsing, s.stageVendorParsing),
		newStage(StageLoadUser, s.stageLoadUser),
		newStage(StageEAPDispatch, s.stageEAPDispatch),
		newStage(StagePluginAuth, s.stagePluginAuth),
	}

	for _, stage := range stages {
		s.authPipeline.Use(stage)
	}
}

func (s *AuthService) stageRequestMetadata(ctx *AuthPipelineContext) error {
	r := ctx.Request

	preferredMethod := s.resolveEapMethod(s.GetEapMethod())
	ctx.EAPMethod = preferredMethod

	if _, err := eap.ParseEAPMessage(r.Packet); err == nil {
		ctx.IsEAP = true
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr.String())
	if err != nil {
		ctx.RemoteIP = r.RemoteAddr.String()
	} else {
		ctx.RemoteIP = host
	}

	ctx.NasIdentifier = rfc2865.NASIdentifier_GetString(r.Packet)
	ctx.Username = rfc2865.UserName_GetString(r.Packet)
	ctx.CallingStationID = rfc2865.CallingStationID_GetString(r.Packet)

	if ctx.Username == "" {
		return radiuserrors.NewAuthErrorWithStage(
			app.MetricsRadiusRejectNotExists,
			"username is empty of client mac",
			StageRequestMetadata,
		)
	}

	return nil
}

func (s *AuthService) stageNasLookup(ctx *AuthPipelineContext) error {
	nas, err := s.GetNas(ctx.RemoteIP, ctx.NasIdentifier)
	if err != nil {
		return err
	}
	ctx.NAS = nas

	if nas != nil {
		secret := []byte(nas.Secret)
		ctx.Request.Secret = secret
		ctx.Request.Secret = secret //nolint:staticcheck
		ctx.Response = ctx.Request.Response(radius.CodeAccessAccept)
	}

	return nil
}

func (s *AuthService) stageRateLimit(ctx *AuthPipelineContext) error {
	if ctx.IsEAP {
		return nil
	}
	if err := s.CheckAuthRateLimit(ctx.Username); err != nil {
		return err
	}
	ctx.RateLimitChecked = true
	return nil
}

func (s *AuthService) stageVendorParsing(ctx *AuthPipelineContext) error {
	if ctx.NAS == nil {
		return fmt.Errorf("nas should not be nil before vendor parsing")
	}
	vendorReq := s.ParseVendor(ctx.Request, ctx.NAS.VendorCode)
	ctx.VendorRequest = vendorReq

	ctx.IsMacAuth = vendorReq.MacAddr != "" && vendorReq.MacAddr == ctx.Username

	ctx.VendorRequestForPlugin = &vendorparsers.VendorRequest{
		MacAddr: vendorReq.MacAddr,
		Vlanid1: vendorReq.Vlanid1,
		Vlanid2: vendorReq.Vlanid2,
	}
	return nil
}

func (s *AuthService) stageLoadUser(ctx *AuthPipelineContext) error {
	user, err := s.GetValidUser(ctx.Username, ctx.IsMacAuth)
	if err != nil {
		return err
	}
	ctx.User = user
	return nil
}

func (s *AuthService) stageEAPDispatch(ctx *AuthPipelineContext) error {
	if !ctx.IsEAP || s.eapHelper == nil {
		return nil
	}

	handled, success, eapErr := s.eapHelper.HandleEAPAuthentication(
		ctx.Writer,
		ctx.Request,
		ctx.User,
		ctx.NAS,
		ctx.VendorRequestForPlugin,
		ctx.Response,
		ctx.EAPMethod,
	)

	if eapErr != nil {
		rejectErr := mapEAPDispatchError(eapErr)
		s.logEAPFailure(ctx, rejectErr)
		_ = s.eapHelper.SendEAPFailure(ctx.Writer, ctx.Request, ctx.NAS.Secret, rejectErr)
		s.eapHelper.CleanupState(ctx.Request)
		ctx.Stop()
		return nil
	}

	if handled {
		if success {
			err := s.AuthenticateUserWithPlugins(ctx.Context, ctx.Request, ctx.Response, ctx.User, ctx.NAS, ctx.VendorRequestForPlugin, ctx.IsMacAuth, SkipPasswordValidation())
			if err != nil {
				rejectErr := mapEAPDispatchError(err)
				s.logEAPFailure(ctx, rejectErr)
				_ = s.eapHelper.SendEAPFailure(ctx.Writer, ctx.Request, ctx.NAS.Secret, rejectErr)
				s.eapHelper.CleanupState(ctx.Request)
				ctx.Stop()
				return nil
			}
			s.sendAcceptResponse(ctx, true)
		}
		ctx.Stop()
	}

	return nil
}

func mapEAPDispatchError(err error) error {
	if err == nil {
		return nil
	}

	if authErr, ok := radiuserrors.GetAuthError(err); ok {
		cloned := *authErr
		if cloned.ErrorStage == "" {
			cloned.ErrorStage = StageEAPDispatch
		}
		return &cloned
	}

	switch {
	case stderrs.Is(err, eap.ErrPasswordMismatch):
		return newEAPDispatchAuthError(app.MetricsRadiusRejectPasswdError, "eap password validation failed", err)
	case stderrs.Is(err, eap.ErrTLSIdentityMismatch):
		return newEAPDispatchAuthError(app.MetricsRadiusRejectUnauthorized, "eap-tls certificate identity mismatch", err)
	case stderrs.Is(err, eap.ErrTLSNoIdentity):
		return newEAPDispatchAuthError(app.MetricsRadiusRejectUnauthorized, "eap-tls certificate identity missing", err)
	case stderrs.Is(err, eap.ErrTLSNotConfigured):
		return newEAPDispatchAuthError(app.MetricsRadiusRejectOther, "eap-tls trust configuration missing", err)
	case stderrs.Is(err, eap.ErrTLSHandshakeFailed):
		return newEAPDispatchAuthError(app.MetricsRadiusRejectOther, "eap-tls handshake failed", err)
	case stderrs.Is(err, eap.ErrTLSUnexpectedFragment):
		return newEAPDispatchAuthError(app.MetricsRadiusRejectOther, "eap-tls fragmentation exchange failed", err)
	case stderrs.Is(err, eap.ErrPEAPInnerProtocol):
		return newEAPDispatchAuthError(app.MetricsRadiusRejectOther, "peap inner eap-mschapv2 protocol violation", err)
	case stderrs.Is(err, eap.ErrPEAPInnerNotImplemented):
		return newEAPDispatchAuthError(app.MetricsRadiusRejectOther, "peap inner eap method unavailable", err)
	case stderrs.Is(err, eap.ErrStateNotFound):
		return newEAPDispatchAuthError(app.MetricsRadiusRejectOther, "eap session state not found", err)
	case stderrs.Is(err, eap.ErrUnsupportedEAPType):
		return newEAPDispatchAuthError(app.MetricsRadiusRejectOther, "eap type not supported", err)
	case stderrs.Is(err, eap.ErrInvalidEAPMessage):
		return newEAPDispatchAuthError(app.MetricsRadiusRejectOther, "invalid eap message", err)
	default:
		return newEAPDispatchAuthError(app.MetricsRadiusRejectOther, "eap authentication failed", err)
	}
}

func newEAPDispatchAuthError(metricsType, message string, cause error) error {
	return &radiuserrors.AuthError{
		MetricsType: metricsType,
		Message:     message,
		ErrorStage:  StageEAPDispatch,
		Cause:       cause,
	}
}

func (s *AuthService) logEAPFailure(ctx *AuthPipelineContext, err error) {
	metricsKey := app.MetricsRadiusRejectOther
	stage := StageEAPDispatch
	if radiusErr, ok := radiuserrors.GetRadiusError(err); ok {
		metricsKey = radiusErr.MetricsKey()
		if radiusErr.Stage() != "" {
			stage = radiusErr.Stage()
		}
	}
	app.IncRadiusMetric(metricsKey)

	zap.L().Warn("radius eap auth rejected",
		zap.String("namespace", "radius"),
		zap.String("stage", stage),
		zap.String("username", ctx.Username),
		zap.String("nasip", ctx.RemoteIP),
		zap.String("metrics", metricsKey),
		zap.String("reason", safeEAPFailureReason(err)),
	)
}

func safeEAPFailureReason(err error) string {
	if authErr, ok := radiuserrors.GetAuthError(err); ok {
		return authErr.Message
	}
	if err == nil {
		return "eap authentication failed"
	}
	return "eap authentication failed"
}

func (s *AuthService) stagePluginAuth(ctx *AuthPipelineContext) error {
	if ctx.IsStopped() {
		return nil
	}

	err := s.AuthenticateUserWithPlugins(ctx.Context, ctx.Request, ctx.Response, ctx.User, ctx.NAS, ctx.VendorRequestForPlugin, ctx.IsMacAuth)
	if err != nil {
		return err
	}

	s.sendAcceptResponse(ctx, false)
	ctx.Stop()
	return nil
}

func (s *AuthService) sendAcceptResponse(ctx *AuthPipelineContext, isEapFlow bool) {
	vendorPlugin := ctx.VendorRequestForPlugin
	if vendorPlugin == nil {
		vendorPlugin = &vendorparsers.VendorRequest{}
	}

	if ctx.NAS == nil || ctx.User == nil {
		zap.L().Warn("skip accept response due to missing context",
			zap.String("namespace", "radius"),
			zap.Bool("is_eap", isEapFlow),
		)
		return
	}

	s.ApplyAcceptEnhancers(ctx.User, ctx.NAS, vendorPlugin, ctx.Response)

	if isEapFlow && s.eapHelper != nil {
		if err := s.eapHelper.SendEAPSuccess(ctx.Writer, ctx.Request, ctx.Response, ctx.NAS.Secret); err != nil {
			zap.L().Error("send eap success failed",
				zap.String("namespace", "radius"),
				zap.Error(err),
			)
		}
		s.eapHelper.CleanupState(ctx.Request)
	} else {
		s.SendAccept(ctx.Writer, ctx.Request, ctx.Response)
	}

	vendorReq := ctx.VendorRequest
	if vendorReq == nil {
		vendorReq = &VendorRequest{}
	}

	if ctx.User != nil {
		s.UpdateBind(ctx.User, vendorReq)
		s.UpdateUserLastOnline(ctx.User.Username)
	}

	zap.L().Info("radius auth success",
		zap.String("namespace", "radius"),
		zap.String("username", ctx.Username),
		zap.String("nasip", ctx.RemoteIP),
		zap.Bool("is_eap", isEapFlow),
		zap.String("result", "success"),
		zap.String("metrics", app.MetricsRadiusAccept),
	)
}

func (s *AuthService) resolveEapMethod(preferred string) string {
	method := strings.TrimSpace(strings.ToLower(preferred))
	if method == "" {
		method = "eap-md5"
	}
	if len(s.allowedEAPHandlers) == 0 {
		return method
	}
	if _, ok := s.allowedEAPHandlers[method]; ok {
		return method
	}
	for _, candidate := range s.allowedEAPHandlersOrder {
		if _, ok := s.allowedEAPHandlers[candidate]; ok {
			zap.L().Warn("preferred EAP method disabled, falling back",
				zap.String("namespace", "radius"),
				zap.String("preferred", method),
				zap.String("fallback", candidate),
			)
			return candidate
		}
	}
	for candidate := range s.allowedEAPHandlers {
		zap.L().Warn("preferred EAP method disabled, falling back",
			zap.String("namespace", "radius"),
			zap.String("preferred", method),
			zap.String("fallback", candidate),
		)
		return candidate
	}
	return method
}

func (s *AuthService) ensurePipeline() {
	if s.authPipeline != nil {
		return
	}
	s.authPipeline = NewAuthPipeline()
	s.registerDefaultStages()
}

func (s *AuthService) buildAllowedEAPHandlers() []string {
	appCtx := s.AppContext()
	if appCtx == nil {
		return nil
	}
	cfgMgr := appCtx.ConfigMgr()
	if cfgMgr == nil {
		return nil
	}

	raw := strings.TrimSpace(cfgMgr.GetString("radius", "EapEnabledHandlers"))
	if raw == "" || raw == "*" {
		return nil
	}

	parts := strings.Split(raw, ",")
	ordered := make([]string, 0, len(parts))
	seen := make(map[string]struct{})

	for _, part := range parts {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" {
			continue
		}
		if name == "*" {
			return nil
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		ordered = append(ordered, name)
	}

	return ordered
}

func (s *AuthService) initAllowedEAPHandlers() []string {
	allowed := s.buildAllowedEAPHandlers()
	if len(allowed) == 0 {
		s.allowedEAPHandlers = nil
		s.allowedEAPHandlersOrder = nil
		return nil
	}

	s.allowedEAPHandlers = make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		s.allowedEAPHandlers[name] = struct{}{}
	}
	s.allowedEAPHandlersOrder = allowed
	return allowed
}
