package radiusd

import (
	"context"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	vendorparserspkg "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"go.uber.org/zap"
	"layeh.com/radius"
)

// HandleAccountingWithPlugins 使用插件系统处理计费请求
func (s *AcctService) HandleAccountingWithPlugins(
	ctx context.Context,
	r *radius.Request,
	vendorReq *vendorparserspkg.VendorRequest,
	username string,
	nas *domain.NetNas,
	nasIP string,
) error {
	// 获取Accounting-Status-Type
	statusTypeAttr := r.Packet.Get(40) // Acct-Status-Type
	if statusTypeAttr == nil {
		return fmt.Errorf("missing Acct-Status-Type attribute")
	}

	// statusType已经在ServeRADIUS中获取，这里简化为直接从statusTypeAttr提取
	// rfc2866的Value常量: Start=1, Stop=2, InterimUpdate=3, AccountingOn=7, AccountingOff=8
	statusType := statusTypeAttr[0]

	// 构建AccountingContext
	acctCtx := &accounting.AccountingContext{
		Context:    ctx,
		Request:    r,
		VendorReq:  vendorReq,
		Username:   username,
		NAS:        nas,
		NASIP:      nasIP,
		StatusType: int(statusType),
	}

	// 获取注册的Accounting Handler
	handlers := registry.GetAccountingHandlers()
	if len(handlers) == 0 {
		return fmt.Errorf("no accounting handlers registered")
	}

	// 遍历handlers，找到能处理该StatusType的handler
	for _, handler := range handlers {
		if handler.CanHandle(acctCtx) {
			err := handler.Handle(acctCtx)
			if err != nil {
				zap.L().Error("accounting handler failed",
					zap.String("namespace", "radius"),
					zap.String("handler", handler.Name()),
					zap.String("username", username),
					zap.Int("status_type", int(statusType)),
					zap.Error(err),
				)
				return err
			}

			// 记录成功处理的metrics
			switch statusType {
			case 1: // Start
				zap.L().Info("radius accounting start",
					zap.String("namespace", "radius"),
					zap.String("metrics", app.MetricsRadiusOline),
				)
			case 2: // Stop
				zap.L().Info("radius accounting stop",
					zap.String("namespace", "radius"),
					zap.String("metrics", app.MetricsRadiusOffline),
				)
			}

			return nil
		}
	}

	return fmt.Errorf("no handler found for status type %d", statusType)
}
