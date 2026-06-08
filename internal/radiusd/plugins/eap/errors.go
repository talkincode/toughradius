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
)
