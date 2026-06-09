package handlers

import (
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
// M9.3 implements inner PAP (RFC 5281 §11.2.5) only:
//
//	peer   -> User-Name AVP + User-Password AVP   (inside the TLS tunnel)
//	server -> Access-Accept + MS-MPPE keys         (success == true)
//
// The User-Password AVP carries the cleartext password (the TLS tunnel, not
// RADIUS obfuscation, protects it), NUL-padded to a 16-octet multiple; the
// padding is stripped and the value compared in constant time against the
// configured password. On success the MS-MPPE-Send/Recv keys are derived from
// the TLS session (RFC 5281 §8 / RFC 5705 / RFC 2548) and added to the outer
// Access-Accept (ctx.Response), and the handler reports success so the
// dispatcher grants access.
//
// A malformed AVP flight rejects with eap.ErrTTLSInnerProtocol; a well-formed
// flight that omits the User-Password AVP (an inner CHAP / MS-CHAP / MS-CHAP-V2
// method, scheduled for M9.4) rejects with eap.ErrTTLSInnerNotImplemented; a
// password mismatch rejects with eap.ErrPasswordMismatch. The handler never
// returns success without a User-Password AVP that matches the configured
// password.
func (h *TTLSHandler) handleInnerAVP(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine, inner []byte) ([]byte, bool, error) {
	if len(inner) == 0 {
		return nil, false, fmt.Errorf("%w: empty inner AVP flight", eap.ErrTTLSInnerProtocol)
	}

	avps, err := parseTTLSAVPs(inner)
	if err != nil {
		return nil, false, err
	}

	rawPwd, ok := findTTLSAVP(avps, ttlsAVPCodeUserPassword)
	if !ok {
		// Only PAP (a User-Password AVP) is supported in M9.3. An inner
		// CHAP/MS-CHAP/MS-CHAP-V2 exchange (RFC 5281 §11.2.2-§11.2.4) carries
		// other AVPs and lands in M9.4; reject it explicitly rather than
		// granting.
		return nil, false, fmt.Errorf("%w: no User-Password AVP (only inner PAP is supported)", eap.ErrTTLSInnerNotImplemented)
	}

	// Record the inner User-Name (the real identity inside the tunnel, which may
	// differ from an anonymous outer identity) for logging/context. Mapping it
	// to a distinct user record for the password lookup is deferred (mirroring
	// PEAP), so the lookup below still uses ctx.User.
	if name, ok := findTTLSAVP(avps, ttlsAVPCodeUserName); ok {
		setString(state, stateKeyInnerIdentity, string(name))
	}

	password, err := ctx.PwdProvider.GetPassword(ctx.User, ctx.IsMacAuth)
	if err != nil {
		return nil, false, err
	}

	presented := stripTTLSPasswordPadding(rawPwd)
	if subtle.ConstantTimeCompare(presented, []byte(password)) != 1 {
		return nil, false, eap.ErrPasswordMismatch
	}

	if err := h.deriveMPPEKeys(ctx.Response, engine); err != nil {
		return nil, false, err
	}
	return nil, true, nil
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
