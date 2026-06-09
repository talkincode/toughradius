package handlers

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsengine"
)

// EAP-TLS dynamic configuration keys (RADIUS settings category). They are
// declared in internal/app/config_schemas.json and editable on the system
// config page, so certificate material can be rotated without restarting the
// RADIUS service (milestone M1.5, TR-F004 / TR-F014).
const (
	// SettingEapTlsCertFile is the PEM server-certificate path presented to the
	// EAP-TLS peer (RFC 5216 §2.2).
	SettingEapTlsCertFile = "EapTlsCertFile"
	// SettingEapTlsKeyFile is the PEM private-key path matching the server
	// certificate.
	SettingEapTlsKeyFile = "EapTlsKeyFile"
	// SettingEapTlsCaFile is the PEM CA bundle path used to verify the peer's
	// client-certificate chain (RFC 5216 §2.2 / §5.3).
	SettingEapTlsCaFile = "EapTlsCaFile"
	// SettingEapTlsMinVersion pins the minimum negotiated TLS version.
	SettingEapTlsMinVersion = "EapTlsMinVersion"
)

// TLSSettingsReader reads EAP-TLS runtime configuration values. It is satisfied
// by *app.ConfigManager (its GetString method), letting the handler pick up
// certificate/CA changes between handshakes without a restart while keeping the
// handlers package free of an import dependency on internal/app.
type TLSSettingsReader interface {
	GetString(category, name string) string
}

// NewSettingsTLSConfigProvider returns a TLSConfigProvider that assembles the
// EAP-TLS material from dynamic settings on every handshake.
//
// EAP-TLS requires the server to present a certificate and to authenticate the
// peer against a trusted CA (RFC 5216 §2.2). Until the certificate, key, and CA
// bundle paths are all configured, the provider returns a nil config so the
// handler rejects safely with eap.ErrTLSNotConfigured and can never
// authenticate a client without configured trust anchors. When the paths are
// set but the material fails to load, it returns an explicit error so the
// failure reason is surfaced rather than silently ignored.
func NewSettingsTLSConfigProvider(reader TLSSettingsReader) TLSConfigProvider {
	return func() (*tlsengine.Config, error) {
		if reader == nil {
			return nil, nil
		}

		certFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsCertFile))
		keyFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsKeyFile))
		caFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsCaFile))

		// EAP-TLS is opt-in: treat it as not configured (safe reject) until the
		// server certificate/key and the client CA bundle are all provided.
		if certFile == "" || keyFile == "" || caFile == "" {
			return nil, nil
		}

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("load EAP-TLS server certificate: %w", err)
		}

		caPEM, err := os.ReadFile(caFile) //nolint:gosec // G304: path is from validated config
		if err != nil {
			return nil, fmt.Errorf("read EAP-TLS client CA bundle: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("parse EAP-TLS client CA bundle %q: no PEM certificates found", caFile)
		}

		return &tlsengine.Config{
			ServerCertificate: cert,
			ClientCAs:         pool,
			MinVersion:        parseTLSMinVersion(reader.GetString("radius", SettingEapTlsMinVersion)),
		}, nil
	}
}

// NewSettingsPEAPConfigProvider returns a TLSConfigProvider for PEAP's outer
// server-authenticated tunnel.
//
// PEAPv0 ([MS-PEAP]) reuses the EAP-TLS fragmentation/framing defined by
// RFC 5216 §2.1.5 and §3.1 but authenticates the peer with an inner EAP method,
// so no client CA is required for the outer TLS handshake. PEAP intentionally
// reuses the existing EAP-TLS server certificate settings
// (radius.EapTlsCertFile / radius.EapTlsKeyFile) and honors the configured
// minimum TLS version (radius.EapTlsMinVersion).
//
// Security note: PEAP is a compatibility-oriented method. The inner
// EAP-MSCHAPv2 exchange carries an NTLMv1-like attack surface (per Microsoft),
// so the outer TLS tunnel must stay strong — this provider keeps ServerOnly
// authentication with the operator-selected minimum TLS version and never
// weakens it. Deployments whose clients support certificates should prefer
// EAP-TLS.
func NewSettingsPEAPConfigProvider(reader TLSSettingsReader) TLSConfigProvider {
	return func() (*tlsengine.Config, error) {
		if reader == nil {
			return nil, nil
		}

		certFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsCertFile))
		keyFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsKeyFile))
		if certFile == "" || keyFile == "" {
			return nil, nil
		}

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("load PEAP server certificate: %w", err)
		}

		return &tlsengine.Config{
			ServerCertificate: cert,
			ServerOnly:        true,
			MinVersion:        parseTLSMinVersion(reader.GetString("radius", SettingEapTlsMinVersion)),
		}, nil
	}
}

// NewSettingsTTLSConfigProvider returns a TLSConfigProvider for EAP-TTLS's outer
// server-authenticated tunnel.
//
// EAP-TTLS (RFC 5281 §7) reuses the EAP-TLS fragmentation/framing defined by
// RFC 5216 §2.1.5 and §3.1 but authenticates the peer with legacy inner
// authentication carried as AVPs inside the tunnel (PAP / CHAP / MS-CHAP /
// MS-CHAP-V2), so no client CA is required for the outer TLS handshake. Like
// PEAP it intentionally reuses the existing EAP-TLS server certificate settings
// (radius.EapTlsCertFile / radius.EapTlsKeyFile) and honors the configured
// minimum TLS version (radius.EapTlsMinVersion).
//
// Security note: the outer tunnel carries cleartext-equivalent inner
// credentials (PAP sends the password inside the tunnel), so the server-only
// TLS tunnel must stay strong — this provider keeps ServerOnly authentication
// with the operator-selected minimum TLS version and never weakens it. Until
// the server certificate/key are configured it returns a nil config so the
// handler rejects safely with eap.ErrTLSNotConfigured.
func NewSettingsTTLSConfigProvider(reader TLSSettingsReader) TLSConfigProvider {
	return func() (*tlsengine.Config, error) {
		if reader == nil {
			return nil, nil
		}

		certFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsCertFile))
		keyFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsKeyFile))
		if certFile == "" || keyFile == "" {
			return nil, nil
		}

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("load EAP-TTLS server certificate: %w", err)
		}

		return &tlsengine.Config{
			ServerCertificate: cert,
			ServerOnly:        true,
			MinVersion:        parseTLSMinVersion(reader.GetString("radius", SettingEapTlsMinVersion)),
			// Pin the outer tunnel to TLS 1.2. EAP-TTLS phase 2 is peer-initiated
			// and ToughRADIUS relies on the TLS 1.2 handshake-completion framing
			// (the server's final flight) to switch into the inner AVP exchange.
			// TLS 1.3 tunneling (half-RTT completion, RFC 9427 key derivation) is
			// a later milestone, so cap the negotiation at TLS 1.2 here.
			MaxVersion: tls.VersionTLS12,
		}, nil
	}
}

// parseTLSMinVersion maps a configured minimum TLS version string to the
// crypto/tls constant. It defaults to TLS 1.2, the interoperability floor for
// modern EAP-TLS deployments; an unrecognized value falls back to the same
// safe default rather than weakening the handshake.
func parseTLSMinVersion(v string) uint16 {
	switch strings.TrimSpace(v) {
	case "1.3":
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}
