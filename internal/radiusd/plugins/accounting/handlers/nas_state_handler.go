package handlers

import (
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"go.uber.org/zap"
	"layeh.com/radius/rfc2866"
)

// NasStateHandler Handle Accounting-On/Off event
type NasStateHandler struct {
	sessionRepo repository.SessionRepository
}

func NewNasStateHandler(sessionRepo repository.SessionRepository) *NasStateHandler {
	return &NasStateHandler{
		sessionRepo: sessionRepo,
	}
}

func (h *NasStateHandler) Name() string {
	return "NasStateHandler"
}

func (h *NasStateHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	return ctx.StatusType == int(rfc2866.AcctStatusType_Value_AccountingOn) ||
		ctx.StatusType == int(rfc2866.AcctStatusType_Value_AccountingOff)
}

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
