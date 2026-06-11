package handlers

import (
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	vendorparserspkg "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
	"layeh.com/radius/rfc3162"
	"layeh.com/radius/rfc4818"
	"layeh.com/radius/rfc6911"
)

// StartHandler handles Accounting-Start packets and persists both online-session
// and accounting rows for newly created sessions.
//
// The handler treats repeated Accounting-Start packets for the same
// Acct-Session-Id as retransmissions and avoids double-billing by skipping
// duplicate accounting inserts.
type StartHandler struct {
	sessionRepo    repository.SessionRepository
	accountingRepo repository.AccountingRepository
}

// NewStartHandler constructs a StartHandler with repositories used to persist
// online-session and accounting state.
func NewStartHandler(
	sessionRepo repository.SessionRepository,
	accountingRepo repository.AccountingRepository,
) *StartHandler {
	return &StartHandler{
		sessionRepo:    sessionRepo,
		accountingRepo: accountingRepo,
	}
}

// Name returns the stable plugin name used by the accounting dispatcher.
func (h *StartHandler) Name() string {
	return "StartHandler"
}

// CanHandle reports whether the context represents an Accounting-Start packet.
func (h *StartHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	return ctx.StatusType == int(rfc2866.AcctStatusType_Value_Start)
}

// Handle writes a new online-session row and a matching accounting row.
//
// Handle is idempotent on Acct-Session-Id: retransmitted starts do not create
// duplicate records. If accounting-row creation fails after the online row was
// created, Handle best-effort rolls back the online row so a later
// retransmission can recreate both rows consistently.
func (h *StartHandler) Handle(acctCtx *accounting.AccountingContext) error {
	vendorReq := acctCtx.VendorReq
	if vendorReq == nil {
		vendorReq = &vendorparserspkg.VendorRequest{}
	}

	// Construct the online session record
	online := h.buildRadiusOnline(acctCtx.Request, vendorReq, acctCtx.NAS, acctCtx.NASIP)

	// Create online session. Create is idempotent on Acct-Session-Id: a
	// retransmitted Accounting-Start returns created=false instead of
	// inserting a duplicate online row.
	created, err := h.sessionRepo.Create(acctCtx.Context, &online)
	if err != nil {
		zap.L().Error("add radius online error",
			zap.String("namespace", "radius"),
			zap.String("username", acctCtx.Username),
			zap.Error(err),
		)
		return err
	}

	// Duplicate Accounting-Start retransmission: the online session already
	// exists for this Acct-Session-Id. Skip creating another accounting record
	// so traffic/time counters are not double-billed.
	if !created {
		zap.L().Debug("duplicate accounting start ignored",
			zap.String("namespace", "radius"),
			zap.String("username", acctCtx.Username),
			zap.String("session_id", online.AcctSessionId),
		)
		return nil
	}

	// Create accounting record (only for newly created sessions)
	accounting := h.buildRadiusAccounting(&online, true)
	if err := h.accountingRepo.Create(acctCtx.Context, &accounting); err != nil {
		zap.L().Error("add radius accounting error",
			zap.String("namespace", "radius"),
			zap.String("username", acctCtx.Username),
			zap.Error(err),
		)
		// Compensating delete: the online row was just inserted but the
		// accounting record could not be created. Remove the online row so the
		// NAS retransmission is treated as a fresh start (Create returns
		// created=true) and both records are recreated, instead of being
		// skipped as a duplicate and leaving the session without accounting.
		if delErr := h.sessionRepo.Delete(acctCtx.Context, online.AcctSessionId); delErr != nil {
			zap.L().Error("rollback online after accounting error failed",
				zap.String("namespace", "radius"),
				zap.String("username", acctCtx.Username),
				zap.String("session_id", online.AcctSessionId),
				zap.Error(delErr),
			)
		}
		return err
	}

	return nil
}

// buildRadiusOnline maps an Accounting-Start request into the canonical online
// session model persisted by SessionRepository.Create.
//
// Missing vendor fields and optional RADIUS attributes are normalized to
// common.NA so later filtering and presentation layers can treat absence
// consistently across NAS implementations.
func (h *StartHandler) buildRadiusOnline(r *radius.Request, vr *vendorparserspkg.VendorRequest, nas *domain.NetNas, nasrip string) domain.RadiusOnline {
	acctInputOctets := int(rfc2866.AcctInputOctets_Get(r.Packet))
	acctInputGigawords := int(rfc2869.AcctInputGigawords_Get(r.Packet))
	acctOutputOctets := int(rfc2866.AcctOutputOctets_Get(r.Packet))
	acctOutputGigawords := int(rfc2869.AcctOutputGigawords_Get(r.Packet))

	getAcctStartTime := func(sessionTime int) time.Time {
		m, _ := time.ParseDuration(fmt.Sprintf("-%ds", sessionTime))
		return time.Now().Add(m)
	}

	return domain.RadiusOnline{
		ID:                  common.UUIDint64(),
		Username:            rfc2865.UserName_GetString(r.Packet),
		NasId:               common.IfEmptyStr(rfc2865.NASIdentifier_GetString(r.Packet), common.NA),
		NasAddr:             nas.Ipaddr,
		NasPaddr:            nasrip,
		SessionTimeout:      int(rfc2865.SessionTimeout_Get(r.Packet)),
		FramedIpaddr:        common.IfEmptyStr(rfc2865.FramedIPAddress_Get(r.Packet).String(), common.NA),
		FramedNetmask:       common.IfEmptyStr(rfc2865.FramedIPNetmask_Get(r.Packet).String(), common.NA),
		FramedIpv6Address:   ipv6AddrOrNA(rfc6911.FramedIPv6Address_Get(r.Packet)),
		FramedIpv6Prefix:    ipv6PrefixOrNA(rfc3162.FramedIPv6Prefix_Get(r.Packet)),
		DelegatedIpv6Prefix: ipv6PrefixOrNA(rfc4818.DelegatedIPv6Prefix_Get(r.Packet)),
		MacAddr:             common.IfEmptyStr(vr.MacAddr, common.NA),
		NasPort:             0,
		NasClass:            common.NA,
		NasPortId:           common.IfEmptyStr(rfc2869.NASPortID_GetString(r.Packet), common.NA),
		NasPortType:         0,
		ServiceType:         0,
		AcctSessionId:       rfc2866.AcctSessionID_GetString(r.Packet),
		AcctSessionTime:     int(rfc2866.AcctSessionTime_Get(r.Packet)),
		AcctInputTotal:      int64(acctInputOctets) + int64(acctInputGigawords)*4*1024*1024*1024,
		AcctOutputTotal:     int64(acctOutputOctets) + int64(acctOutputGigawords)*4*1024*1024*1024,
		AcctInputPackets:    int(rfc2866.AcctInputPackets_Get(r.Packet)),
		AcctOutputPackets:   int(rfc2866.AcctOutputPackets_Get(r.Packet)),
		AcctStartTime:       getAcctStartTime(int(rfc2866.AcctSessionTime_Get(r.Packet))),
		LastUpdate:          time.Now(),
	}
}

// buildRadiusAccounting projects an online-session snapshot into the accounting
// row model used by the historical ledger.
//
// When start is false, buildRadiusAccounting stamps AcctStopTime for the
// terminal record shape used by stop-event updates.
func (h *StartHandler) buildRadiusAccounting(online *domain.RadiusOnline, start bool) domain.RadiusAccounting {
	accounting := domain.RadiusAccounting{
		ID:                  common.UUIDint64(),
		Username:            online.Username,
		AcctSessionId:       online.AcctSessionId,
		NasId:               online.NasId,
		NasAddr:             online.NasAddr,
		NasPaddr:            online.NasPaddr,
		SessionTimeout:      online.SessionTimeout,
		FramedIpaddr:        online.FramedIpaddr,
		FramedNetmask:       online.FramedNetmask,
		FramedIpv6Prefix:    online.FramedIpv6Prefix,
		FramedIpv6Address:   online.FramedIpv6Address,
		DelegatedIpv6Prefix: online.DelegatedIpv6Prefix,
		MacAddr:             online.MacAddr,
		NasPort:             online.NasPort,
		NasClass:            online.NasClass,
		NasPortId:           online.NasPortId,
		NasPortType:         online.NasPortType,
		ServiceType:         online.ServiceType,
		AcctSessionTime:     online.AcctSessionTime,
		AcctInputTotal:      online.AcctInputTotal,
		AcctOutputTotal:     online.AcctOutputTotal,
		AcctInputPackets:    online.AcctInputPackets,
		AcctOutputPackets:   online.AcctOutputPackets,
		LastUpdate:          time.Now(),
		AcctStartTime:       online.AcctStartTime,
	}

	if !start {
		accounting.AcctStopTime = time.Now()
	}

	return accounting
}
