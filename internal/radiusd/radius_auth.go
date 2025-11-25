package radiusd

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type AuthService struct {
	*RadiusService
	eapHelper               *EAPAuthHelper
	authPipeline            *AuthPipeline
	allowedEAPHandlers      map[string]struct{}
	allowedEAPHandlersOrder []string
}

func NewAuthService(radiusService *RadiusService) *AuthService {
	authService := &AuthService{
		RadiusService: radiusService,
	}
	allowed := authService.initAllowedEAPHandlers()
	authService.eapHelper = NewEAPAuthHelper(radiusService, allowed)
	authService.authPipeline = NewAuthPipeline()
	authService.registerDefaultStages()
	return authService
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

	if s.Config().Radiusd.Debug {
		zap.S().Info(FmtRequest(r))
	}

	s.ensurePipeline()
	pipelineCtx := NewAuthPipelineContext(s, w, r)
	defer func() {
		if pipelineCtx != nil && pipelineCtx.RateLimitChecked && pipelineCtx.Username != "" {
			s.ReleaseAuthRateLimit(pipelineCtx.Username)
		}
	}()
	if err := s.authPipeline.Execute(pipelineCtx); err != nil {
		s.handleAuthError("auth_pipeline", r, pipelineCtx.User, pipelineCtx.NAS, pipelineCtx.VendorRequestForPlugin, pipelineCtx.IsMacAuth, pipelineCtx.Username, pipelineCtx.RemoteIP, err)
	}
}

// Pipeline exposes the underlying auth pipeline for customization.
func (s *AuthService) Pipeline() *AuthPipeline {
	s.ensurePipeline()
	return s.authPipeline
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

	if s.Config().Radiusd.Debug {
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
	if s.Config().Radiusd.Debug {
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
		"stage":         stage,
		"config_mgr":    s.AppContext().ConfigMgr(),    // Add config manager for enhancers
		"profile_cache": s.AppContext().ProfileCache(), // Add profile cache for dynamic attribute resolution
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
