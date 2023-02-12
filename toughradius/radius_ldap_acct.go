package toughradius

import (
	"github.com/talkincode/toughradius/models"
	"layeh.com/radius"
	"layeh.com/radius/rfc2866"
)

func (s *AcctService) LdapUserAcct(r *radius.Request, vr *VendorRequest, username string, vpe *models.NetVpe, nasrip string) {
	statusType := rfc2866.AcctStatusType_Get(r.Packet)
	switch statusType {
	case rfc2866.AcctStatusType_Value_Start:
		s.DoAcctStart(r, vr, username, vpe, nasrip)
	case rfc2866.AcctStatusType_Value_InterimUpdate:
		s.DoAcctUpdate(r, vr, username, vpe, nasrip)
	case rfc2866.AcctStatusType_Value_Stop:
		s.DoAcctStop(r, vr, username, vpe, nasrip)
	case rfc2866.AcctStatusType_Value_AccountingOn:
		s.DoAcctNasOn(r)
	case rfc2866.AcctStatusType_Value_AccountingOff:
		s.DoAcctNasOff(r)
	}
}
