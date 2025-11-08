package radiusd

import (
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// CheckVlanBind
// vlanid binding detection
// Only if both user vlanid and request vlanid are valid.
// If user vlanid is empty, update user vlanid directly.
func (s *AuthService) CheckVlanBind(user *domain.RadiusUser, vendorReq *VendorRequest) error {
	if user.BindVlan == 0 {
		return nil
	}
	reqvid1 := int(vendorReq.Vlanid1)
	reqvid2 := int(vendorReq.Vlanid2)
	if user.Vlanid1 != 0 && vendorReq.Vlanid1 != 0 && user.Vlanid1 != reqvid1 {
		return NewAuthError(app.MetricsRadiusRejectBindError, "user vlanid1 bind not match")
	}

	if user.Vlanid2 != 0 && reqvid2 != 0 && user.Vlanid2 != reqvid2 {
		return NewAuthError(app.MetricsRadiusRejectBindError, "user vlanid2 bind not match")
	}

	return nil
}

// CheckMacBind
// mac binding detection
// Detected only if both user mac and request mac are valid.
// If user mac is empty, update user mac directly.
func (s *AuthService) CheckMacBind(user *domain.RadiusUser, vendorReq *VendorRequest) error {
	if user.BindMac == 0 {
		return nil
	}

	if common.IsNotEmptyAndNA(user.MacAddr) && vendorReq.MacAddr != "" && user.MacAddr != vendorReq.MacAddr {
		return NewAuthError(app.MetricsRadiusRejectBindError, "user mac bind not match")
	}
	return nil
}

// UpdateBind
// update mac or vlan
func (s *AuthService) UpdateBind(user *domain.RadiusUser, vendorReq *VendorRequest) {
	if user.MacAddr != vendorReq.MacAddr {
		s.UpdateUserMac(user.Username, vendorReq.MacAddr)
	}
	reqvid1 := int(vendorReq.Vlanid1)
	reqvid2 := int(vendorReq.Vlanid2)
	if user.Vlanid1 != reqvid1 {
		s.UpdateUserVlanid2(user.Username, reqvid1)
	}
	if user.Vlanid2 != reqvid2 {
		s.UpdateUserVlanid2(user.Username, reqvid2)
	}
}
