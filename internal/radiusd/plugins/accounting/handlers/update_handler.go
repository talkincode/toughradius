package handlers

import (
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	vendorparserspkg "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"go.uber.org/zap"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
)

// UpdateHandler Accounting Update handler
type UpdateHandler struct {
	sessionRepo repository.SessionRepository
}

// NewUpdateHandler CreateAccounting Update handler
func NewUpdateHandler(sessionRepo repository.SessionRepository) *UpdateHandler {
	return &UpdateHandler{sessionRepo: sessionRepo}
}

func (h *UpdateHandler) Name() string {
	return "UpdateHandler"
}

func (h *UpdateHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	return ctx.StatusType == int(rfc2866.AcctStatusType_Value_InterimUpdate)
}

func (h *UpdateHandler) Handle(acctCtx *accounting.AccountingContext) error {
	vendorReq := acctCtx.VendorReq
	if vendorReq == nil {
		vendorReq = &vendorparserspkg.VendorRequest{}
	}

	// Build online session data
	online := buildOnlineFromRequest(acctCtx, vendorReq)

	// Update the online session record
	err := h.sessionRepo.Update(acctCtx.Context, &online)
	if err != nil {
		zap.L().Error("update radius online error",
			zap.String("namespace", "radius"),
			zap.String("username", acctCtx.Username),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// buildOnlineFromRequest Build online session data from request（used for Update）
func buildOnlineFromRequest(acctCtx *accounting.AccountingContext, vr *vendorparserspkg.VendorRequest) domain.RadiusOnline {
	r := acctCtx.Request
	acctInputOctets := int(rfc2866.AcctInputOctets_Get(r.Packet))
	acctInputGigawords := int(rfc2869.AcctInputGigawords_Get(r.Packet))
	acctOutputOctets := int(rfc2866.AcctOutputOctets_Get(r.Packet))
	acctOutputGigawords := int(rfc2869.AcctOutputGigawords_Get(r.Packet))

	return domain.RadiusOnline{
		AcctSessionId:     rfc2866.AcctSessionID_GetString(r.Packet),
		AcctSessionTime:   int(rfc2866.AcctSessionTime_Get(r.Packet)),
		AcctInputTotal:    int64(acctInputOctets) + int64(acctInputGigawords)*4*1024*1024*1024,
		AcctOutputTotal:   int64(acctOutputOctets) + int64(acctOutputGigawords)*4*1024*1024*1024,
		AcctInputPackets:  int(rfc2866.AcctInputPackets_Get(r.Packet)),
		AcctOutputPackets: int(rfc2866.AcctOutputPackets_Get(r.Packet)),
		LastUpdate:        time.Now(),
	}
}
