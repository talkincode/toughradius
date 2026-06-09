package handlers

import (
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsengine"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsfragment"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	EAPMethodTLS = "eap-tls"

	// EAP-TLS Flags bits (RFC 5216 §3.1). These mirror tlsfragment's canonical
	// definitions and are retained for readability at the handler boundary.
	//
	//0 1 2 3 4 5 6 7
	//+-+-+-+-+-+-+-+-+
	//|L M S R R R R R|
	//+-+-+-+-+-+-+-+-+
	TLSFlagLengthIncluded = tlsfragment.FlagLengthIncluded // L: TLS Message Length field present
	TLSFlagMoreFragments  = tlsfragment.FlagMoreFragments  // M: more fragments follow
	TLSFlagStart          = tlsfragment.FlagStart          // S: EAP-TLS Start

	// State keys for persisting EAP-TLS reassembly progress across rounds.
	stateKeyRxBuf      = "tls_rx_buf"      // accumulated inbound TLS bytes
	stateKeyRxDeclared = "tls_rx_declared" // declared TLS Message Length (L flag)
	stateKeyRxHasLen   = "tls_rx_haslen"   // whether a TLS Message Length was declared

	// State keys for driving the TLS handshake engine and outbound
	// fragmentation across rounds.
	stateKeyEngine         = "tls_engine"          // *tlsengine.Engine (live handshake)
	stateKeyOutFrags       = "tls_out_frags"       // []*tlsfragment.Packet queued for sending
	stateKeyHandshakeDone  = "tls_handshake_done"  // bool: TLS handshake completed, draining final flight
	stateKeyPendingSuccess = "tls_pending_success" // bool: awaiting the peer ACK before EAP-Success

	// defaultMaxTLSFragment bounds the TLS data carried by a single EAP-TLS
	// fragment so that each EAP-Request comfortably fits within one RADIUS
	// packet (RFC 5216 §2.1.5 / RFC 7499).
	defaultMaxTLSFragment = 1024
)

// TLSConfigProvider supplies the per-handshake TLS materials for EAP-TLS and
// server-only tunneled methods such as PEAP. It returns a nil config (and nil
// error) when the method is not configured, in which case the handler rejects
// safely. The provider is consulted at the start of each handshake so that
// certificate/CA changes take effect without restarting the handler.
type TLSConfigProvider func() (*tlsengine.Config, error)

// TLSHandler is the EAP-TLS authentication handler.
//
// Milestone M1.2 added TLS handshake state management and fragmentation
// reassembly (RFC 5216 §2.1.5 / RFC 7499). Milestone M1.3 drives the actual
// server-side TLS handshake through the tlsengine package: it validates the
// peer certificate against the configured CA chain
// (ClientAuth=RequireAndVerifyClientCert, RFC 5216 §2.2 / §5.3) and maps the
// certificate identity to the RADIUS User-Name (RFC 5216 §5.2).
//
// When no TLS configuration is available the handler rejects safely with
// eap.ErrTLSNotConfigured, so it can never authenticate a client without
// configured trust anchors.
type TLSHandler struct {
	configProvider TLSConfigProvider
	maxFragment    int
}

// NewTLSHandler creates an EAP-TLS handler without TLS material configured. It
// will accept EAP-TLS Start exchanges but reject handshake attempts with
// eap.ErrTLSNotConfigured until a configuration provider is supplied (the
// runtime wiring of certificate configuration is delivered in milestone M1.5).
func NewTLSHandler() *TLSHandler {
	return &TLSHandler{maxFragment: defaultMaxTLSFragment}
}

// NewTLSHandlerWithConfig creates an EAP-TLS handler that drives handshakes
// using the TLS material returned by provider.
func NewTLSHandlerWithConfig(provider TLSConfigProvider) *TLSHandler {
	return &TLSHandler{configProvider: provider, maxFragment: defaultMaxTLSFragment}
}

func (h *TLSHandler) newTunnel() *tlsTunnel {
	return &tlsTunnel{
		eapType:        eap.TypeTLS,
		maxFragment:    h.maxFragment,
		configProvider: h.configProvider,
		onHandshakeComplete: func(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine) (bool, error) {
			return h.finalizeWithEngine(ctx, state, engine)
		},
	}
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

	// Send the EAP-TLS Start in an Access-Challenge.
	return true, h.writeChallenge(ctx, stateID, eapData)
}

// HandleResponse handles EAP-Response (TLS handshake messages).
//
// It reassembles fragmented inbound TLS data across EAP rounds (RFC 5216
// §2.1.5), drives the server-side TLS handshake through the engine, fragments
// the server's flights back to the peer, and—once the handshake completes—maps
// the verified certificate identity to the RADIUS User-Name before granting
// access. Every failure path returns an explicit error and never authenticates.
func (h *TLSHandler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	return h.newTunnel().HandleResponse(ctx)
}

func (h *TLSHandler) finalizeWithEngine(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine) (bool, error) {
	defer func() { _ = engine.Close() }()

	identity, err := engine.Identity()
	if err != nil {
		return false, fmt.Errorf("%w: %v", eap.ErrTLSNoIdentity, err)
	}
	if identity.Name == "" {
		return false, eap.ErrTLSNoIdentity
	}

	// Bind the TLS-authenticated certificate identity to the RADIUS User-Name.
	if state.Username != "" && !identity.Matches(state.Username) {
		return false, fmt.Errorf("%w: certificate identity %q does not match user %q",
			eap.ErrTLSIdentityMismatch, identity.Name, state.Username)
	}

	state.Success = true
	clearKey(state, stateKeyEngine)
	if err := ctx.StateManager.SetState(state.StateID, state); err != nil {
		return false, err
	}
	return true, nil
}

// writeChallenge sends eapData inside a RADIUS Access-Challenge, echoing the
// handshake State attribute and protecting the response with a
// Message-Authenticator.
func (h *TLSHandler) writeChallenge(ctx *eap.EAPContext, stateID string, eapData []byte) error {
	response := ctx.Request.Response(radius.CodeAccessChallenge)
	_ = rfc2865.State_SetString(response, stateID) //nolint:errcheck
	eap.SetEAPMessageAndAuth(response, eapData, ctx.Secret)
	return ctx.ResponseWriter.Write(response)
}

// buildStartRequest constructs an EAP-TLS Start request (RFC 5216 §3.1).
//
// EAP-TLS Start format: Code (1) | Identifier (1) | Length (2) | Type (1) |
// Flags (1). The Start (S) bit is set and no TLS data is included, so the L bit
// is clear and the TLS Message Length field is absent.
func (h *TLSHandler) buildStartRequest(identifier uint8) []byte {
	return h.buildEAPRequest(identifier, &tlsfragment.Packet{Flags: tlsfragment.FlagStart})
}

// buildEAPRequest wraps an EAP-TLS payload in an EAP-Request header.
func (h *TLSHandler) buildEAPRequest(identifier uint8, frag *tlsfragment.Packet) []byte {
	msg := &eap.EAPMessage{
		Code:       eap.CodeRequest,
		Identifier: identifier,
		Type:       eap.TypeTLS,
		Data:       frag.Encode(),
	}
	return msg.Encode()
}

// loadReassembler rebuilds the reassembly state persisted on the EAP state.
func loadReassembler(state *eap.EAPState) *tlsfragment.Reassembler {
	var (
		buf      []byte
		declared uint32
		hasLen   bool
	)
	if state.Data != nil {
		if v, ok := state.Data[stateKeyRxBuf].([]byte); ok {
			buf = v
		}
		if v, ok := state.Data[stateKeyRxDeclared].(uint32); ok {
			declared = v
		}
		if v, ok := state.Data[stateKeyRxHasLen].(bool); ok {
			hasLen = v
		}
	}
	return tlsfragment.LoadReassembler(buf, declared, hasLen, tlsfragment.DefaultMaxMessageLength)
}

// saveReassembler persists the reassembly progress back onto the EAP state.
func saveReassembler(state *eap.EAPState, r *tlsfragment.Reassembler) {
	if state.Data == nil {
		state.Data = make(map[string]interface{})
	}
	declared, hasLen := r.Declared()
	state.Data[stateKeyRxBuf] = r.Buffer()
	state.Data[stateKeyRxDeclared] = declared
	state.Data[stateKeyRxHasLen] = hasLen
}

// resetReassembler clears the persisted inbound reassembly buffer so the next
// inbound flight starts fresh.
func resetReassembler(state *eap.EAPState) {
	clearKey(state, stateKeyRxBuf)
	clearKey(state, stateKeyRxDeclared)
	clearKey(state, stateKeyRxHasLen)
}

// --- small typed helpers over the untyped state.Data map ------------------

func getFragments(state *eap.EAPState) []*tlsfragment.Packet {
	if state.Data == nil {
		return nil
	}
	if v, ok := state.Data[stateKeyOutFrags].([]*tlsfragment.Packet); ok {
		return v
	}
	return nil
}

func setFragments(state *eap.EAPState, frags []*tlsfragment.Packet) {
	if state.Data == nil {
		state.Data = make(map[string]interface{})
	}
	state.Data[stateKeyOutFrags] = frags
}

func getBool(state *eap.EAPState, key string) bool {
	if state.Data == nil {
		return false
	}
	v, _ := state.Data[key].(bool)
	return v
}

func setBool(state *eap.EAPState, key string, val bool) {
	if state.Data == nil {
		state.Data = make(map[string]interface{})
	}
	state.Data[key] = val
}

func clearKey(state *eap.EAPState, key string) {
	if state.Data != nil {
		delete(state.Data, key)
	}
}

func closeEngine(state *eap.EAPState) {
	if state.Data == nil {
		return
	}
	if eng, ok := state.Data[stateKeyEngine].(*tlsengine.Engine); ok && eng != nil {
		_ = eng.Close()
	}
	delete(state.Data, stateKeyEngine)
}
