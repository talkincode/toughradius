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
) // StartHandler Accounting Start handler
type StartHandler struct {
	sessionRepo    repository.SessionRepository
	accountingRepo repository.AccountingRepository
}

// NewStartHandler CreateAccounting Start handler
func NewStartHandler(
	sessionRepo repository.SessionRepository,
	accountingRepo repository.AccountingRepository,
) *StartHandler {
	return &StartHandler{
		sessionRepo:    sessionRepo,
		accountingRepo: accountingRepo,
	}
}

func (h *StartHandler) Name() string {
	return "StartHandler"
}

func (h *StartHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	return ctx.StatusType == int(rfc2866.AcctStatusType_Value_Start)
}

func (h *StartHandler) Handle(acctCtx *accounting.AccountingContext) error {
	vendorReq := acctCtx.VendorReq
	if vendorReq == nil {
		vendorReq = &vendorparserspkg.VendorRequest{}
	}

	// Construct the online session record
	online := h.buildRadiusOnline(acctCtx.Request, vendorReq, acctCtx.NAS, acctCtx.NASIP)

	// Create online session
	err := h.sessionRepo.Create(acctCtx.Context, &online)
	if err != nil {
		zap.L().Error("add radius online error",
			zap.String("namespace", "radius"),
			zap.String("username", acctCtx.Username),
			zap.Error(err),
		)
		return err
	}

	// Create accounting record
	accounting := h.buildRadiusAccounting(&online, true)
	if err := h.accountingRepo.Create(acctCtx.Context, &accounting); err != nil {
		zap.L().Error("add radius accounting error",
			zap.String("namespace", "radius"),
			zap.String("username", acctCtx.Username),
			zap.Error(err),
		)
		return err
	}

	return nil
}

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
		FramedIpv6Address:   common.NA, // Set based on vendor-specific logic
		FramedIpv6Prefix:    common.IfEmptyStr(rfc3162.FramedIPv6Prefix_Get(r.Packet).String(), common.NA),
		DelegatedIpv6Prefix: common.IfEmptyStr(rfc4818.DelegatedIPv6Prefix_Get(r.Packet).String(), common.NA),
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
