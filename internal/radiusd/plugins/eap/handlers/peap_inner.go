package handlers

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsengine"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/microsoft"
	"layeh.com/radius"
	"layeh.com/radius/rfc2759"
)

// Inner EAP-MSCHAPv2 sub-states, persisted on the EAP state under
// stateKeyInnerPhase. They drive the tunneled exchange one RADIUS round at a
// time (PEAPv0 [MS-PEAP]; inner method RFC 2759).
const (
	// innerPhaseIdentity: a tunneled EAP-Request/Identity has been sent and the
	// peer's EAP-Response/Identity is awaited.
	innerPhaseIdentity = "identity"
	// innerPhaseChallenge: an EAP-Request/EAP-MSCHAPv2 Challenge has been sent
	// and the peer's EAP-Response/EAP-MSCHAPv2 Response is awaited.
	innerPhaseChallenge = "challenge"
	// innerPhaseSuccessAck: the NT-Response validated, an EAP-Request/
	// EAP-MSCHAPv2 Success was sent, and the peer's success acknowledgement is
	// awaited before the outer Access-Accept.
	innerPhaseSuccessAck = "success-ack"
)

// peapMPPEKeyLabel is the RFC 5705 exporter label used to derive the PEAP MSK
// from the TLS session (RFC 5216 §2.3). The 64-octet export is split into the
// MS-MPPE-Recv-Key (octets 0..31) and MS-MPPE-Send-Key (octets 32..63) per
// RFC 2548.
const peapMPPEKeyLabel = "client EAP encryption"

// peapMSKLength is the number of octets exported for the PEAP MSK: 32 for each
// of the MS-MPPE-Recv-Key and MS-MPPE-Send-Key.
const peapMSKLength = 64

// handleInnerEAP runs the PEAP inner EAP-MSCHAPv2 state machine over the
// established TLS tunnel. It is invoked by the tunnel's application-data phase:
// inner is the decrypted inbound inner EAP message (nil on the opening call,
// which asks the handler to emit the first tunneled request). It returns the
// next inner EAP message to send back through the tunnel (reply), whether
// authentication succeeded, or an error to reject.
//
// The exchange follows PEAPv0 with EAP-MSCHAPv2 as the inner method:
//
//	server -> EAP-Request/Identity
//	peer   -> EAP-Response/Identity (real username)
//	server -> EAP-Request/EAP-MSCHAPv2 Challenge   (RFC 2759 §4)
//	peer   -> EAP-Response/EAP-MSCHAPv2 Response
//	server -> EAP-Request/EAP-MSCHAPv2 Success     (validated)
//	peer   -> EAP-Response/EAP-MSCHAPv2 Success    (acknowledgement)
//	server -> Access-Accept + MS-MPPE keys         (success == true)
//
// On success the MS-MPPE-Send/Recv keys are derived from the TLS session (RFC
// 5705 exporter) — not from the inner MSCHAPv2 keys — and added to the outer
// Access-Accept (ctx.Response). A failed NT-Response validation returns
// eap.ErrPasswordMismatch so the dispatcher emits an Access-Reject.
func (h *PEAPHandler) handleInnerEAP(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine, inner []byte) ([]byte, bool, error) {
	var (
		reply   []byte
		success bool
		err     error
	)
	switch getString(state, stateKeyInnerPhase) {
	case "":
		reply, success, err = h.innerStartIdentity(state)
	case innerPhaseIdentity:
		reply, success, err = h.innerHandleIdentity(state, inner)
	case innerPhaseChallenge:
		reply, success, err = h.innerHandleChallengeResponse(ctx, state, inner)
	case innerPhaseSuccessAck:
		reply, success, err = h.innerHandleSuccessAck(ctx, state, engine, inner)
	default:
		return nil, false, eap.ErrPEAPInnerProtocol
	}
	if err != nil || success || len(reply) == 0 {
		return reply, success, err
	}
	// PEAPv0 (hostapd / Microsoft [MS-PEAP]): inner method payloads other
	// than EAP-Request/Identity are Type+Data only. A full EAP header is
	// re-wrapped by the peer using the outer identifier, which would make
	// an EAP-MSCHAPv2 Challenge look like another Identity request.
	return peapv0WirePayload(reply), false, nil
}

// peapv0WirePayload converts a full inner EAP packet to the PEAPv0 on-the-wire
// form. EAP-Request/Identity keeps its 5-octet header (hostapd special-cases
// that exact frame); every other method is reduced to Type+Data.
func peapv0WirePayload(inner []byte) []byte {
	if len(inner) >= 5 && looksLikeFullEAP(inner) && inner[4] == eap.TypeIdentity {
		return inner
	}
	if len(inner) >= 5 && looksLikeFullEAP(inner) {
		return append([]byte(nil), inner[4:]...)
	}
	return inner
}

// innerStartIdentity emits the opening tunneled EAP-Request/Identity.
func (h *PEAPHandler) innerStartIdentity(state *eap.EAPState) ([]byte, bool, error) {
	id := uint8(1)
	setUint8(state, stateKeyInnerID, id)
	setString(state, stateKeyInnerPhase, innerPhaseIdentity)
	return buildInnerIdentityRequest(id), false, nil
}

// innerHandleIdentity consumes the peer's EAP-Response/Identity, captures the
// inner username, and sends the EAP-MSCHAPv2 Challenge.
func (h *PEAPHandler) innerHandleIdentity(state *eap.EAPState, inner []byte) ([]byte, bool, error) {
	msg, err := parseInnerEAP(inner)
	if err != nil {
		return nil, false, err
	}
	if msg.Code != eap.CodeResponse || msg.Type != eap.TypeIdentity {
		return nil, false, fmt.Errorf("%w: expected inner EAP-Response/Identity", eap.ErrPEAPInnerProtocol)
	}
	if identity := string(msg.Data); identity != "" {
		setString(state, stateKeyInnerIdentity, identity)
	}

	challenge, err := eap.GenerateRandomBytes(MSCHAPChallengeSize)
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate inner challenge: %w", err)
	}
	state.Challenge = challenge

	id := getUint8(state, stateKeyInnerID) + 1
	setUint8(state, stateKeyInnerID, id)
	setString(state, stateKeyInnerPhase, innerPhaseChallenge)

	// buildChallengeRequest emits a complete inner EAP-Request/EAP-MSCHAPv2
	// Challenge packet, reused unchanged from the native EAP-MSCHAPv2 handler.
	return (&MSCHAPv2Handler{}).buildChallengeRequest(id, challenge), false, nil
}

// innerHandleChallengeResponse validates the peer's EAP-MSCHAPv2 Response
// (RFC 2759) and, on success, sends the EAP-MSCHAPv2 Success request carrying
// the Authenticator Response.
func (h *PEAPHandler) innerHandleChallengeResponse(ctx *eap.EAPContext, state *eap.EAPState, inner []byte) ([]byte, bool, error) {
	msg, err := parseInnerEAP(inner)
	if err != nil {
		return nil, false, err
	}
	if msg.Code != eap.CodeResponse || msg.Type != eap.TypeMSCHAPv2 {
		return nil, false, fmt.Errorf("%w: expected inner EAP-Response/EAP-MSCHAPv2", eap.ErrPEAPInnerProtocol)
	}

	msResp, err := (&MSCHAPv2Handler{}).parseResponse(msg.Data)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", eap.ErrPEAPInnerProtocol, err)
	}

	username := h.innerUsername(ctx, state)
	password, err := ctx.PwdProvider.GetPassword(ctx.User, ctx.IsMacAuth)
	if err != nil {
		return nil, false, err
	}

	byteUser := []byte(username)
	bytePwd := []byte(password)

	expectedNT, err := rfc2759.GenerateNTResponse(state.Challenge, msResp.PeerChallenge, byteUser, bytePwd)
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate inner NT-Response: %w", err)
	}
	if !bytes.Equal(expectedNT, msResp.NTResponse) {
		return nil, false, eap.ErrPasswordMismatch
	}

	authResp, err := rfc2759.GenerateAuthenticatorResponse(
		state.Challenge, msResp.PeerChallenge, expectedNT, byteUser, bytePwd)
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate inner authenticator response: %w", err)
	}

	id := getUint8(state, stateKeyInnerID) + 1
	setUint8(state, stateKeyInnerID, id)
	setString(state, stateKeyInnerPhase, innerPhaseSuccessAck)

	return buildInnerMSCHAPv2Success(id, msResp.MsIdentifier, authResp), false, nil
}

// innerHandleSuccessAck consumes the peer's EAP-MSCHAPv2 Success
// acknowledgement, derives the MS-MPPE keys from the TLS session, and reports
// success so the dispatcher emits the outer Access-Accept.
func (h *PEAPHandler) innerHandleSuccessAck(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine, inner []byte) ([]byte, bool, error) {
	if !innerSuccessAcknowledged(inner) {
		return nil, false, fmt.Errorf("%w: expected inner success acknowledgement", eap.ErrPEAPInnerProtocol)
	}

	if err := h.deriveMPPEKeys(ctx.Response, engine); err != nil {
		return nil, false, err
	}
	return nil, true, nil
}

// innerUsername returns the username used for the inner MSCHAPv2 computation:
// the identity carried in the inner EAP-Response/Identity when present,
// otherwise the outer User-Name. Mapping an anonymous outer identity to a user
// record for the password lookup is deferred to M8.4.
func (h *PEAPHandler) innerUsername(ctx *eap.EAPContext, state *eap.EAPState) string {
	if id := getString(state, stateKeyInnerIdentity); id != "" {
		return id
	}
	if ctx.User != nil {
		return ctx.User.Username
	}
	return state.Username
}

// deriveMPPEKeys exports the PEAP MSK from the TLS session (RFC 5705) and adds
// the MS-MPPE-Recv-Key / MS-MPPE-Send-Key plus encryption policy to the outer
// Access-Accept (RFC 2548). Unlike native EAP-MSCHAPv2 the keys come from the
// TLS exporter, not the inner MSCHAPv2 secrets.
func (h *PEAPHandler) deriveMPPEKeys(response *radius.Packet, engine *tlsengine.Engine) error {
	msk, err := engine.ExportKey(peapMPPEKeyLabel, nil, peapMSKLength)
	if err != nil {
		return fmt.Errorf("failed to export PEAP MPPE keys: %w", err)
	}
	if len(msk) != peapMSKLength {
		return fmt.Errorf("unexpected PEAP MSK length: %d", len(msk))
	}

	recvKey := msk[:32]
	sendKey := msk[32:64]

	_ = microsoft.MSMPPERecvKey_Add(response, recvKey) //nolint:errcheck
	_ = microsoft.MSMPPESendKey_Add(response, sendKey) //nolint:errcheck
	_ = microsoft.MSMPPEEncryptionPolicy_Add(response, //nolint:errcheck
		microsoft.MSMPPEEncryptionPolicy_Value_EncryptionAllowed)
	_ = microsoft.MSMPPEEncryptionTypes_Add(response, //nolint:errcheck
		microsoft.MSMPPEEncryptionTypes_Value_RC440or128BitAllowed)
	return nil
}

// buildInnerIdentityRequest builds a tunneled EAP-Request/Identity packet.
func buildInnerIdentityRequest(identifier uint8) []byte {
	return (&eap.EAPMessage{
		Code:       eap.CodeRequest,
		Identifier: identifier,
		Type:       eap.TypeIdentity,
	}).Encode()
}

// buildInnerMSCHAPv2Success builds a tunneled EAP-Request/EAP-MSCHAPv2 Success
// packet (RFC 2759 §5). The message body is the Authenticator Response string
// ("S=<40 hex>") generated during validation.
func buildInnerMSCHAPv2Success(identifier, msIdentifier uint8, authResponse string) []byte {
	body := []byte(authResponse)
	// OpCode(1) + MS-CHAPv2-ID(1) + MS-Length(2) + Message
	msLen := 4 + len(body)
	data := make([]byte, msLen)
	data[0] = MSCHAPv2Success
	data[1] = msIdentifier
	binary.BigEndian.PutUint16(data[2:4], uint16(msLen)) //nolint:gosec // G115: msLen is bounded by EAP packet size
	copy(data[4:], body)

	return (&eap.EAPMessage{
		Code:       eap.CodeRequest,
		Identifier: identifier,
		Type:       eap.TypeMSCHAPv2,
		Data:       data,
	}).Encode()
}

// parseInnerEAP parses a raw inner EAP packet (the decrypted tunnel plaintext)
// into its header fields and method data.
//
// PEAPv0 peers (hostapd / eapol_test / some Windows stacks) do not always send
// a textbook EAP header as TLS application data. normalizeInnerEAP accepts the
// common on-the-wire variants before the length check so a framing difference
// rejects as a protocol error only when the payload is genuinely not EAP.
func parseInnerEAP(data []byte) (*eap.EAPMessage, error) {
	data = normalizeInnerEAP(data)
	if len(data) < 4 {
		return nil, fmt.Errorf("%w: inner EAP packet too short (%d bytes)", eap.ErrInvalidEAPMessage, len(data))
	}
	declared := binary.BigEndian.Uint16(data[2:4])
	if int(declared) != len(data) {
		return nil, fmt.Errorf("%w: inner EAP length %d != %d", eap.ErrInvalidEAPMessage, declared, len(data))
	}
	// RFC 3748 §4.1: Success/Failure have no Type/Type-Data (Length=4).
	if declared == 4 {
		return &eap.EAPMessage{
			Code:       data[0],
			Identifier: data[1],
			Length:     declared,
		}, nil
	}
	if len(data) < 5 {
		return nil, fmt.Errorf("%w: inner EAP packet too short (%d bytes)", eap.ErrInvalidEAPMessage, len(data))
	}
	return &eap.EAPMessage{
		Code:       data[0],
		Identifier: data[1],
		Length:     declared,
		Type:       data[4],
		Data:       data[5:],
	}, nil
}

// normalizeInnerEAP rewrites decrypted PEAP phase-2 plaintext into a full EAP
// packet. It only strips a 4-octet length prefix when that prefix exactly
// covers the remainder, truncates trailing padding when the EAP Length field
// is already plausible, and reconstructs a Response header for compressed
// Type+Data payloads used by some PEAPv0 peers.
func normalizeInnerEAP(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	if len(data) >= 9 {
		prefix := binary.BigEndian.Uint32(data[:4])
		if int(prefix) == len(data)-4 && prefix >= 5 {
			data = data[4:]
		}
	}
	if looksLikeFullEAP(data) {
		declared := binary.BigEndian.Uint16(data[2:4])
		return data[:declared]
	}
	if isInnerMethodType(data[0]) {
		return (&eap.EAPMessage{
			Code:       eap.CodeResponse,
			Identifier: 0,
			Type:       data[0],
			Data:       append([]byte(nil), data[1:]...),
		}).Encode()
	}
	return data
}

func looksLikeFullEAP(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	code := data[0]
	if code < eap.CodeRequest || code > eap.CodeFailure {
		return false
	}
	declared := int(binary.BigEndian.Uint16(data[2:4]))
	return declared >= 4 && declared <= len(data)
}

func isInnerMethodType(typ uint8) bool {
	switch typ {
	case eap.TypeIdentity, eap.TypeNak, eap.TypeMSCHAPv2, eap.TypeTLV:
		return true
	default:
		return false
	}
}

// innerSuccessAcknowledged reports whether decrypted phase-2 plaintext is a
// PEAP inner success acknowledgement. Accepted forms:
//
//   - empty application data (bare ACK after the MSCHAPv2 Success request)
//   - EAP-Response/EAP-MSCHAPv2 with Success opcode (RFC 2759 §5)
//   - EAP-TLV Result=Success (Microsoft [MS-PEAP] §2.2.8)
//   - tunneled EAP-Success (code 3)
func innerSuccessAcknowledged(inner []byte) bool {
	if len(inner) == 0 {
		return true
	}
	msg, err := parseInnerEAP(inner)
	if err != nil {
		return false
	}
	if msg.Code == eap.CodeSuccess {
		return true
	}
	if msg.Code == eap.CodeResponse && msg.Type == eap.TypeMSCHAPv2 && len(msg.Data) > 0 && msg.Data[0] == MSCHAPv2Success {
		return true
	}
	if msg.Code == eap.CodeResponse && msg.Type == eap.TypeTLV {
		return tlvResultSuccess(msg.Data)
	}
	return false
}

// tlvResultSuccess reports whether data is an EAP-TLV payload containing a
// Result TLV with status Success (1). Unknown TLVs (including Cryptobinding)
// are skipped. A Result=Failure TLV is not success.
func tlvResultSuccess(data []byte) bool {
	const (
		tlvTypeResult    = 3
		tlvStatusSuccess = 1
	)
	i := 0
	for i+4 <= len(data) {
		typ := binary.BigEndian.Uint16(data[i:i+2]) & 0x3fff
		n := int(binary.BigEndian.Uint16(data[i+2 : i+4]))
		i += 4
		if i+n > len(data) {
			return false
		}
		if typ == tlvTypeResult && n >= 2 {
			return binary.BigEndian.Uint16(data[i:i+2]) == tlvStatusSuccess
		}
		i += n
	}
	return false
}
