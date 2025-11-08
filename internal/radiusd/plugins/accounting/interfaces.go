package accounting

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	vendorparserspkg "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"layeh.com/radius"
)

// AccountingContext 计费上下文（统一命名）
type AccountingContext struct {
	Context    context.Context
	Request    *radius.Request
	VendorReq  *vendorparserspkg.VendorRequest
	Username   string
	NAS        *domain.NetNas
	NASIP      string
	StatusType int // rfc2866: Start=1, Stop=2, InterimUpdate=3, AccountingOn=7, AccountingOff=8
}

// AccountingHandler 计费处理器接口
type AccountingHandler interface {
	// Name 返回处理器名称
	Name() string

	// CanHandle 判断是否能处理该计费请求
	CanHandle(ctx *AccountingContext) bool

	// Handle 处理计费请求
	Handle(ctx *AccountingContext) error
}
