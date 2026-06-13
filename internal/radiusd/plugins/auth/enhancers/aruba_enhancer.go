package enhancers

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/aruba"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// arubaMaxVLANID is the largest assignable IEEE 802.1Q VLAN ID. VLAN 0 and 4095
// are reserved, so a valid subscriber VLAN is 1..4094. Bounding Vlanid1 to this
// range both enforces a valid 802.1Q tag and keeps the int->uint32 conversion in
// Enhance safe from overflow.
const arubaMaxVLANID = 4094

// ArubaAcceptEnhancer adds Aruba's canonical Access-Accept policy attributes for
// sessions terminating on an Aruba/HPE NAS (SMI Private Enterprise Code 14823).
type ArubaAcceptEnhancer struct{}

// NewArubaAcceptEnhancer returns an ArubaAcceptEnhancer ready for registration.
func NewArubaAcceptEnhancer() *ArubaAcceptEnhancer {
	return &ArubaAcceptEnhancer{}
}

// Name returns the response-enhancer registry key for the Aruba enhancer.
func (e *ArubaAcceptEnhancer) Name() string {
	return "accept-aruba"
}

// Enhance appends Aruba's primary policy attributes to the Access-Accept:
//
//   - Aruba-User-Vlan (type 2, integer): the subscriber's assigned VLAN, taken
//     from the user's configured Vlanid1. Aruba controllers honor this attribute
//     to place the client into a RADIUS-assigned VLAN. Only a valid 802.1Q ID
//     (1..4094) is emitted; Vlanid2 (typically the QinQ outer tag) has no
//     Aruba-User-Vlan equivalent and is intentionally not sent.
//   - Aruba-User-Role (type 1, string): the subscriber's firewall/user role,
//     sourced from the generic Domain vendor-policy field via GetDomain (so it
//     honors profile inheritance). This mirrors the Huawei enhancer, which maps
//     the same Domain field to Huawei-Domain-Name: in both cases Domain carries
//     the NAS-specific primary policy selector.
//
// Enhance is a no-op for a nil context/response/user, for non-Aruba NAS devices,
// and when the source fields are unset, so it never mutates an Access-Accept it
// does not own.
func (e *ArubaAcceptEnhancer) Enhance(ctx context.Context, authCtx *auth.AuthContext) error {
	if authCtx == nil || authCtx.Response == nil || authCtx.User == nil {
		return nil
	}
	if !matchVendor(authCtx, vendors.CodeAruba) {
		return nil
	}

	user := authCtx.User
	resp := authCtx.Response

	// Get profile cache from metadata
	var profileCache interface{}
	if authCtx.Metadata != nil {
		profileCache = authCtx.Metadata["profile_cache"]
	}

	// Aruba-User-Vlan: assign the subscriber VLAN when a valid 802.1Q ID is
	// configured. The 1..4094 bound also keeps the int->uint32 conversion safe.
	if vlan := user.Vlanid1; vlan >= 1 && vlan <= arubaMaxVLANID {
		_ = aruba.ArubaUserVlan_Set(resp, aruba.ArubaUserVlan(vlan)) //nolint:errcheck,gosec // G115: vlan is bounded to 1..4094 above
	}

	// Aruba-User-Role: reuse the generic Domain vendor-policy field as the Aruba
	// user role, mirroring the Huawei enhancer's Domain->Huawei-Domain-Name map.
	role := user.GetDomain(profileCache)
	if common.IsNotEmptyAndNA(role) {
		_ = aruba.ArubaUserRole_SetString(resp, role) //nolint:errcheck
	}

	return nil
}
