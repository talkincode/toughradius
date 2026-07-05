package enhancers

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/cisco"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// ciscoAddrPoolAVPair is the Cisco-AVPair directive prefix that assigns a
// subscriber's framed address from a NAS-local address pool. Cisco IOS /
// IOS-XE / IOS-XR and ISG honor the string form "ip:addr-pool=<pool-name>";
// several of these platforms select a local pool only from this vendor AVPair
// and ignore the standard Framed-Pool (RFC 2869, attribute 88) that
// DefaultAcceptEnhancer also emits, so sending both maximizes cross-platform
// pool assignment.
const ciscoAddrPoolAVPair = "ip:addr-pool="

// CiscoAcceptEnhancer adds Cisco's canonical Cisco-AVPair (SMI Private
// Enterprise Code 9, VSA type 1) reply attributes to an Access-Accept for
// sessions terminating on a Cisco NAS.
type CiscoAcceptEnhancer struct{}

// NewCiscoAcceptEnhancer returns a CiscoAcceptEnhancer ready for registration.
func NewCiscoAcceptEnhancer() *CiscoAcceptEnhancer {
	return &CiscoAcceptEnhancer{}
}

// Name returns the response-enhancer registry key for the Cisco enhancer.
func (e *CiscoAcceptEnhancer) Name() string {
	return "accept-cisco"
}

// Enhance appends Cisco's primary provisioning attribute to the Access-Accept:
//
//   - Cisco-AVPair "ip:addr-pool=<pool>" (type 1, string): assigns the framed
//     address from a NAS-local pool named by the user's configured address pool,
//     resolved through GetAddrPool so profile inheritance is honored. This is
//     the most widely deployed Cisco reply AVPair for subscriber IP
//     provisioning. It is appended (not set) because Cisco-AVPair is a
//     multi-valued attribute: each directive is carried as its own instance, so
//     appending never clobbers another AVPair sharing the packet.
//
// Enhance is a no-op for a nil context/response/user, for non-Cisco NAS
// devices, and when the address pool is unset, so it never mutates an
// Access-Accept it does not own.
func (e *CiscoAcceptEnhancer) Enhance(ctx context.Context, authCtx *auth.AuthContext) error {
	if authCtx == nil || authCtx.Response == nil || authCtx.User == nil {
		return nil
	}
	if !matchVendor(authCtx, vendors.CodeCisco) {
		return nil
	}

	user := authCtx.User
	resp := authCtx.Response

	// Get profile cache from metadata
	var profileCache interface{}
	if authCtx.Metadata != nil {
		profileCache = authCtx.Metadata["profile_cache"]
	}

	// Cisco-AVPair "ip:addr-pool=<pool>": assign the framed address from a
	// NAS-local pool when the user (or its profile) names one.
	addrPool := user.GetAddrPool(profileCache)
	if common.IsNotEmptyAndNA(addrPool) {
		_ = cisco.CiscoAVPair_AddString(resp, ciscoAddrPoolAVPair+addrPool) //nolint:errcheck
	}

	return nil
}
