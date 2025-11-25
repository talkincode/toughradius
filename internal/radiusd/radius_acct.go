package radiusd

import (
	"context"
	"strings"

	"github.com/talkincode/toughradius/v9/internal/app"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	vendorparserspkg "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

// Accounting service
type AcctService struct {
	*RadiusService
}

func NewAcctService(radiusService *RadiusService) *AcctService {
	return &AcctService{RadiusService: radiusService}
}

func (s *AcctService) ServeRADIUS(w radius.ResponseWriter, r *radius.Request) {
	// Recover from unexpected panics only (programming errors)
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
			zap.L().Error("radius accounting unexpected panic",
				zap.Error(err),
				zap.String("namespace", "radius"),
				zap.String("metrics", app.MetricsRadiusAcctDrop),
				zap.Stack("stacktrace"),
			)
		}
	}()

	if r == nil {
		return
	}

	if s.Config().Radiusd.Debug {
		zap.S().Debug(FmtRequest(r))
	}

	// NAS Access check
	raddrstr := r.RemoteAddr.String()
	nasrip := raddrstr[:strings.Index(raddrstr, ":")]
	var identifier = rfc2865.NASIdentifier_GetString(r.Packet)

	nas, err := s.GetNas(nasrip, identifier)
	if err != nil {
		s.logAcctError("nas_lookup", nasrip, "", err)
		return
	}

	// Reset packet secret
	r.Secret = []byte(nas.Secret)
	r.Secret = []byte(nas.Secret) //nolint:staticcheck

	statusType := rfc2866.AcctStatusType_Get(r.Packet)

	// UsernameCheck
	var username string
	if statusType != rfc2866.AcctStatusType_Value_AccountingOn &&
		statusType != rfc2866.AcctStatusType_Value_AccountingOff {
		username = rfc2865.UserName_GetString(r.Packet)
		if username == "" {
			s.logAcctError("validate_username", nasrip, "", radiuserrors.NewAcctUsernameEmptyError())
			return
		}
	}

	defer s.ReleaseAuthRateLimit(username)

	// s.CheckRequestSecret(r.Packet, []byte(nas.Secret))

	vendorReq := s.ParseVendor(r, nas.VendorCode)

	s.SendResponse(w, r)

	zap.S().Info("radius accounting",
		zap.String("namespace", "radius"),
		zap.String("metrics", app.MetricsRadiusAccounting),
	)

	// async process accounting with back-pressure aware submit
	task := func() {
		vendorReqForPlugin := &vendorparserspkg.VendorRequest{
			MacAddr: vendorReq.MacAddr,
			Vlanid1: vendorReq.Vlanid1,
			Vlanid2: vendorReq.Vlanid2,
		}

		ctx := context.Background()
		err := s.HandleAccountingWithPlugins(ctx, r, vendorReqForPlugin, username, nas, nasrip)
		if err != nil {
			zap.L().Error("accounting plugin processing error",
				zap.String("namespace", "radius"),
				zap.String("username", username),
				zap.Int("status_type", int(statusType)),
				zap.Error(err),
			)
		}
	}

	if err := s.TaskPool.Submit(task); err != nil {
		zap.L().Warn("accounting task pool saturated, running fallback goroutine",
			zap.String("namespace", "radius"),
			zap.String("metrics", app.MetricsRadiusAcctDrop),
			zap.Error(err),
		)
		go task()
	}
}

// logAcctError logs accounting errors with appropriate metrics.
func (s *AcctService) logAcctError(stage, nasip, username string, err error) {
	metricsKey := app.MetricsRadiusAcctDrop
	if radiusErr, ok := radiuserrors.GetRadiusError(err); ok {
		metricsKey = radiusErr.MetricsKey()
	}

	fields := []zap.Field{
		zap.Error(err),
		zap.String("namespace", "radius"),
		zap.String("metrics", metricsKey),
		zap.String("stage", stage),
	}
	if nasip != "" {
		fields = append(fields, zap.String("nasip", nasip))
	}
	if username != "" {
		fields = append(fields, zap.String("username", username))
	}

	zap.L().Error("radius accounting error", fields...)
}

func (s *AcctService) SendResponse(w radius.ResponseWriter, r *radius.Request) {
	resp := r.Response(radius.CodeAccountingResponse)
	if err := w.Write(resp); err != nil {
		zap.L().Error("radius accounting response error",
			zap.Error(err),
			zap.String("namespace", "radius"),
			zap.String("metrics", app.MetricsRadiusAcctDrop),
		)
		return
	}

	if s.Config().Radiusd.Debug {
		zap.S().Debug(FmtResponse(resp, r.RemoteAddr))
	}
}
