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
