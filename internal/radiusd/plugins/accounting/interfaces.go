package accounting

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	vendorparserspkg "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"layeh.com/radius"
)

// AccountingContext is the shared accounting context
type AccountingContext struct {
	Context    context.Context
	Request    *radius.Request
	VendorReq  *vendorparserspkg.VendorRequest
	Username   string
	NAS        *domain.NetNas
	NASIP      string
	StatusType int // rfc2866: Start=1, Stop=2, InterimUpdate=3, AccountingOn=7, AccountingOff=8
}

// AccountingHandler defines the accounting handler interface
type AccountingHandler interface {
	// Name Returnshandlernames
	Name() string

	// CanHandle determines whether the handler can process this accounting request
	CanHandle(ctx *AccountingContext) bool

	// Handle HandleAccountingrequest
	Handle(ctx *AccountingContext) error
}
