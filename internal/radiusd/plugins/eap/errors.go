package eap

import "errors"

// EAP 相关错误
var (
	ErrInvalidEAPMessage    = errors.New("invalid EAP message")
	ErrStateNotFound        = errors.New("EAP state not found")
	ErrPasswordMismatch     = errors.New("password mismatch")
	ErrUnsupportedEAPType   = errors.New("unsupported EAP type")
	ErrAuthenticationFailed = errors.New("authentication failed")
)
