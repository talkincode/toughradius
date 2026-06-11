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

// UpdateHandler handles Accounting-Interim-Update packets and refreshes
// counters on the active online-session row.
//
// Interim updates mutate only live-session state; they do not append historical
// accounting rows.
type UpdateHandler struct {
	sessionRepo repository.SessionRepository
}

// NewUpdateHandler constructs an UpdateHandler with the session repository used
// for interim counter updates.
func NewUpdateHandler(sessionRepo repository.SessionRepository) *UpdateHandler {
	return &UpdateHandler{sessionRepo: sessionRepo}
}

// Name returns the stable plugin name used by the accounting dispatcher.
func (h *UpdateHandler) Name() string {
	return "UpdateHandler"
}

// CanHandle reports whether the context represents an Accounting-Interim-Update
// packet.
func (h *UpdateHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	return ctx.StatusType == int(rfc2866.AcctStatusType_Value_InterimUpdate)
}

// Handle updates the online-session row with the latest traffic and session
// counters carried in the interim packet.
//
// Handle returns repository errors unchanged so callers can apply the standard
// accounting pipeline retry and logging policy.
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

// buildOnlineFromRequest maps the interim packet payload to the subset of
// RadiusOnline fields that are refreshed during update handling.
//
// The function intentionally omits immutable identity fields because the update
// repository path keys by Acct-Session-Id and only patches counters/timestamps.
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
