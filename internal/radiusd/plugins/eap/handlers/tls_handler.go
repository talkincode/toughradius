package handlers

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	EAPMethodTLS = "eap-tls"

	// EAP-TLS Flags bits (RFC 5216 §3.1).
	//
	//	0 1 2 3 4 5 6 7
	//	+-+-+-+-+-+-+-+-+
	//	|L M S R R R R R|
	//	+-+-+-+-+-+-+-+-+
	TLSFlagLengthIncluded = 0x80 // L: TLS Message Length field present
	TLSFlagMoreFragments  = 0x40 // M: more fragments follow
	TLSFlagStart          = 0x20 // S: EAP-TLS Start
)

// TLSHandler is the EAP-TLS authentication handler.
//
// This is the milestone M1.1 skeleton: it registers the EAP-TLS method, sends a
// well-formed EAP-TLS Start request on identity, and manages handshake state via
// the shared state manager. The actual TLS handshake, fragmentation/reassembly
// (RFC 5216 §2.1.5 / RFC 7499) and certificate validation are delivered in later
// M1 subtasks. Until then HandleResponse rejects safely so the skeleton can never
// authenticate a client.
type TLSHandler struct{}

// NewTLSHandler creates an EAP-TLS handler.
func NewTLSHandler() *TLSHandler {
	return &TLSHandler{}
}

// Name returns the handler name.
func (h *TLSHandler) Name() string {
	return EAPMethodTLS
}

// EAPType returns the EAP type code handled (13, RFC 5216 §3.1).
func (h *TLSHandler) EAPType() uint8 {
	return eap.TypeTLS
}

// CanHandle reports whether this handler can process the EAP message.
func (h *TLSHandler) CanHandle(ctx *eap.EAPContext) bool {
	if ctx == nil || ctx.EAPMessage == nil {
		return false
	}
	return ctx.EAPMessage.Type == eap.TypeTLS
}

// HandleIdentity handles EAP-Response/Identity by sending an EAP-TLS Start
// request (RFC 5216 §2.1.1: an EAP-Request with EAP-Type=EAP-TLS, the Start (S)
// bit set, and no data).
func (h *TLSHandler) HandleIdentity(ctx *eap.EAPContext) (bool, error) {
	eapData := h.buildStartRequest(ctx.EAPMessage.Identifier)

	// Create RADIUS Access-Challenge response.
	response := ctx.Request.Response(radius.CodeAccessChallenge)

	// Generate and save the handshake state.
	stateID := common.UUID()
	username := rfc2865.UserName_GetString(ctx.Request.Packet)

	state := &eap.EAPState{
		Username: username,
		StateID:  stateID,
		Method:   EAPMethodTLS,
		Success:  false,
	}

	if err := ctx.StateManager.SetState(stateID, state); err != nil {
		return false, err
	}

	// Set the State attribute so the client echoes it on the next request.
	_ = rfc2865.State_SetString(response, stateID) //nolint:errcheck

	// Set the EAP-Message and Message-Authenticator.
	eap.SetEAPMessageAndAuth(response, eapData, ctx.Secret)

	// Send the EAP-TLS Start.
	return true, ctx.ResponseWriter.Write(response)
}

// HandleResponse handles EAP-Response (TLS handshake messages).
//
// The TLS handshake is not implemented yet (milestone M1.2). To avoid exposing
// an EAP method that could grant access without a completed, validated TLS
// handshake, this skeleton always rejects with an explicit reason.
func (h *TLSHandler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	stateID := rfc2865.State_GetString(ctx.Request.Packet)
	if stateID == "" {
		return false, eap.ErrStateNotFound
	}

	if _, err := ctx.StateManager.GetState(stateID); err != nil {
		return false, err
	}

	return false, eap.ErrTLSHandshakeNotImplemented
}

// buildStartRequest constructs an EAP-TLS Start request (RFC 5216 §3.1).
//
// EAP-TLS Start format: Code (1) | Identifier (1) | Length (2) | Type (1) |
// Flags (1). The Start (S) bit is set and no TLS data is included, so the L bit
// is clear and the TLS Message Length field is absent.
func (h *TLSHandler) buildStartRequest(identifier uint8) []byte {
	const totalLen = 6 // EAP header (4) + Type (1) + Flags (1)

	buffer := make([]byte, totalLen)
	buffer[0] = eap.CodeRequest
	buffer[1] = identifier
	buffer[2] = byte(totalLen >> 8)
	buffer[3] = byte(totalLen)
	buffer[4] = eap.TypeTLS
	buffer[5] = TLSFlagStart

	return buffer
}
