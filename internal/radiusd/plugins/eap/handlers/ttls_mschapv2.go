package handlers

import (
	"crypto/subtle"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsengine"
	"layeh.com/radius/rfc2759"
)

// EAP-TTLS inner MS-CHAP-V2 constants (RFC 5281 §11.2.4; AVP/attribute formats
// RFC 2548, algorithm RFC 2759). Inside the tunnel the MS-CHAP attributes are
// Diameter vendor AVPs (V flag set, Vendor-ID 311) whose AVP Code is the RADIUS
// vendor-type.
const (
	// ttlsVendorMicrosoft is Microsoft's SMI Network Management Private
	// Enterprise Code, the Vendor-ID of the MS-CHAP AVPs (RFC 2548 §1).
	ttlsVendorMicrosoft = 311

	// ttlsMSCHAPChallengeCode is the MS-CHAP-Challenge AVP code (RFC 2548 §2.1):
	// the server's implicitly-derived authenticator challenge echoed by the peer.
	ttlsMSCHAPChallengeCode = 11
	// ttlsMSCHAP2ResponseCode is the MS-CHAP2-Response AVP code (RFC 2548 §2.3.2):
	// the peer's Ident, Flags, Peer-Challenge, Reserved and NT-Response.
	ttlsMSCHAP2ResponseCode = 25
	// ttlsMSCHAP2SuccessCode is the MS-CHAP2-Success AVP code (RFC 2548 §2.3.3):
	// the Ident and the "S=" authenticator response the server tunnels back.
	ttlsMSCHAP2SuccessCode = 26

	// ttlsChallengeLabel is the TLS PRF / RFC 5705 exporter label EAP-TTLS uses to
	// derive the implicit challenge for legacy challenge-based inner methods
	// (RFC 5281 §11.1). It is distinct from the keying-material label so that the
	// challenge and the MSK are independent.
	ttlsChallengeLabel = "ttls challenge"
	// ttlsMSCHAPv2ChallengeLen is the number of implicit-challenge octets
	// MS-CHAP-V2 consumes: a 16-octet MS-CHAP-Challenge followed by a 1-octet
	// Ident (RFC 5281 §11.2.4).
	ttlsMSCHAPv2ChallengeLen = 17

	// ttlsMSCHAP2ResponseLen is the fixed MS-CHAP2-Response value length: Ident(1)
	// + Flags(1) + Peer-Challenge(16) + Reserved(8) + NT-Response(24)
	// (RFC 2548 §2.3.2).
	ttlsMSCHAP2ResponseLen = 50

	// ttlsInnerPhaseMSCHAPv2Ack marks the inner state awaiting the peer's empty
	// acknowledgement after an MS-CHAP2-Success has been tunneled (RFC 5281
	// §11.2.4). It is stored under stateKeyInnerPhase, which a given conversation
	// uses for either PEAP or EAP-TTLS but never both.
	ttlsInnerPhaseMSCHAPv2Ack = "ttls-mschapv2-ack"
)

// ttlsMSCHAP2Response is a decoded MS-CHAP2-Response AVP value (RFC 2548 §2.3.2).
type ttlsMSCHAP2Response struct {
	// Ident echoes the implicitly-derived challenge Ident octet.
	Ident uint8
	// Flags is reserved and set to zero by the peer (RFC 5281 §11.2.4).
	Flags uint8
	// PeerChallenge is the 16-octet peer challenge.
	PeerChallenge []byte
	// NTResponse is the 24-octet MS-CHAP-V2 NT-Response.
	NTResponse []byte
}

// parseTTLSMSCHAP2Response decodes a 50-octet MS-CHAP2-Response AVP value. The
// 8-octet Reserved field (RFC 2548 §2.3.2) is ignored.
func parseTTLSMSCHAP2Response(data []byte) (*ttlsMSCHAP2Response, error) {
	if len(data) != ttlsMSCHAP2ResponseLen {
		return nil, fmt.Errorf("%w: MS-CHAP2-Response is %d octets, expected %d",
			eap.ErrTTLSInnerProtocol, len(data), ttlsMSCHAP2ResponseLen)
	}
	return &ttlsMSCHAP2Response{
		Ident:         data[0],
		Flags:         data[1],
		PeerChallenge: data[2:18],
		NTResponse:    data[26:50],
	}, nil
}

// handleInnerMSCHAPv2 validates the peer's opening inner MS-CHAP-V2 flight
// (RFC 5281 §11.2.4) and, on success, tunnels an MS-CHAP2-Success AVP and waits
// for the peer's empty-frame acknowledgement before granting.
//
// To stop a client from choosing a weak challenge, both peers derive the
// 17-octet implicit challenge from the TLS session under the "ttls challenge"
// label (RFC 5281 §11.1): a 16-octet MS-CHAP-Challenge and a 1-octet Ident. The
// handler rejects unless the peer's MS-CHAP-Challenge AVP and the Ident in its
// MS-CHAP2-Response equal the derived values, then verifies the NT-Response with
// RFC 2759. The keys added on success come from the TLS exporter (deriveMPPEKeys,
// RFC 5281 §8), not from the MS-CHAP-V2 secrets.
func (h *TTLSHandler) handleInnerMSCHAPv2(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine, avps []ttlsAVP) ([]byte, bool, error) {
	chal, err := engine.ExportKey(ttlsChallengeLabel, nil, ttlsMSCHAPv2ChallengeLen)
	if err != nil {
		return nil, false, fmt.Errorf("failed to derive EAP-TTLS implicit challenge: %w", err)
	}
	authChallenge := chal[:16]
	expectedIdent := chal[16]

	mscChal, ok := findTTLSVendorAVP(avps, ttlsVendorMicrosoft, ttlsMSCHAPChallengeCode)
	if !ok {
		return nil, false, fmt.Errorf("%w: missing MS-CHAP-Challenge AVP", eap.ErrTTLSInnerProtocol)
	}
	if subtle.ConstantTimeCompare(mscChal, authChallenge) != 1 {
		return nil, false, fmt.Errorf("%w: MS-CHAP-Challenge AVP does not match the derived challenge", eap.ErrTTLSInnerProtocol)
	}

	respAVP, ok := findTTLSVendorAVP(avps, ttlsVendorMicrosoft, ttlsMSCHAP2ResponseCode)
	if !ok {
		return nil, false, fmt.Errorf("%w: missing MS-CHAP2-Response AVP", eap.ErrTTLSInnerProtocol)
	}
	resp, err := parseTTLSMSCHAP2Response(respAVP)
	if err != nil {
		return nil, false, err
	}
	if resp.Ident != expectedIdent {
		return nil, false, fmt.Errorf("%w: MS-CHAP2-Response Ident %d does not match the derived Ident %d",
			eap.ErrTTLSInnerProtocol, resp.Ident, expectedIdent)
	}

	username := h.innerUsername(ctx, state)
	password, err := ctx.PwdProvider.GetPassword(ctx.User, ctx.IsMacAuth)
	if err != nil {
		return nil, false, err
	}
	byteUser := []byte(username)
	bytePwd := []byte(password)

	expectedNT, err := rfc2759.GenerateNTResponse(authChallenge, resp.PeerChallenge, byteUser, bytePwd)
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate inner NT-Response: %w", err)
	}
	if subtle.ConstantTimeCompare(expectedNT, resp.NTResponse) != 1 {
		return nil, false, eap.ErrPasswordMismatch
	}

	authResp, err := rfc2759.GenerateAuthenticatorResponse(authChallenge, resp.PeerChallenge, expectedNT, byteUser, bytePwd)
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate inner authenticator response: %w", err)
	}

	// MS-CHAP2-Success value: the Ident octet followed by the "S=<40 hex>"
	// authenticator response string (RFC 2548 §2.3.3).
	success := make([]byte, 1+len(authResp))
	success[0] = resp.Ident
	copy(success[1:], authResp)
	reply := encodeTTLSAVP(ttlsMSCHAP2SuccessCode, ttlsVendorMicrosoft, true, success)

	// Authentication is not complete until the peer accepts the authenticator
	// response and acknowledges with an empty EAP-TTLS frame (RFC 5281 §11.2.4).
	setString(state, stateKeyInnerPhase, ttlsInnerPhaseMSCHAPv2Ack)
	return reply, false, nil
}

// handleMSCHAPv2Ack completes the inner MS-CHAP-V2 exchange. After the
// MS-CHAP2-Success AVP is tunneled the peer acknowledges with a zero-length
// EAP-TTLS frame (RFC 5281 §11.2.4), which the outer tunnel routes here as an
// empty inner flight; the handler then derives the MS-MPPE keys from the TLS
// session and grants. Any non-empty flight at this point is an unexpected
// continuation and rejects, so the handler never grants without first having
// verified the NT-Response.
func (h *TTLSHandler) handleMSCHAPv2Ack(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine, inner []byte) ([]byte, bool, error) {
	if len(inner) != 0 {
		return nil, false, fmt.Errorf("%w: expected an empty MS-CHAP-V2 acknowledgement", eap.ErrTTLSInnerProtocol)
	}
	if err := h.deriveMPPEKeys(ctx.Response, engine); err != nil {
		return nil, false, err
	}
	return nil, true, nil
}
