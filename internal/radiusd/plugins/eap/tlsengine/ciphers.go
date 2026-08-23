package tlsengine

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// Cipher-profile names accepted by ResolveCipherSuites. They are the values of
// the radius.EapTlsCipherProfile setting (TR-F004 / TR-F014).
const (
	// CipherProfileModern keeps crypto/tls default suites (ECDHE + AEAD). This
	// is the TR-F004 default: the outer tunnel is not weakened.
	CipherProfileModern = "modern"
	// CipherProfileLegacyRSACBC opts into the RSA+AES-CBC suites still spoken
	// by hostapd internal TLS / eapol_test and similar CBC-only PEAP peers
	// (issue #598). It has no forward secrecy and requires an RSA server
	// certificate. RC4, 3DES, and DHE are never included: Go does not
	// implement DHE, and RC4/3DES are forbidden. crypto/tls also does not
	// implement TLS_RSA_WITH_AES_256_CBC_SHA256 (0x003d).
	CipherProfileLegacyRSACBC = "legacy-rsa-cbc"
	// CipherProfileCustom uses the operator-supplied cipher list.
	CipherProfileCustom = "custom"
)

// Errors returned while resolving or validating a cipher policy.
var (
	// ErrUnknownCipherProfile is returned when the configured profile name is
	// not modern, legacy-rsa-cbc, or custom.
	ErrUnknownCipherProfile = errors.New("tlsengine: unknown cipher profile")
	// ErrEmptyCustomCipherSuites is returned when profile=custom but no suite
	// list was provided.
	ErrEmptyCustomCipherSuites = errors.New("tlsengine: custom cipher profile requires EapTlsCipherSuites")
	// ErrForbiddenCipherSuite is returned when a requested suite is RC4 or
	// 3DES, which this package never enables.
	ErrForbiddenCipherSuite = errors.New("tlsengine: RC4 and 3DES cipher suites are not allowed")
	// ErrUnknownCipherSuite is returned when a custom suite name or id is not
	// implemented by crypto/tls.
	ErrUnknownCipherSuite = errors.New("tlsengine: unknown or unimplemented cipher suite")
	// ErrLegacyNeedsRSACertificate is returned when a selected suite uses RSA
	// key exchange but the server certificate is not RSA (typically ECDSA).
	// The error is raised before the handshake so operators see a
	// configuration problem instead of "no cipher suite supported".
	ErrLegacyNeedsRSACertificate = errors.New("tlsengine: RSA key-exchange cipher suites require an RSA server certificate")
	// ErrLegacyIncompatibleTLS13 is returned when a non-empty TLS 1.2 cipher
	// list is combined with a TLS 1.3-only minimum version. TLS 1.3 ignores
	// CipherSuites, so the compatibility list would have no effect.
	ErrLegacyIncompatibleTLS13 = errors.New("tlsengine: TLS 1.2 cipher profiles cannot be used when the minimum version is TLS 1.3")
)

// LegacyRSACBCSuites is the opt-in TLS 1.2 list for CBC-only peers. Preference
// order puts SHA-256 variants first. These are the only suites from a typical
// hostapd internal-TLS ClientHello that crypto/tls can actually negotiate
// (DHE_* is unimplemented; RC4/3DES are refused).
var LegacyRSACBCSuites = []uint16{
	tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA,
}

// ResolveCipherSuites maps a profile name and optional custom list onto a
// crypto/tls CipherSuites value. An empty or "modern" profile returns nil so
// the engine keeps Go defaults. The custom list is a comma-separated set of
// IANA names (TLS_RSA_WITH_AES_128_CBC_SHA256) or 0xNNNN ids.
func ResolveCipherSuites(profile, custom string) ([]uint16, error) {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "", CipherProfileModern:
		return nil, nil
	case CipherProfileLegacyRSACBC:
		out := make([]uint16, len(LegacyRSACBCSuites))
		copy(out, LegacyRSACBCSuites)
		return out, nil
	case CipherProfileCustom:
		return parseCustomCipherSuites(custom)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownCipherProfile, profile)
	}
}

// ValidateCipherPolicy checks that a resolved suite list can actually be
// negotiated with the configured certificate and minimum TLS version. A nil
// suite list (modern) always passes.
func ValidateCipherPolicy(cert tls.Certificate, minVersion uint16, suites []uint16) error {
	if len(suites) == 0 {
		return nil
	}
	if minVersion == tls.VersionTLS13 {
		return ErrLegacyIncompatibleTLS13
	}
	if cipherListNeedsRSAKeyExchange(suites) && !certificateHasRSAKey(cert) {
		return ErrLegacyNeedsRSACertificate
	}
	return nil
}

func parseCustomCipherSuites(custom string) ([]uint16, error) {
	parts := strings.Split(custom, ",")
	out := make([]uint16, 0, len(parts))
	seen := make(map[uint16]struct{}, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		id, err := lookupCipherSuite(name)
		if err != nil {
			return nil, err
		}
		if cipherSuiteForbidden(id) {
			return nil, fmt.Errorf("%w: %s", ErrForbiddenCipherSuite, name)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil, ErrEmptyCustomCipherSuites
	}
	return out, nil
}

func lookupCipherSuite(name string) (uint16, error) {
	if id, ok := cipherSuiteIDs()[strings.ToUpper(name)]; ok {
		return id, nil
	}
	if strings.HasPrefix(name, "0x") || strings.HasPrefix(name, "0X") {
		raw, err := hex.DecodeString(name[2:])
		if err == nil && len(raw) == 2 {
			id := uint16(raw[0])<<8 | uint16(raw[1])
			if _, ok := implementedCipherSuiteIDs()[id]; ok {
				if cipherSuiteForbidden(id) {
					return 0, fmt.Errorf("%w: %s", ErrForbiddenCipherSuite, name)
				}
				return id, nil
			}
		}
	}
	return 0, fmt.Errorf("%w: %s", ErrUnknownCipherSuite, name)
}

func cipherSuiteIDs() map[string]uint16 {
	m := make(map[string]uint16)
	for _, s := range tls.CipherSuites() {
		m[s.Name] = s.ID
	}
	for _, s := range tls.InsecureCipherSuites() {
		m[s.Name] = s.ID
	}
	return m
}

func implementedCipherSuiteIDs() map[uint16]struct{} {
	m := make(map[uint16]struct{})
	for _, s := range tls.CipherSuites() {
		m[s.ID] = struct{}{}
	}
	for _, s := range tls.InsecureCipherSuites() {
		m[s.ID] = struct{}{}
	}
	return m
}

func cipherSuiteForbidden(id uint16) bool {
	switch id {
	case tls.TLS_RSA_WITH_RC4_128_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:
		return true
	default:
		return false
	}
}

func cipherListNeedsRSAKeyExchange(suites []uint16) bool {
	for _, id := range suites {
		if cipherSuiteUsesRSAKeyExchange(id) {
			return true
		}
	}
	return false
}

func cipherSuiteUsesRSAKeyExchange(id uint16) bool {
	switch id {
	case tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384:
		return true
	default:
		return false
	}
}

func certificateHasRSAKey(cert tls.Certificate) bool {
	if _, ok := cert.PrivateKey.(*rsa.PrivateKey); ok {
		return true
	}
	leaf := cert.Leaf
	if leaf == nil && len(cert.Certificate) > 0 {
		parsed, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return false
		}
		leaf = parsed
	}
	return leaf != nil && leaf.PublicKeyAlgorithm == x509.RSA
}
