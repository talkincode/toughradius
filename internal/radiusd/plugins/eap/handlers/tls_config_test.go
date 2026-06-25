package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

// fakeSettingsReader is an in-memory TLSSettingsReader for testing the
// settings-driven EAP-TLS config provider.
type fakeSettingsReader struct {
	values map[string]string
}

func (f *fakeSettingsReader) GetString(category, name string) string {
	return f.values[category+"."+name]
}

func newFakeReader(values map[string]string) *fakeSettingsReader {
	return &fakeSettingsReader{values: values}
}

// fakeCertResolver is an in-memory CertResolver for testing managed-certificate
// resolution. It is the sole source of EAP-TLS/PEAP/TTLS material.
type fakeCertResolver struct {
	certPEM []byte
	keyPEM  []byte
	caPEM   []byte
	err     error
}

func (f *fakeCertResolver) ServerKeyPair(string) ([]byte, []byte, error) {
	if f.err != nil {
		return nil, nil, f.err
	}
	return f.certPEM, f.keyPEM, nil
}

func (f *fakeCertResolver) CABundle(string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.caPEM, nil
}

// genTestCertPEM generates a self-signed certificate and returns the certificate
// and private key as in-memory PEM bytes.
func genTestCertPEM(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "toughradius-managed-cert"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM
}

func TestNewSettingsTLSConfigProvider_NilReader(t *testing.T) {
	provider := NewSettingsTLSConfigProvider(nil)
	cfg, err := provider()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Fatalf("expected nil config for nil reader, got %#v", cfg)
	}
}

func TestNewSettingsTLSConfigProvider_NotConfigured(t *testing.T) {
	certPEM, keyPEM := genTestCertPEM(t)
	resolver := &fakeCertResolver{certPEM: certPEM, keyPEM: keyPEM, caPEM: certPEM}

	cases := []struct {
		name     string
		values   map[string]string
		resolver CertResolver
	}{
		{
			name:     "no resolver even with refs",
			values:   map[string]string{"radius." + SettingEapTlsServerCert: "srv", "radius." + SettingEapTlsClientCa: "ca"},
			resolver: nil,
		},
		{
			name:     "resolver but no refs",
			values:   map[string]string{},
			resolver: resolver,
		},
		{
			name:     "resolver with server ref but no client CA ref",
			values:   map[string]string{"radius." + SettingEapTlsServerCert: "srv"},
			resolver: resolver,
		},
		{
			name:     "whitespace-only refs",
			values:   map[string]string{"radius." + SettingEapTlsServerCert: "  ", "radius." + SettingEapTlsClientCa: "\t"},
			resolver: resolver,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewSettingsTLSConfigProvider(newFakeReader(tc.values), tc.resolver)
			cfg, err := provider()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg != nil {
				t.Fatalf("expected nil config when not fully configured, got %#v", cfg)
			}
		})
	}
}

func TestNewSettingsTLSConfigProvider_ValidMaterial(t *testing.T) {
	certPEM, keyPEM := genTestCertPEM(t)
	resolver := &fakeCertResolver{certPEM: certPEM, keyPEM: keyPEM, caPEM: certPEM}
	reader := newFakeReader(map[string]string{
		"radius." + SettingEapTlsServerCert: "srv",
		"radius." + SettingEapTlsClientCa:   "ca",
	})

	cfg, err := NewSettingsTLSConfigProvider(reader, resolver)()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config for valid material")
	}
	if len(cfg.ServerCertificate.Certificate) == 0 || cfg.ServerCertificate.PrivateKey == nil {
		t.Fatal("expected a usable server certificate")
	}
	if cfg.ClientCAs == nil {
		t.Fatal("expected a non-nil client CA pool")
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("expected default MinVersion TLS 1.2, got %#x", cfg.MinVersion)
	}
}

func TestNewSettingsTLSConfigProvider_MinVersion(t *testing.T) {
	certPEM, keyPEM := genTestCertPEM(t)
	resolver := &fakeCertResolver{certPEM: certPEM, keyPEM: keyPEM, caPEM: certPEM}
	base := map[string]string{
		"radius." + SettingEapTlsServerCert: "srv",
		"radius." + SettingEapTlsClientCa:   "ca",
	}

	tests := []struct {
		configured string
		want       uint16
	}{
		{"1.2", tls.VersionTLS12},
		{"1.3", tls.VersionTLS13},
		{"", tls.VersionTLS12},
		{"bogus", tls.VersionTLS12},
	}
	for _, tc := range tests {
		t.Run("min="+tc.configured, func(t *testing.T) {
			values := make(map[string]string, len(base)+1)
			for k, v := range base {
				values[k] = v
			}
			values["radius."+SettingEapTlsMinVersion] = tc.configured

			cfg, err := NewSettingsTLSConfigProvider(newFakeReader(values), resolver)()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("expected non-nil config")
			}
			if cfg.MinVersion != tc.want {
				t.Fatalf("MinVersion for %q = %#x, want %#x", tc.configured, cfg.MinVersion, tc.want)
			}
		})
	}
}

func TestNewSettingsTLSConfigProvider_ResolverErrorSurfaces(t *testing.T) {
	resolver := &fakeCertResolver{err: errResolver}
	reader := newFakeReader(map[string]string{
		"radius." + SettingEapTlsServerCert: "srv",
		"radius." + SettingEapTlsClientCa:   "ca",
	})
	if _, err := NewSettingsTLSConfigProvider(reader, resolver)(); err == nil {
		t.Fatal("expected resolver error to surface")
	}
}

func TestNewSettingsTLSConfigProvider_ClientCANoPEM(t *testing.T) {
	certPEM, keyPEM := genTestCertPEM(t)
	// The CA bundle returned by the resolver contains no PEM certificates, so the
	// provider must surface an explicit parse error rather than build an empty
	// trust pool.
	resolver := &fakeCertResolver{certPEM: certPEM, keyPEM: keyPEM, caPEM: []byte("not a pem certificate")}
	reader := newFakeReader(map[string]string{
		"radius." + SettingEapTlsServerCert: "srv",
		"radius." + SettingEapTlsClientCa:   "ca",
	})
	_, err := NewSettingsTLSConfigProvider(reader, resolver)()
	if err == nil {
		t.Fatal("expected error for CA bundle with no certificates")
	}
	if !strings.Contains(err.Error(), "no PEM certificates") {
		t.Fatalf("expected 'no PEM certificates' error, got %v", err)
	}
}

func TestNewSettingsPEAPConfigProvider_NotConfigured(t *testing.T) {
	certPEM, keyPEM := genTestCertPEM(t)
	resolver := &fakeCertResolver{certPEM: certPEM, keyPEM: keyPEM}

	cases := []struct {
		name     string
		reader   TLSSettingsReader
		resolver CertResolver
	}{
		{name: "nil reader", reader: nil, resolver: resolver},
		{name: "resolver but no ref", reader: newFakeReader(map[string]string{}), resolver: resolver},
		{name: "ref but no resolver", reader: newFakeReader(map[string]string{"radius." + SettingEapTlsServerCert: "srv"}), resolver: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := NewSettingsPEAPConfigProvider(tc.reader, tc.resolver)()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg != nil {
				t.Fatalf("expected nil config when PEAP is not configured, got %#v", cfg)
			}
		})
	}
}

func TestNewSettingsPEAPConfigProvider_ManagedCert(t *testing.T) {
	certPEM, keyPEM := genTestCertPEM(t)
	resolver := &fakeCertResolver{certPEM: certPEM, keyPEM: keyPEM}
	reader := newFakeReader(map[string]string{
		"radius." + SettingEapTlsServerCert: "srv",
		"radius." + SettingEapTlsMinVersion: "1.3",
	})
	cfg, err := NewSettingsPEAPConfigProvider(reader, resolver)()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil PEAP config from managed certificate")
	}
	if !cfg.ServerOnly {
		t.Fatal("expected PEAP config to use server-only TLS")
	}
	if cfg.ClientCAs != nil {
		t.Fatal("PEAP outer TLS must not require a client CA")
	}
	if cfg.MinVersion != tls.VersionTLS13 {
		t.Fatalf("expected MinVersion TLS 1.3, got %#x", cfg.MinVersion)
	}
}

func TestNewSettingsTTLSConfigProvider_NotConfigured(t *testing.T) {
	certPEM, keyPEM := genTestCertPEM(t)
	resolver := &fakeCertResolver{certPEM: certPEM, keyPEM: keyPEM}

	cases := []struct {
		name     string
		reader   TLSSettingsReader
		resolver CertResolver
	}{
		{name: "nil reader", reader: nil, resolver: resolver},
		{name: "resolver but no ref", reader: newFakeReader(map[string]string{}), resolver: resolver},
		{name: "ref but no resolver", reader: newFakeReader(map[string]string{"radius." + SettingEapTlsServerCert: "srv"}), resolver: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := NewSettingsTTLSConfigProvider(tc.reader, tc.resolver)()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg != nil {
				t.Fatalf("expected nil config when TTLS is not configured, got %#v", cfg)
			}
		})
	}
}

func TestNewSettingsTTLSConfigProvider_ManagedCert(t *testing.T) {
	certPEM, keyPEM := genTestCertPEM(t)
	resolver := &fakeCertResolver{certPEM: certPEM, keyPEM: keyPEM}
	reader := newFakeReader(map[string]string{
		"radius." + SettingEapTlsServerCert: "srv",
	})
	cfg, err := NewSettingsTTLSConfigProvider(reader, resolver)()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil TTLS config from managed certificate")
	}
	if !cfg.ServerOnly {
		t.Fatal("expected TTLS config to use server-only TLS")
	}
	if cfg.ClientCAs != nil {
		t.Fatal("TTLS outer TLS must not require a client CA")
	}
	// EAP-TTLS pins the outer tunnel to TLS 1.2 (see provider docs).
	if cfg.MaxVersion != tls.VersionTLS12 {
		t.Fatalf("expected MaxVersion TLS 1.2, got %#x", cfg.MaxVersion)
	}
}

var errResolver = errTest("resolver failure")

type errTest string

func (e errTest) Error() string { return string(e) }
