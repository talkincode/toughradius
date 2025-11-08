package radiusd

import (
	"context"
	"errors"
	"strings"

	"github.com/talkincode/toughradius/v9/internal/app"
	vendorparserspkg "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/pkg/common"
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
				zap.S().Error("radius accounting error",
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
		zap.S().Debug(FmtRequest(r))
	}

	// NAS 接入检查
	raddrstr := r.RemoteAddr.String()
	nasrip := raddrstr[:strings.Index(raddrstr, ":")]
	var identifier = rfc2865.NASIdentifier_GetString(r.Packet)
	nas, err := s.GetNas(nasrip, identifier)
	common.Must(err)

	// 重新设置数据报文秘钥
	r.Secret = []byte(nas.Secret)
	r.Packet.Secret = []byte(nas.Secret)

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

	// s.CheckRequestSecret(r.Packet, []byte(nas.Secret))

	vendorReq := s.ParseVendor(r, nas.VendorCode)

	s.SendResponse(w, r)

	zap.S().Info("radius accounting",
		zap.String("namespace", "radius"),
		zap.String("metrics", app.MetricsRadiusAccounting),
	)

	// async process accounting
	common.Must(s.TaskPool.Submit(func() {
		// 将vendorReq转换为vendorparsers.VendorRequest用于插件
		vendorReqForPlugin := &vendorparserspkg.VendorRequest{
			MacAddr: vendorReq.MacAddr,
			Vlanid1: vendorReq.Vlanid1,
			Vlanid2: vendorReq.Vlanid2,
		}

		// 使用插件系统处理计费请求
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
	}))
}

func (s *AcctService) SendResponse(w radius.ResponseWriter, r *radius.Request) {
	resp := r.Response(radius.CodeAccountingResponse)
	err := w.Write(resp)
	if err != nil {
		zap.L().Error("radius accounting response error",
			zap.Error(err),
			zap.String("namespace", "radius"),
			zap.String("metrics", app.MetricsRadiusAcctDrop),
		)
		return
	}

	if app.GConfig().Radiusd.Debug {
		zap.S().Debug(FmtResponse(resp, r.RemoteAddr))
	}

}
