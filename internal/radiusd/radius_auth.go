package radiusd

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
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
	// Recover from unexpected panics only (programming errors)
	// Normal errors should be handled via error return values
	defer func() {
		if ret := recover(); ret != nil {
			var err error
			switch v := ret.(type) {
			case error:
				err = v
			case string:
				err = radiuserrors.NewError(v)
			default:
				err = radiuserrors.NewError("unknown panic")
			}
			zap.L().Error("radius auth unexpected panic",
				zap.Error(err),
				zap.String("namespace", "radius"),
				zap.String("metrics", app.MetricsRadiusAuthDrop),
				zap.Stack("stacktrace"),
			)
			s.SendReject(w, r, err)
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
		// Process error through guards and log appropriately
		finalErr := s.processAuthError("auth_pipeline", r, pipelineCtx.User, pipelineCtx.NAS,
			pipelineCtx.VendorRequestForPlugin, pipelineCtx.IsMacAuth,
			pipelineCtx.Username, pipelineCtx.RemoteIP, err)
		if finalErr != nil {
			s.logAndReject(w, r, finalErr)
		}
	}
}

// Pipeline exposes the underlying auth pipeline for customization.
func (s *AuthService) Pipeline() *AuthPipeline {
	s.ensurePipeline()
	return s.authPipeline
}

func (s *AuthService) SendAccept(w radius.ResponseWriter, r *radius.Request, resp *radius.Packet) {
	if err := w.Write(resp); err != nil {
		zap.L().Error("radius write accept error",
			zap.String("namespace", "radius"),
			zap.String("metrics", app.MetricsRadiusAuthDrop),
			zap.Error(err),
		)
		return
	}

	if s.eapHelper != nil {
		s.eapHelper.CleanupState(r)
	}

	if s.Config().Radiusd.Debug {
		zap.S().Debug(FmtResponse(resp, r.RemoteAddr))
	}
}

func (s *AuthService) SendReject(w radius.ResponseWriter, r *radius.Request, err error) {
	var code = radius.CodeAccessReject
	var resp = r.Response(code)
	if err != nil {
		msg := err.Error()
		if len(msg) > 253 {
			msg = msg[:253]
		}
		_ = rfc2865.ReplyMessage_SetString(resp, msg)
	}

	if writeErr := w.Write(resp); writeErr != nil {
		zap.L().Error("radius write reject response error",
			zap.String("namespace", "radius"),
			zap.String("metrics", app.MetricsRadiusAuthDrop),
			zap.Error(writeErr),
		)
	}

	if s.eapHelper != nil {
		s.eapHelper.CleanupState(r)
	}

	// debug message
	if s.Config().Radiusd.Debug {
		zap.S().Info(FmtResponse(resp, r.RemoteAddr))
	}
}

// logAndReject logs the error with appropriate metrics and sends reject response.
func (s *AuthService) logAndReject(w radius.ResponseWriter, r *radius.Request, err error) {
	metricsKey := app.MetricsRadiusAuthDrop
	if radiusErr, ok := radiuserrors.GetRadiusError(err); ok {
		metricsKey = radiusErr.MetricsKey()
	}

	zap.L().Error("radius auth error",
		zap.Error(err),
		zap.String("namespace", "radius"),
		zap.String("metrics", metricsKey),
	)

	s.SendReject(w, r, err)
}

// processAuthError processes authentication errors through registered guards.
// It returns the final error after all guards have been consulted.
// This replaces the old handleAuthError which used panic for flow control.
//
// Parameters:
//   - stage: The pipeline stage where the error occurred
//   - r: The RADIUS request
//   - user: The user being authenticated (may be nil)
//   - nas: The NAS device (may be nil)
//   - vendorReq: Vendor-specific request data
//   - isMacAuth: Whether this is MAC authentication
//   - username: The username (for logging)
//   - nasip: The NAS IP (for logging)
//   - err: The original error
//
// Returns:
//   - error: The final error after guard processing, or nil if suppressed
func (s *AuthService) processAuthError(
	stage string,
	r *radius.Request,
	user interface{},
	nas *domain.NetNas,
	vendorReq *vendorparsers.VendorRequest,
	isMacAuth bool,
	username string,
	nasip string,
	err error,
) error {
	if err == nil {
		return nil
	}

	var radiusUser *domain.RadiusUser
	if u, ok := user.(*domain.RadiusUser); ok {
		radiusUser = u
	}

	metadata := map[string]interface{}{
		"stage": stage,
	}
	if appCtx := s.AppContext(); appCtx != nil {
		metadata["config_mgr"] = appCtx.ConfigMgr()
		metadata["profile_cache"] = appCtx.ProfileCache()
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

	ctx := context.Background()
	currentErr := err

	// Process through all guards
	for _, guard := range registry.GetAuthGuards() {
		// Try new interface first
		if result := guard.OnAuthError(ctx, authCtx, stage, currentErr); result != nil {
			switch result.Action {
			case auth.GuardActionStop:
				// Use the error from guard and stop processing
				if result.Err != nil {
					return result.Err
				}
				return currentErr
			case auth.GuardActionSuppress:
				// Error is suppressed, treat as success
				return nil
			case auth.GuardActionContinue:
				// Update error if guard modified it
				if result.Err != nil {
					currentErr = result.Err
				}
				continue
			}
		}

		// Fallback to old interface for backward compatibility
		if guardErr := guard.OnError(ctx, authCtx, stage, currentErr); guardErr != nil {
			// Old behavior: guard returns error means replace current error
			currentErr = guardErr
		}
	}

	return currentErr
}
