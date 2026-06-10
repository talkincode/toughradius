// Package handlers implements concrete EAP method handlers used by the
// coordinator in package eap.
//
// It contains classic challenge methods and TLS-based tunneled methods
// (EAP-TLS, PEAP, EAP-TTLS), including handshake progression, fragmentation,
// and method-specific success/failure contracts.
package handlers
