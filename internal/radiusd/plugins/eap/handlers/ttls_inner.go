package handlers

import (
	"context"
	"crypto/subtle"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsengine"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/microsoft"
	"layeh.com/radius"
)

// ttlsMPPEKeyLabel is the RFC 5705 exporter label EAP-TTLS uses to derive its
// keying material from the TLS session (RFC 5281 §8: the keying material is
// generated with the ASCII label "ttls keying material"). The first 64 octets
// form the MSK, which ToughRADIUS splits into the MS-MPPE-Recv-Key (octets
// 0..31) and MS-MPPE-Send-Key (octets 32..63) per RFC 2548. This differs from
// PEAP's "client EAP encryption" label.
const ttlsMPPEKeyLabel = "ttls keying material"

// ttlsMSKLength is the number of MSK octets exported for the EAP-TTLS MPPE keys:
// 32 for each of the MS-MPPE-Recv-Key and MS-MPPE-Send-Key.
const ttlsMSKLength = 64

// handleInnerAVP runs EAP-TTLS phase 2 over the established TLS tunnel. EAP-TTLS
// is peer-initiated (RFC 5281 §7.3): the supplicant sends its inner
// authentication AVPs immediately after the outer handshake, so inner carries
// the decrypted AVP flight (RFC 5281 §10/§11) and the handler validates it
// without first emitting an inner request.
//
// Two inner methods are supported:
//
//   - PAP (RFC 5281 §11.2.5), a single round: the peer sends a User-Name and a
//     User-Password AVP; on a constant-time password match the MS-MPPE keys are
//     derived and the handler grants.
//   - MS-CHAP-V2 (RFC 5281 §11.2.4), two rounds: the peer sends User-Name,
//     MS-CHAP-Challenge and MS-CHAP2-Response AVPs validated against an
//     implicitly-derived challenge (see ttls_mschapv2.go); the handler tunnels
//     an MS-CHAP2-Success AVP and grants once the peer acknowledges with an empty
//     EAP-TTLS frame (inner == nil on that final round).
//
// On success the MS-MPPE-Send/Recv keys are always derived from the TLS session
// (RFC 5281 §8 / RFC 5705 / RFC 2548) — not from any inner secret — and added to
// the outer Access-Accept (ctx.Response). A malformed AVP flight rejects with
// eap.ErrTTLSInnerProtocol; a well-formed flight carrying neither a User-Password
// nor an MS-CHAP2-Response AVP (an inner CHAP / MS-CHAP / tunneled-EAP method,
// not yet supported) rejects with eap.ErrTTLSInnerNotImplemented; a password or
// response mismatch rejects with eap.ErrPasswordMismatch. The handler never
// reports success without a verified inner credential.
func (h *TTLSHandler) handleInnerAVP(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine, inner []byte) ([]byte, bool, error) {
	if getString(state, stateKeyInnerPhase) == ttlsInnerPhaseMSCHAPv2Ack {
		return h.handleMSCHAPv2Ack(ctx, state, engine, inner)
	}
	return h.handleInnerStart(ctx, state, engine, inner)
}

// handleInnerStart processes the peer's opening inner AVP flight and dispatches
// to the matching inner authentication method.
func (h *TTLSHandler) handleInnerStart(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine, inner []byte) ([]byte, bool, error) {
	if len(inner) == 0 {
		return nil, false, fmt.Errorf("%w: empty inner AVP flight", eap.ErrTTLSInnerProtocol)
	}

	avps, err := parseTTLSAVPs(inner)
	if err != nil {
		return nil, false, err
	}

	// Record the inner User-Name (the real identity inside the tunnel, which may
	// differ from an anonymous outer identity) for logging/context and for the
	// MS-CHAP-V2 username hash. Mapping it to a distinct user record for the
	// password lookup is deferred (mirroring PEAP), so the lookup still uses
	// ctx.User.
	if name, ok := findTTLSAVP(avps, ttlsAVPCodeUserName); ok {
		setString(state, stateKeyInnerIdentity, string(name))
	}

	if rawPwd, ok := findTTLSAVP(avps, ttlsAVPCodeUserPassword); ok {
		return h.handleInnerPAP(ctx, engine, rawPwd)
	}
	if _, ok := findTTLSVendorAVP(avps, ttlsVendorMicrosoft, ttlsMSCHAP2ResponseCode); ok {
		return h.handleInnerMSCHAPv2(ctx, state, engine, avps)
	}

	// Neither PAP nor MS-CHAP-V2: an inner CHAP / MS-CHAP / tunneled-EAP method
	// (RFC 5281 §11.2.2-§11.2.3 / §11.3), not yet supported. Reject explicitly
	// rather than granting.
	return nil, false, fmt.Errorf("%w: no User-Password or MS-CHAP2-Response AVP (only inner PAP and MS-CHAP-V2 are supported)", eap.ErrTTLSInnerNotImplemented)
}

// handleInnerPAP validates an inner PAP exchange (RFC 5281 §11.2.5). The
// User-Password AVP carries the cleartext password (the TLS tunnel, not RADIUS
// obfuscation, protects it), NUL-padded to a 16-octet multiple. The padding is
// stripped and the password verified: against an external directory by binding
// when an LDAP-style CredentialVerifier is active, otherwise in constant time
// against the locally configured password. On success the MS-MPPE keys are
// derived and the handler grants in a single round.
func (h *TTLSHandler) handleInnerPAP(ctx *eap.EAPContext, engine *tlsengine.Engine, rawPwd []byte) ([]byte, bool, error) {
	presented := stripTTLSPasswordPadding(rawPwd)

	if ctx.Verifier != nil && ctx.Verifier.Active() {
		if ctx.User == nil {
			return nil, false, fmt.Errorf("%w: missing user record for ldap bind", eap.ErrTTLSInnerProtocol)
		}
		bindCtx := ctx.Context
		if bindCtx == nil {
			bindCtx = context.Background()
		}
		if err := ctx.Verifier.VerifyCleartext(bindCtx, ctx.User.Username, string(presented)); err != nil {
			return nil, false, err
		}
		if err := h.deriveMPPEKeys(ctx.Response, engine); err != nil {
			return nil, false, err
		}
		return nil, true, nil
	}

	password, err := ctx.PwdProvider.GetPassword(ctx.User, ctx.IsMacAuth)
	if err != nil {
		return nil, false, err
	}

	if subtle.ConstantTimeCompare(presented, []byte(password)) != 1 {
		return nil, false, eap.ErrPasswordMismatch
	}

	if err := h.deriveMPPEKeys(ctx.Response, engine); err != nil {
		return nil, false, err
	}
	return nil, true, nil
}

// innerUsername returns the username used for inner credential checks: the inner
// User-Name AVP value when present, otherwise the outer User-Name. Mapping an
// anonymous outer identity to a distinct user record for the password lookup is
// deferred (mirroring PEAP).
func (h *TTLSHandler) innerUsername(ctx *eap.EAPContext, state *eap.EAPState) string {
	if id := getString(state, stateKeyInnerIdentity); id != "" {
		return id
	}
	if ctx.User != nil {
		return ctx.User.Username
	}
	return state.Username
}

// deriveMPPEKeys exports the EAP-TTLS MSK from the TLS session (RFC 5705 with the
// "ttls keying material" label, RFC 5281 §8) and adds the MS-MPPE-Recv-Key /
// MS-MPPE-Send-Key plus encryption policy to the outer Access-Accept (RFC 2548).
// Like PEAP, the keys come from the TLS exporter rather than the inner
// authentication secrets.
func (h *TTLSHandler) deriveMPPEKeys(response *radius.Packet, engine *tlsengine.Engine) error {
	msk, err := engine.ExportKey(ttlsMPPEKeyLabel, nil, ttlsMSKLength)
	if err != nil {
		return fmt.Errorf("failed to export EAP-TTLS MPPE keys: %w", err)
	}
	if len(msk) != ttlsMSKLength {
		return fmt.Errorf("unexpected EAP-TTLS MSK length: %d", len(msk))
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
