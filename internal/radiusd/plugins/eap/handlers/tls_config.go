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
	// SettingEapTlsServerCert is the local name of a managed server certificate
	// (see internal/domain.SysCert / the Certificates page) presented to the
	// EAP peer. When set it takes precedence over SettingEapTlsCertFile /
	// SettingEapTlsKeyFile, letting operators select a certificate instead of
	// editing on-disk paths.
	SettingEapTlsServerCert = "EapTlsServerCert"
	// SettingEapTlsClientCa is the local name of a managed CA certificate used
	// to verify the peer's client-certificate chain. When set it takes
	// precedence over SettingEapTlsCaFile.
	SettingEapTlsClientCa = "EapTlsClientCa"
	// SettingEapTlsCertFile is the legacy PEM server-certificate path presented
	// to the EAP-TLS peer (RFC 5216 §2.2). Used as a fallback when no managed
	// server certificate is selected.
	SettingEapTlsCertFile = "EapTlsCertFile"
	// SettingEapTlsKeyFile is the legacy PEM private-key path matching the
	// server certificate. Used as a fallback when no managed server certificate
	// is selected.
	SettingEapTlsKeyFile = "EapTlsKeyFile"
	// SettingEapTlsCaFile is the legacy PEM CA bundle path used to verify the
	// peer's client-certificate chain (RFC 5216 §2.2 / §5.3). Used as a fallback
	// when no managed CA certificate is selected.
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

// CertResolver resolves managed certificate material by its local name. It is
// satisfied by *app.CertStore and lets the EAP providers prefer
// database-backed certificates (selected on the system config page) over on-disk
// file paths, while keeping the handlers package free of an import dependency on
// internal/app or the domain models. A nil resolver disables managed-certificate
// resolution and the providers fall back to the legacy file paths.
type CertResolver interface {
	// ServerKeyPair returns the PEM certificate chain and PEM private key for
	// the managed server certificate with the given name.
	ServerKeyPair(name string) (certPEM, keyPEM []byte, err error)
	// CABundle returns the PEM CA bundle for the managed CA certificate with the
	// given name.
	CABundle(name string) (caPEM []byte, err error)
}

// firstResolver returns the first non-nil resolver, or nil when none is
// provided. It lets the provider constructors accept the resolver as an optional
// variadic argument so existing callers (and tests) that pass only a reader keep
// compiling and operating on the legacy file paths.
func firstResolver(resolvers []CertResolver) CertResolver {
	for _, r := range resolvers {
		if r != nil {
			return r
		}
	}
	return nil
}

// serverCertConfigured reports whether a server certificate source is selected,
// without performing any I/O. A source is selected when a managed certificate is
// named (and a resolver is available to load it) or when both legacy file paths
// are set. The presence check is kept separate from loading so the providers
// preserve their safe-reject semantics: EAP is treated as not configured until
// every required source is selected, and only then is the material loaded (so
// load failures surface as errors rather than silent rejects).
func serverCertConfigured(reader TLSSettingsReader, resolver CertResolver) bool {
	if resolver != nil && strings.TrimSpace(reader.GetString("radius", SettingEapTlsServerCert)) != "" {
		return true
	}
	certFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsCertFile))
	keyFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsKeyFile))
	return certFile != "" && keyFile != ""
}

// clientCAConfigured reports whether a client-CA source is selected, without
// performing any I/O (see serverCertConfigured).
func clientCAConfigured(reader TLSSettingsReader, resolver CertResolver) bool {
	if resolver != nil && strings.TrimSpace(reader.GetString("radius", SettingEapTlsClientCa)) != "" {
		return true
	}
	return strings.TrimSpace(reader.GetString("radius", SettingEapTlsCaFile)) != ""
}

// loadServerCertificate loads the EAP server certificate from the selected
// source, preferring a managed certificate named by SettingEapTlsServerCert and
// falling back to the legacy SettingEapTlsCertFile / SettingEapTlsKeyFile paths.
// Callers must first confirm a source is selected via serverCertConfigured; any
// error returned here is a genuine load failure that should be surfaced.
func loadServerCertificate(reader TLSSettingsReader, resolver CertResolver, what string) (tls.Certificate, error) {
	if resolver != nil {
		if name := strings.TrimSpace(reader.GetString("radius", SettingEapTlsServerCert)); name != "" {
			certPEM, keyPEM, err := resolver.ServerKeyPair(name)
			if err != nil {
				return tls.Certificate{}, fmt.Errorf("resolve %s server certificate %q: %w", what, name, err)
			}
			cert, err := tls.X509KeyPair(certPEM, keyPEM)
			if err != nil {
				return tls.Certificate{}, fmt.Errorf("load %s server certificate %q: %w", what, name, err)
			}
			return cert, nil
		}
	}

	certFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsCertFile))
	keyFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsKeyFile))
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("load %s server certificate: %w", what, err)
	}
	return cert, nil
}

// loadClientCAs loads the EAP-TLS client CA pool from the selected source,
// preferring a managed CA named by SettingEapTlsClientCa and falling back to the
// legacy SettingEapTlsCaFile path. Callers must first confirm a source is
// selected via clientCAConfigured.
func loadClientCAs(reader TLSSettingsReader, resolver CertResolver) (*x509.CertPool, error) {
	var caPEM []byte
	if resolver != nil {
		if name := strings.TrimSpace(reader.GetString("radius", SettingEapTlsClientCa)); name != "" {
			pem, err := resolver.CABundle(name)
			if err != nil {
				return nil, fmt.Errorf("resolve EAP-TLS client CA %q: %w", name, err)
			}
			caPEM = pem
		}
	}
	if caPEM == nil {
		caFile := strings.TrimSpace(reader.GetString("radius", SettingEapTlsCaFile))
		pem, err := os.ReadFile(caFile) //nolint:gosec // G304: path is from validated config
		if err != nil {
			return nil, fmt.Errorf("read EAP-TLS client CA bundle: %w", err)
		}
		caPEM = pem
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("parse EAP-TLS client CA bundle: no PEM certificates found")
	}
	return pool, nil
}

// NewSettingsTLSConfigProvider returns a TLSConfigProvider that assembles the
// EAP-TLS material from dynamic settings on every handshake.
//
// EAP-TLS requires the server to present a certificate and to authenticate the
// peer against a trusted CA (RFC 5216 §2.2). Until a server certificate source
// (a managed certificate selected via radius.EapTlsServerCert, or the legacy
// radius.EapTlsCertFile / radius.EapTlsKeyFile paths) and a client-CA source (a
// managed certificate selected via radius.EapTlsClientCa, or the legacy
// radius.EapTlsCaFile path) are both selected, the provider returns a nil config
// so the handler rejects safely with eap.ErrTLSNotConfigured and can never
// authenticate a client without configured trust anchors. When the sources are
// selected but the material fails to load, it returns an explicit error so the
// failure reason is surfaced rather than silently ignored.
//
// The optional resolver loads managed certificates by name; when it is nil (or
// no managed certificate is selected) the provider uses the legacy file paths,
// preserving backward compatibility.
func NewSettingsTLSConfigProvider(reader TLSSettingsReader, resolvers ...CertResolver) TLSConfigProvider {
	resolver := firstResolver(resolvers)
	return func() (*tlsengine.Config, error) {
		if reader == nil {
			return nil, nil
		}

		// EAP-TLS is opt-in: treat it as not configured (safe reject) until both
		// the server certificate and the client CA sources are selected.
		if !serverCertConfigured(reader, resolver) || !clientCAConfigured(reader, resolver) {
			return nil, nil
		}

		cert, err := loadServerCertificate(reader, resolver, "EAP-TLS")
		if err != nil {
			return nil, err
		}
		pool, err := loadClientCAs(reader, resolver)
		if err != nil {
			return nil, err
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
// reuses the existing EAP-TLS server certificate selection (a managed
// certificate via radius.EapTlsServerCert, or the legacy radius.EapTlsCertFile /
// radius.EapTlsKeyFile paths) and honors the configured minimum TLS version
// (radius.EapTlsMinVersion).
//
// Security note: PEAP is a compatibility-oriented method. The inner
// EAP-MSCHAPv2 exchange carries an NTLMv1-like attack surface (per Microsoft),
// so the outer TLS tunnel must stay strong — this provider keeps ServerOnly
// authentication with the operator-selected minimum TLS version and never
// weakens it. Deployments whose clients support certificates should prefer
// EAP-TLS.
func NewSettingsPEAPConfigProvider(reader TLSSettingsReader, resolvers ...CertResolver) TLSConfigProvider {
	resolver := firstResolver(resolvers)
	return func() (*tlsengine.Config, error) {
		if reader == nil {
			return nil, nil
		}

		if !serverCertConfigured(reader, resolver) {
			return nil, nil
		}

		cert, err := loadServerCertificate(reader, resolver, "PEAP")
		if err != nil {
			return nil, err
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
// PEAP it intentionally reuses the existing EAP-TLS server certificate selection
// (a managed certificate via radius.EapTlsServerCert, or the legacy
// radius.EapTlsCertFile / radius.EapTlsKeyFile paths) and honors the configured
// minimum TLS version (radius.EapTlsMinVersion).
//
// Security note: the outer tunnel carries cleartext-equivalent inner
// credentials — inner PAP sends the password inside the tunnel, and inner
// MS-CHAP-V2 shares the NTLMv1-like attack surface noted by Microsoft — so the
// server-only TLS tunnel must stay strong: this provider keeps ServerOnly
// authentication with the operator-selected minimum TLS version and never
// weakens it. Until the server certificate is configured it returns a nil
// config so the handler rejects safely with eap.ErrTLSNotConfigured.
func NewSettingsTTLSConfigProvider(reader TLSSettingsReader, resolvers ...CertResolver) TLSConfigProvider {
	resolver := firstResolver(resolvers)
	return func() (*tlsengine.Config, error) {
		if reader == nil {
			return nil, nil
		}

		if !serverCertConfigured(reader, resolver) {
			return nil, nil
		}

		cert, err := loadServerCertificate(reader, resolver, "EAP-TTLS")
		if err != nil {
			return nil, err
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
