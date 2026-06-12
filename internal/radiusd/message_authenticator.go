package radiusd

import (
	"crypto/hmac"
	"crypto/md5"
	"strings"

	"github.com/talkincode/toughradius/v9/internal/app"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2869"
)

// Message-Authenticator (RFC 3579 §3.2) enforcement modes. They are configured
// via the radius.RequireMessageAuthenticator setting and govern how non-EAP
// Access-Request packets are validated and how Access-Accept/Access-Reject
// responses are signed. This is the RADIUS/UDP hardening lever for CVE-2024-3596
// ("BlastRADIUS"); full protection still requires RadSec or IPsec.
const (
	// MsgAuthModeDisabled keeps the legacy behavior: incoming packets are not
	// validated and outgoing non-EAP responses are left unsigned.
	MsgAuthModeDisabled = "disabled"
	// MsgAuthModeWarn signs every non-EAP response, validates the
	// Message-Authenticator when a NAS includes it (a mismatch is discarded),
	// and only logs a warning when the attribute is missing.
	MsgAuthModeWarn = "warn"
	// MsgAuthModeEnforce behaves like MsgAuthModeWarn but additionally discards
	// Access-Request packets that omit the Message-Authenticator attribute.
	MsgAuthModeEnforce = "enforce"
)

// messageAuthResult reports the outcome of validating the optional RFC 3579
// Message-Authenticator attribute on an incoming Access-Request.
type messageAuthResult int

const (
	// msgAuthAbsent indicates the request carried no Message-Authenticator.
	msgAuthAbsent messageAuthResult = iota
	// msgAuthValid indicates the Message-Authenticator matched the shared secret.
	msgAuthValid
	// msgAuthInvalid indicates the Message-Authenticator did not match, was
	// malformed, or appeared more than once.
	msgAuthInvalid
)

// MessageAuthenticatorMode resolves the configured radius.RequireMessageAuthenticator
// setting, defaulting to MsgAuthModeWarn for any unrecognized value.
func (s *RadiusService) MessageAuthenticatorMode() string {
	cfgMgr := s.appCtx.ConfigMgr()
	if cfgMgr == nil {
		return MsgAuthModeWarn
	}
	switch strings.ToLower(strings.TrimSpace(cfgMgr.GetString("radius", "RequireMessageAuthenticator"))) {
	case MsgAuthModeDisabled:
		return MsgAuthModeDisabled
	case MsgAuthModeEnforce:
		return MsgAuthModeEnforce
	case MsgAuthModeWarn:
		return MsgAuthModeWarn
	default:
		return MsgAuthModeWarn
	}
}

// computeMessageAuthenticator returns the RFC 3579 HMAC-MD5 Message-Authenticator
// for the wire-format RADIUS packet, treating every Message-Authenticator
// attribute value (type 80, length 18) as sixteen zero octets during the
// computation. The shared secret is used as the HMAC key. The supplied buffer is
// copied so the caller's data is never mutated.
func computeMessageAuthenticator(wire, secret []byte) []byte {
	buf := make([]byte, len(wire))
	copy(buf, wire)

	// Attributes start after the 20-byte header (Code, Identifier, Length, and
	// the 16-byte Authenticator). Walk the TLV list and zero the value of any
	// Message-Authenticator attribute in place before hashing.
	for i := 20; i+2 <= len(buf); {
		attrLen := int(buf[i+1])
		if attrLen < 2 || i+attrLen > len(buf) {
			break
		}
		if buf[i] == byte(rfc2869.MessageAuthenticator_Type) && attrLen == 18 {
			for j := i + 2; j < i+attrLen; j++ {
				buf[j] = 0
			}
		}
		i += attrLen
	}

	mac := hmac.New(md5.New, secret)
	mac.Write(buf)
	return mac.Sum(nil)
}

// verifyMessageAuthenticator validates the Message-Authenticator attribute of an
// incoming Access-Request against the shared secret. It returns msgAuthAbsent
// when the attribute is not present so the caller can apply the configured
// warn/enforce policy.
func (s *RadiusService) verifyMessageAuthenticator(r *radius.Packet, secret []byte) messageAuthResult {
	if len(secret) == 0 {
		return msgAuthInvalid
	}

	values, err := rfc2869.MessageAuthenticator_Gets(r)
	if err != nil {
		return msgAuthInvalid
	}
	switch len(values) {
	case 0:
		return msgAuthAbsent
	case 1:
		// single attribute, validated below
	default:
		// RFC 3579 §3.2 allows at most one Message-Authenticator; reject extras.
		return msgAuthInvalid
	}

	received := values[0]
	if len(received) != md5.Size {
		return msgAuthInvalid
	}

	wire, err := r.MarshalBinary()
	if err != nil {
		return msgAuthInvalid
	}

	expected := computeMessageAuthenticator(wire, secret)
	if hmac.Equal(received, expected) {
		return msgAuthValid
	}
	return msgAuthInvalid
}

// verifyResponseMessageAuthenticator validates the OPTIONAL Message-Authenticator
// attribute (RFC 3579 §3.2) that a NAS may include on a CoA/Disconnect reply
// (ACK or NAK), per RFC 5176 §3.4. A Dynamic Authorization Client that receives a
// reply carrying the attribute MUST recompute it and silently discard the packet
// on mismatch; an absent attribute is accepted because it is optional on
// responses.
//
// It returns msgAuthAbsent when no attribute is present (accept), msgAuthValid
// when it verifies, and msgAuthInvalid when the reply must be discarded (wrong
// value, malformed, duplicated, or unverifiable because no secret is available).
//
// Per RFC 5176 §3.4 the reply digest is HMAC-MD5 over the response packet with
// the Message-Authenticator value treated as sixteen octets of zero and the
// Authenticator field set to the Request Authenticator of the corresponding
// request (reqAuth). A parsed reply carries the Response Authenticator in its
// Authenticator field, so that field is overwritten with reqAuth before hashing.
//
// Presence is checked before the secret so an unsigned reply is still accepted
// even when the shared secret is missing; a signed reply with no secret to verify
// against fails closed.
func verifyResponseMessageAuthenticator(resp *radius.Packet, reqAuth [16]byte, secret []byte) messageAuthResult {
	if resp == nil {
		return msgAuthAbsent
	}

	values, err := rfc2869.MessageAuthenticator_Gets(resp)
	if err != nil {
		return msgAuthInvalid
	}
	switch len(values) {
	case 0:
		return msgAuthAbsent
	case 1:
		// single attribute, validated below
	default:
		// RFC 3579 §3.2 allows at most one Message-Authenticator; reject extras.
		return msgAuthInvalid
	}

	if len(secret) == 0 {
		return msgAuthInvalid
	}

	received := values[0]
	if len(received) != md5.Size {
		return msgAuthInvalid
	}

	wire, err := resp.MarshalBinary()
	if err != nil || len(wire) < 20 {
		return msgAuthInvalid
	}
	// RFC 5176 §3.4: the reply digest is computed with the Request Authenticator
	// of the corresponding request, not the Response Authenticator carried on the
	// reply, so overwrite the marshaled Authenticator field before hashing.
	copy(wire[4:20], reqAuth[:])

	expected := computeMessageAuthenticator(wire, secret)
	if hmac.Equal(received, expected) {
		return msgAuthValid
	}
	return msgAuthInvalid
}

// addResponseMessageAuthenticator signs a non-EAP Access-Accept/Access-Reject
// response with a freshly computed Message-Authenticator. It must be called
// before the response is written so the attribute is covered by the Response
// Authenticator (RFC 2869 §5.14). It is a no-op when signing is disabled or no
// usable shared secret is available.
func (s *RadiusService) addResponseMessageAuthenticator(resp *radius.Packet, secret string) {
	if resp == nil {
		return
	}
	if s.MessageAuthenticatorMode() == MsgAuthModeDisabled {
		return
	}
	if secret == "" || secret == unknownNasSecret {
		return
	}

	// Zero the attribute first so it is excluded from its own HMAC, then hash the
	// packet and store the result.
	if err := rfc2869.MessageAuthenticator_Set(resp, make([]byte, md5.Size)); err != nil {
		zap.L().Error("failed to reset message-authenticator",
			zap.String("namespace", "radius"),
			zap.Error(err),
		)
		return
	}
	wire, err := resp.MarshalBinary()
	if err != nil {
		zap.L().Error("failed to marshal response for message-authenticator",
			zap.String("namespace", "radius"),
			zap.Error(err),
		)
		return
	}
	mac := computeMessageAuthenticator(wire, []byte(secret))
	if err := rfc2869.MessageAuthenticator_Set(resp, mac); err != nil {
		zap.L().Error("failed to set message-authenticator",
			zap.String("namespace", "radius"),
			zap.Error(err),
		)
	}
}

// messageAuthDecision maps a validation result and the configured mode to the
// resulting action. discard reports that the Access-Request must be silently
// dropped; warnMissing reports that an absent (but tolerated) attribute should
// be logged. It is the pure policy core of enforceMessageAuthenticator.
func messageAuthDecision(mode string, result messageAuthResult) (discard, warnMissing bool) {
	if mode == MsgAuthModeDisabled {
		return false, false
	}
	switch result {
	case msgAuthInvalid:
		// A mismatch indicates a wrong shared secret or active tampering, so it
		// is discarded in both warn and enforce modes per RFC 3579 §3.2.
		return true, false
	case msgAuthAbsent:
		if mode == MsgAuthModeEnforce {
			return true, false
		}
		return false, true
	default: // msgAuthValid
		return false, false
	}
}

// enforceMessageAuthenticator validates the incoming Access-Request according to
// the configured mode and reports whether the request should be silently
// discarded (RFC 3579 §3.2). When discard is true the pipeline must stop without
// writing any response.
func (s *AuthService) enforceMessageAuthenticator(ctx *AuthPipelineContext) (discard bool) {
	mode := s.MessageAuthenticatorMode()
	if mode == MsgAuthModeDisabled || ctx.NAS == nil {
		return false
	}

	result := s.verifyMessageAuthenticator(ctx.Request.Packet, []byte(ctx.NAS.Secret))
	drop, warnMissing := messageAuthDecision(mode, result)
	if drop {
		reason := "invalid message-authenticator"
		if result == msgAuthAbsent {
			reason = "missing message-authenticator"
		}
		s.logMessageAuthDiscard(ctx, reason)
		return true
	}
	if warnMissing {
		zap.L().Warn("radius access-request without message-authenticator",
			zap.String("namespace", "radius"),
			zap.String("username", ctx.Username),
			zap.String("nasip", ctx.RemoteIP),
			zap.String("mode", mode),
		)
	}
	return false
}

// logMessageAuthDiscard records a silently discarded Access-Request and bumps the
// dedicated rejection metric.
func (s *AuthService) logMessageAuthDiscard(ctx *AuthPipelineContext, reason string) {
	app.IncRadiusMetric(app.MetricsRadiusRejectMsgAuth)
	zap.L().Warn("radius access-request discarded",
		zap.String("namespace", "radius"),
		zap.String("reason", reason),
		zap.String("username", ctx.Username),
		zap.String("nasip", ctx.RemoteIP),
		zap.String("metrics", app.MetricsRadiusRejectMsgAuth),
	)
}
