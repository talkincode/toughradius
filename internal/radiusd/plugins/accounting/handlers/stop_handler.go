package handlers

import (
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	vendorparserspkg "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"go.uber.org/zap"
	"layeh.com/radius/rfc2866"
)

// StopHandler handles Accounting-Stop packets and finalizes session
// accounting.
type StopHandler struct {
	sessionRepo    repository.SessionRepository
	accountingRepo repository.AccountingRepository
}

// NewStopHandler constructs a StopHandler with repositories used to update the
// final accounting counters and clear online-session state.
func NewStopHandler(
	sessionRepo repository.SessionRepository,
	accountingRepo repository.AccountingRepository,
) *StopHandler {
	return &StopHandler{
		sessionRepo:    sessionRepo,
		accountingRepo: accountingRepo,
	}
}

// Name returns the stable plugin name used by the accounting dispatcher.
func (h *StopHandler) Name() string {
	return "StopHandler"
}

// CanHandle reports whether the context represents an Accounting-Stop packet.
func (h *StopHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	return ctx.StatusType == int(rfc2866.AcctStatusType_Value_Stop)
}

// Handle updates the session's stop-time counters and removes the matching
// online-session row.
//
// When updating the accounting row fails, Handle logs the error and still tries
// to remove the online-session record so stale sessions do not remain online
// forever.
func (h *StopHandler) Handle(acctCtx *accounting.AccountingContext) error {
	vendorReq := acctCtx.VendorReq
	if vendorReq == nil {
		vendorReq = &vendorparserspkg.VendorRequest{}
	}

	// Build online session data
	online := buildOnlineFromRequest(acctCtx, vendorReq)
	sessionId := rfc2866.AcctSessionID_GetString(acctCtx.Request.Packet)

	// Update accounting record stop time
	acctRecord := domain.RadiusAccounting{
		AcctInputTotal:    online.AcctInputTotal,
		AcctOutputTotal:   online.AcctOutputTotal,
		AcctInputPackets:  online.AcctInputPackets,
		AcctOutputPackets: online.AcctOutputPackets,
		AcctSessionTime:   online.AcctSessionTime,
	}

	err := h.accountingRepo.UpdateStop(acctCtx.Context, sessionId, &acctRecord)
	if err != nil {
		zap.L().Error("update radius accounting stop error",
			zap.String("namespace", "radius"),
			zap.String("username", acctCtx.Username),
			zap.String("session_id", sessionId),
			zap.Error(err),
		)
	}

	// Delete the online session
	err = h.sessionRepo.Delete(acctCtx.Context, sessionId)
	if err != nil {
		zap.L().Error("delete radius online error",
			zap.String("namespace", "radius"),
			zap.String("username", acctCtx.Username),
			zap.String("session_id", sessionId),
			zap.Error(err),
		)
		return err
	}

	return nil
}
