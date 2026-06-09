package handlers

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsengine"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsfragment"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	// EAPMethodPEAP is the configuration / handler name for PEAPv0.
	EAPMethodPEAP = "eap-peap"

	// peapVersion0 is the PEAP version advertised in the low two bits of the
	// PEAP Flags octet. PEAP repurposes EAP-TLS's two Reserved bits (RFC 5216
	// §3.1) as a Version field; PEAPv0 is version 0 (Microsoft [MS-PEAP] §2.2.2,
	// draft-kamath-pppext-peapv0 §1.1). Encoding version 0 makes the PEAPv0
	// Start byte-identical to an EAP-TLS Start, so the shared tlsfragment
	// framing applies unchanged.
	peapVersion0 = 0x00
)

// PEAPHandler is the PEAPv0 (Protected EAP) authentication handler.
//
// PEAP establishes a TLS tunnel using only a server certificate (no client
// certificate is required) and then runs an inner EAP method — for ToughRADIUS,
// EAP-MSCHAPv2 — inside that tunnel, exporting MPPE keys from the TLS master
// secret. The outer tunnel reuses the EAP-TLS Flags/fragmentation framing
// (RFC 5216 §3.1) carried over RADIUS EAP-Message attributes (RFC 3579).
//
// PEAPv0 has no formal IETF RFC; it is specified by Microsoft [MS-PEAP] and
// IETF draft-kamath-pppext-peapv0. It is a compatibility-first method for
// Windows / Active Directory estates: the inner MS-CHAPv2 exchange carries the
// same NTLMv1-class weaknesses Microsoft documents, so PEAP is appropriate for
// "must serve legacy devices and AD users" deployments, not as a security
// selling point.
//
// Milestone scope:
//   - M8.1: registered the method (EAP type 25, name "eap-peap") and answered
//     EAP-Response/Identity with a PEAPv0 Start.
//   - M8.2: established the outer TLS tunnel and fragmentation reassembly,
//     reusing the EAP-TLS state machine.
//   - M8.3a: post-handshake application-data exchange and RFC 5705 key export
//     in the tlsengine package.
//   - M8.3b (current): runs the inner EAP-MSCHAPv2 exchange (RFC 2759) inside
//     the tunnel, validating the NT-Response and, on success, deriving the
//     MS-MPPE-Send/Recv keys from the TLS session (RFC 5705 / RFC 2548).
type PEAPHandler struct {
	configProvider TLSConfigProvider
	maxFragment    int
}

// NewPEAPHandler creates a PEAPv0 handler without TLS material configured. It
// accepts the PEAP Start exchange but rejects handshake attempts safely until a
// server-certificate config provider is supplied.
func NewPEAPHandler() *PEAPHandler {
	return &PEAPHandler{maxFragment: defaultMaxTLSFragment}
}

// NewPEAPHandlerWithConfig creates a PEAPv0 handler that drives the outer TLS
// tunnel using the TLS material returned by provider.
func NewPEAPHandlerWithConfig(provider TLSConfigProvider) *PEAPHandler {
	return &PEAPHandler{configProvider: provider, maxFragment: defaultMaxTLSFragment}
}

func (h *PEAPHandler) newTunnel() *tlsTunnel {
	t := &tlsTunnel{
		eapType:           eap.TypePEAP,
		maxFragment:       h.maxFragment,
		configProvider:    h.configProvider,
		onApplicationData: h.handleInnerEAP,
	}
	// Once the outer tunnel is established, PEAP starts the inner EAP-MSCHAPv2
	// exchange (RFC 2759) inside the tunnel and derives MPPE keys from the TLS
	// session (RFC 5705 / RFC 2548) instead of granting directly. The TLS
	// engine is kept alive across inner rounds.
	t.onHandshakeComplete = func(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine) (bool, error) {
		return t.startInner(ctx, state, engine)
	}
	return t
}

// Name returns the handler name ("eap-peap").
func (h *PEAPHandler) Name() string {
	return EAPMethodPEAP
}

// EAPType returns the EAP method type code handled (25, PEAP).
func (h *PEAPHandler) EAPType() uint8 {
	return eap.TypePEAP
}

// CanHandle reports whether this handler can process the EAP message.
func (h *PEAPHandler) CanHandle(ctx *eap.EAPContext) bool {
	if ctx == nil || ctx.EAPMessage == nil {
		return false
	}
	return ctx.EAPMessage.Type == eap.TypePEAP
}

// HandleIdentity handles EAP-Response/Identity by sending a PEAPv0 Start request
// (an EAP-Request with EAP-Type=PEAP, the Start (S) bit set, version bits 0, and
// no TLS data; Microsoft [MS-PEAP] §3.1.5.1, framing per RFC 5216 §3.1). It
// persists handshake state keyed by the RADIUS State attribute so subsequent
// rounds (milestone M8.2) can correlate the tunnel.
func (h *PEAPHandler) HandleIdentity(ctx *eap.EAPContext) (bool, error) {
	eapData := h.buildStartRequest(ctx.EAPMessage.Identifier)

	stateID := common.UUID()
	username := rfc2865.UserName_GetString(ctx.Request.Packet)

	state := &eap.EAPState{
		Username: username,
		StateID:  stateID,
		Method:   EAPMethodPEAP,
		Success:  false,
	}

	if err := ctx.StateManager.SetState(stateID, state); err != nil {
		return false, err
	}

	return true, h.writeChallenge(ctx, stateID, eapData)
}

// HandleResponse drives PEAP's outer TLS tunnel and fragmentation using the
// shared EAP-TLS state machine (RFC 5216 §2.1.5/§3.1; PEAPv0 [MS-PEAP]). Once
// the tunnel is established it carries the inner EAP-MSCHAPv2 exchange (RFC
// 2759) as TLS application data and, on success, derives the MS-MPPE keys from
// the TLS session (RFC 5705 / RFC 2548); see handleInnerEAP.
func (h *PEAPHandler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	return h.newTunnel().HandleResponse(ctx)
}

// buildStartRequest constructs a PEAPv0 Start request: an EAP-Request with
// EAP-Type=PEAP and a single Flags octet carrying the Start (S) bit and PEAP
// version 0. No TLS data is included, so the L bit is clear and the TLS Message
// Length field is absent.
func (h *PEAPHandler) buildStartRequest(identifier uint8) []byte {
	frag := &tlsfragment.Packet{Flags: tlsfragment.FlagStart | peapVersion0}
	msg := &eap.EAPMessage{
		Code:       eap.CodeRequest,
		Identifier: identifier,
		Type:       eap.TypePEAP,
		Data:       frag.Encode(),
	}
	return msg.Encode()
}

// writeChallenge sends eapData inside a RADIUS Access-Challenge, echoing the
// handshake State attribute and protecting the response with a
// Message-Authenticator (RFC 3579 §3.2).
func (h *PEAPHandler) writeChallenge(ctx *eap.EAPContext, stateID string, eapData []byte) error {
	response := ctx.Request.Response(radius.CodeAccessChallenge)
	_ = rfc2865.State_SetString(response, stateID) //nolint:errcheck
	eap.SetEAPMessageAndAuth(response, eapData, ctx.Secret)
	return ctx.ResponseWriter.Write(response)
}
