package handlers

import (
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"go.uber.org/zap"
	"layeh.com/radius/rfc2866"
)

// NasStateHandler handles Accounting-On and Accounting-Off events by clearing
// online sessions for the NAS that restarted or went offline.
type NasStateHandler struct {
	sessionRepo repository.SessionRepository
}

// NewNasStateHandler constructs a NasStateHandler with the session repository
// used for NAS-scoped bulk cleanup.
func NewNasStateHandler(sessionRepo repository.SessionRepository) *NasStateHandler {
	return &NasStateHandler{
		sessionRepo: sessionRepo,
	}
}

// Name returns the stable plugin name used by the accounting dispatcher.
func (h *NasStateHandler) Name() string {
	return "NasStateHandler"
}

// CanHandle reports whether the context represents Accounting-On or
// Accounting-Off.
func (h *NasStateHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	return ctx.StatusType == int(rfc2866.AcctStatusType_Value_AccountingOn) ||
		ctx.StatusType == int(rfc2866.AcctStatusType_Value_AccountingOff)
}

// Handle removes all online sessions bound to the NAS in the event context.
//
// Handle returns an error when repository access is unavailable or when NAS
// identity data is missing, because both are required to scope the cleanup.
func (h *NasStateHandler) Handle(ctx *accounting.AccountingContext) error {
	if h.sessionRepo == nil {
		return fmt.Errorf("session repository is not available")
	}
	if ctx.NAS == nil {
		return fmt.Errorf("nas information is missing")
	}

	if err := h.sessionRepo.BatchDeleteByNas(ctx.Context, ctx.NASIP, ctx.NAS.Identifier); err != nil {
		zap.L().Error("failed to clear sessions on NAS state change",
			zap.String("namespace", "radius"),
			zap.String("nas_ip", ctx.NASIP),
			zap.String("nas_id", ctx.NAS.Identifier),
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("cleared sessions due to NAS state change",
		zap.String("namespace", "radius"),
		zap.String("nas_ip", ctx.NASIP),
		zap.String("nas_id", ctx.NAS.Identifier),
		zap.Int("status_type", ctx.StatusType),
	)
	return nil
}
