package eap

import "errors"

// EAP-related errors
var (
	ErrInvalidEAPMessage    = errors.New("invalid EAP message")
	ErrStateNotFound        = errors.New("EAP state not found")
	ErrPasswordMismatch     = errors.New("password mismatch")
	ErrUnsupportedEAPType   = errors.New("unsupported EAP type")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrOTPNotConfigured     = errors.New("EAP-OTP validation is not configured")
	// ErrTLSHandshakeNotImplemented is returned by the EAP-TLS handler skeleton
	// until the TLS handshake / fragmentation logic is implemented (milestone
	// M1.2). It guarantees the skeleton never grants access.
	ErrTLSHandshakeNotImplemented = errors.New("EAP-TLS handshake is not implemented yet")
	// ErrTLSNotConfigured is returned by the EAP-TLS handler when no server
	// certificate / client CA material has been provided, so the TLS handshake
	// cannot be driven. It guarantees the handler never authenticates a client
	// without configured trust anchors.
	ErrTLSNotConfigured = errors.New("EAP-TLS is not configured (missing server certificate or client CA)")
	// ErrTLSHandshakeFailed is returned when the TLS handshake fails, including
	// when the peer certificate does not chain to a configured CA (RFC 5216
	// §2.2 / §5.3).
	ErrTLSHandshakeFailed = errors.New("EAP-TLS handshake failed")
	// ErrTLSIdentityMismatch is returned when the authenticated certificate
	// identity does not match the RADIUS User-Name (RFC 5216 §5.2 identity
	// mapping).
	ErrTLSIdentityMismatch = errors.New("EAP-TLS certificate identity does not match the requested user")
	// ErrTLSNoIdentity is returned when no usable identity can be derived from
	// the verified peer certificate.
	ErrTLSNoIdentity = errors.New("EAP-TLS peer certificate carries no usable identity")
	// ErrTLSUnexpectedFragment is returned when a peer response violates the
	// EAP-TLS fragmentation exchange (e.g. a non-ACK arrives while the server is
	// still sending fragments, RFC 5216 §2.1.5).
	ErrTLSUnexpectedFragment = errors.New("EAP-TLS unexpected fragment in handshake exchange")
	// ErrPEAPInnerNotImplemented is returned after PEAP's outer TLS tunnel has
	// been established but before the inner EAP method is implemented (milestone
	// M8.3). It guarantees M8.2 can never grant access.
	ErrPEAPInnerNotImplemented = errors.New("PEAP inner EAP method is not implemented yet")
	// ErrPEAPInnerProtocol is returned when the inner EAP-MSCHAPv2 exchange
	// carried inside the established PEAP tunnel violates the expected protocol
	// (unexpected EAP code/type/opcode for the current inner sub-state). It
	// guarantees a malformed inner exchange rejects rather than grants.
	ErrPEAPInnerProtocol = errors.New("PEAP inner EAP-MSCHAPv2 protocol violation")
	// ErrTTLSInnerNotImplemented is returned after EAP-TTLS's outer TLS tunnel
	// has been established but before the inner AVP authentication (PAP /
	// MS-CHAP-V2) is implemented (milestones M9.3 / M9.4). It guarantees the
	// M9.2 outer tunnel can establish and fragment correctly yet never grant
	// access (RFC 5281 §10/§11 inner authentication is pending).
	ErrTTLSInnerNotImplemented = errors.New("EAP-TTLS inner authentication is not implemented yet")
)
