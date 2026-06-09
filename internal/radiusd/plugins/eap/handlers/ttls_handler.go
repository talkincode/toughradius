package handlers

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsfragment"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	// EAPMethodTTLS is the configuration / handler name for EAP-TTLSv0.
	EAPMethodTTLS = "eap-ttls"

	// ttlsVersion0 is the EAP-TTLS version advertised in the low three bits of
	// the Flags octet. RFC 5281 §9.1 defines the Flags octet as
	// "L M S R R V V V", where the three-bit Version field (V) is 000 for
	// EAP-TTLSv0. Encoding version 0 makes the EAP-TTLS Start byte-identical to
	// an EAP-TLS Start (RFC 5216 §3.1), so the shared tlsfragment framing
	// applies unchanged.
	ttlsVersion0 = 0x00
)

// TTLSHandler is the EAP-TTLSv0 (Tunneled TLS, RFC 5281) authentication handler.
//
// EAP-TTLS establishes a TLS tunnel using only a server certificate (no client
// certificate is required) and then carries legacy inner authentication —
// PAP / CHAP / MS-CHAP / MS-CHAP-V2 or an inner EAP method — inside that tunnel
// as Diameter AVPs (RFC 5281 §10/§11). Its practical value is back-end
// adaptation: an existing username/password user store (LDAP, legacy databases,
// mixed clients) can be protected by the server-side TLS tunnel without
// migrating every client to a certificate identity. The outer tunnel reuses the
// EAP-TLS Flags/fragmentation framing (RFC 5216 §3.1) carried over RADIUS
// EAP-Message attributes (RFC 3579).
//
// Milestone scope:
//   - M9.1 (current): registers the method (EAP type 21, name "eap-ttls") and
//     answers EAP-Response/Identity with an EAP-TTLSv0 Start. HandleResponse is
//     a safe stub that always rejects with eap.ErrTTLSNotImplemented until the
//     outer tunnel lands, so the skeleton can never grant access.
//   - M9.2: establishes the outer TLS tunnel and fragmentation reassembly,
//     reusing the EAP-TLS state machine.
//   - M9.3: tunneled AVP encapsulation and inner PAP authentication.
//   - M9.4: inner MS-CHAP-V2 authentication and key derivation.
type TTLSHandler struct{}

// NewTTLSHandler creates an EAP-TTLSv0 handler skeleton. It accepts the EAP-TTLS
// Start exchange but rejects challenge responses safely with
// eap.ErrTTLSNotImplemented until the outer TLS tunnel is implemented (M9.2), so
// it can never authenticate a client.
func NewTTLSHandler() *TTLSHandler {
	return &TTLSHandler{}
}

// Name returns the handler name ("eap-ttls").
func (h *TTLSHandler) Name() string {
	return EAPMethodTTLS
}

// EAPType returns the EAP method type code handled (21, EAP-TTLS).
func (h *TTLSHandler) EAPType() uint8 {
	return eap.TypeTTLS
}

// CanHandle reports whether this handler can process the EAP message.
func (h *TTLSHandler) CanHandle(ctx *eap.EAPContext) bool {
	if ctx == nil || ctx.EAPMessage == nil {
		return false
	}
	return ctx.EAPMessage.Type == eap.TypeTTLS
}

// HandleIdentity handles EAP-Response/Identity by sending an EAP-TTLSv0 Start
// request (an EAP-Request with EAP-Type=TTLS, the Start (S) bit set, version
// bits 0, and no TLS data; RFC 5281 §9.2, framing per RFC 5216 §3.1). It
// persists handshake state keyed by the RADIUS State attribute so subsequent
// rounds (milestone M9.2) can correlate the tunnel.
func (h *TTLSHandler) HandleIdentity(ctx *eap.EAPContext) (bool, error) {
	eapData := h.buildStartRequest(ctx.EAPMessage.Identifier)

	stateID := common.UUID()
	username := rfc2865.UserName_GetString(ctx.Request.Packet)

	state := &eap.EAPState{
		Username: username,
		StateID:  stateID,
		Method:   EAPMethodTTLS,
		Success:  false,
	}

	if err := ctx.StateManager.SetState(stateID, state); err != nil {
		return false, err
	}

	return true, h.writeChallenge(ctx, stateID, eapData)
}

// HandleResponse is the M9.1 safe stub for EAP-TTLS challenge responses. The
// outer TLS tunnel (M9.2) and inner AVP authentication (M9.3+) are not yet
// implemented, so it always rejects with eap.ErrTTLSNotImplemented and never
// grants access.
func (h *TTLSHandler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	return false, eap.ErrTTLSNotImplemented
}

// buildStartRequest constructs an EAP-TTLSv0 Start request: an EAP-Request with
// EAP-Type=TTLS and a single Flags octet carrying the Start (S) bit and version
// 0. No TLS data is included, so the L bit is clear and the TLS Message Length
// field is absent (RFC 5281 §9.1/§9.2).
func (h *TTLSHandler) buildStartRequest(identifier uint8) []byte {
	frag := &tlsfragment.Packet{Flags: tlsfragment.FlagStart | ttlsVersion0}
	msg := &eap.EAPMessage{
		Code:       eap.CodeRequest,
		Identifier: identifier,
		Type:       eap.TypeTTLS,
		Data:       frag.Encode(),
	}
	return msg.Encode()
}

// writeChallenge sends eapData inside a RADIUS Access-Challenge, echoing the
// handshake State attribute and protecting the response with a
// Message-Authenticator (RFC 3579 §3.2).
func (h *TTLSHandler) writeChallenge(ctx *eap.EAPContext, stateID string, eapData []byte) error {
	response := ctx.Request.Response(radius.CodeAccessChallenge)
	_ = rfc2865.State_SetString(response, stateID) //nolint:errcheck
	eap.SetEAPMessageAndAuth(response, eapData, ctx.Secret)
	return ctx.ResponseWriter.Write(response)
}
