package handlers

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
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
	// EAP peer (RFC 5216 §2.2). It is the sole source of the EAP-TLS/PEAP/TTLS
	// server certificate: operators select a managed certificate rather than
	// editing on-disk paths. When empty, certificate-based EAP is disabled.
	SettingEapTlsServerCert = "EapTlsServerCert"
	// SettingEapTlsClientCa is the local name of a managed CA certificate used
	// to verify the peer's client-certificate chain (RFC 5216 §2.2 / §5.3). It
	// is the sole source of the EAP-TLS client CA bundle. When empty, EAP-TLS is
	// disabled (peap/ttls tunnels are server-only and do not require it).
	SettingEapTlsClientCa = "EapTlsClientCa"
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
// satisfied by *app.CertStore and is the sole source of EAP-TLS/PEAP/TTLS
// certificate material (selected on the system config page), while keeping the
// handlers package free of an import dependency on internal/app or the domain
// models. A nil resolver disables certificate-based EAP: the providers return a
// nil config so the handler rejects safely with eap.ErrTLSNotConfigured.
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
// variadic argument so callers (and tests) that pass no resolver compile and
// operate with certificate-based EAP disabled.
func firstResolver(resolvers []CertResolver) CertResolver {
	for _, r := range resolvers {
		if r != nil {
			return r
		}
	}
	return nil
}

// serverCertConfigured reports whether a managed server certificate is selected,
// without performing any I/O. A certificate is selected when a resolver is
// available and SettingEapTlsServerCert names one. The presence check is kept
// separate from loading so the providers preserve their safe-reject semantics:
// EAP is treated as not configured until a certificate is selected, and only
// then is the material loaded (so load failures surface as errors rather than
// silent rejects).
func serverCertConfigured(reader TLSSettingsReader, resolver CertResolver) bool {
	return resolver != nil && strings.TrimSpace(reader.GetString("radius", SettingEapTlsServerCert)) != ""
}

// clientCAConfigured reports whether a managed client-CA certificate is
// selected, without performing any I/O (see serverCertConfigured).
func clientCAConfigured(reader TLSSettingsReader, resolver CertResolver) bool {
	return resolver != nil && strings.TrimSpace(reader.GetString("radius", SettingEapTlsClientCa)) != ""
}

// loadServerCertificate loads the EAP server certificate from the managed
// certificate named by SettingEapTlsServerCert. Callers must first confirm a
// certificate is selected via serverCertConfigured; any error returned here is a
// genuine load failure that should be surfaced.
func loadServerCertificate(reader TLSSettingsReader, resolver CertResolver, what string) (tls.Certificate, error) {
	name := strings.TrimSpace(reader.GetString("radius", SettingEapTlsServerCert))
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

// loadClientCAs loads the EAP-TLS client CA pool from the managed CA named by
// SettingEapTlsClientCa. Callers must first confirm a certificate is selected
// via clientCAConfigured.
func loadClientCAs(reader TLSSettingsReader, resolver CertResolver) (*x509.CertPool, error) {
	name := strings.TrimSpace(reader.GetString("radius", SettingEapTlsClientCa))
	caPEM, err := resolver.CABundle(name)
	if err != nil {
		return nil, fmt.Errorf("resolve EAP-TLS client CA %q: %w", name, err)
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
// peer against a trusted CA (RFC 5216 §2.2). Until a managed server certificate
// (selected via radius.EapTlsServerCert) and a managed client CA (selected via
// radius.EapTlsClientCa) are both selected, the provider returns a nil config so
// the handler rejects safely with eap.ErrTLSNotConfigured and can never
// authenticate a client without configured trust anchors. When the certificates
// are selected but the material fails to load, it returns an explicit error so
// the failure reason is surfaced rather than silently ignored.
//
// The resolver loads managed certificates by name; when it is nil (or no managed
// certificate is selected) certificate-based EAP stays disabled.
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
// certificate via radius.EapTlsServerCert) and honors the configured minimum TLS
// version (radius.EapTlsMinVersion).
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
// (a managed certificate via radius.EapTlsServerCert) and honors the configured
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
