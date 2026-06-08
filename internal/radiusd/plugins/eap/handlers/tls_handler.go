package handlers

import (
"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
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
)

// TLSHandler is the EAP-TLS authentication handler.
//
// Milestone M1.2 adds TLS handshake state management and fragmentation
// reassembly (RFC 5216 §2.1.5 / RFC 7499): inbound EAP-TLS fragments are
// reassembled across RADIUS rounds, with fragment ACKs sent per the EAP
// ACK/NAK model, and the reassembly buffer persisted via the shared state
// manager. The TLS handshake engine, certificate validation and user-identity
// mapping are delivered in later M1 subtasks (M1.3); until then HandleResponse
// rejects safely once a complete TLS message has been assembled, so the handler
// can never authenticate a client yet.
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
// §2.1.5): while the peer signals more fragments (M bit) the handler stores the
// accumulated bytes and replies with a fragment ACK; once the final fragment
// arrives the full TLS message is available. The TLS handshake engine and
// certificate validation land in milestone M1.3, so for now a fully assembled
// message is rejected with an explicit reason rather than authenticated. This
// guarantees the handler never grants access before a validated handshake.
func (h *TLSHandler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
stateID := rfc2865.State_GetString(ctx.Request.Packet)
if stateID == "" {
return false, eap.ErrStateNotFound
}

state, err := ctx.StateManager.GetState(stateID)
if err != nil {
return false, err
}

frag, err := tlsfragment.Parse(ctx.EAPMessage.Data)
if err != nil {
return false, err
}

reassembler := loadReassembler(state)
complete, err := reassembler.Accept(frag)
if err != nil {
return false, err
}

if !complete {
// More fragments to come: persist progress and acknowledge this
// fragment so the peer sends the next one (RFC 5216 §2.1.5).
saveReassembler(state, reassembler)
if err := ctx.StateManager.SetState(stateID, state); err != nil {
return false, err
}

ackData := h.buildFragmentACK(ctx.EAPMessage.Identifier + 1)
return false, h.writeChallenge(ctx, stateID, ackData)
}

// A complete TLS message has been assembled. The TLS handshake engine and
// certificate validation are not implemented yet (milestone M1.3); reject
// safely so the handler can never authenticate a client.
return false, eap.ErrTLSHandshakeNotImplemented
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

// buildFragmentACK constructs an EAP-Request/EAP-TLS fragment acknowledgement: a
// payload with a single flags octet (all flags clear) and no TLS data
// (RFC 5216 §2.1.5).
func (h *TLSHandler) buildFragmentACK(identifier uint8) []byte {
return h.buildEAPRequest(identifier, &tlsfragment.Packet{})
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
