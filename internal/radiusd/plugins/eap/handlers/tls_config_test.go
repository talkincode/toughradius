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
	"os"
	"path/filepath"
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

// writeTestCertFiles generates a self-signed certificate and writes the
// certificate, its private key, and a CA bundle (the same self-signed cert) to
// PEM files in a temp dir, returning their paths.
func writeTestCertFiles(t *testing.T) (certFile, keyFile, caFile string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "toughradius-eap-tls-test"},
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

	dir := t.TempDir()
	certFile = filepath.Join(dir, "server.crt")
	keyFile = filepath.Join(dir, "server.key")
	caFile = filepath.Join(dir, "ca.crt")

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	mustWrite(t, certFile, certPEM)
	mustWrite(t, keyFile, keyPEM)
	mustWrite(t, caFile, certPEM)
	return certFile, keyFile, caFile
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
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
	cases := map[string]map[string]string{
		"all empty": {},
		"only cert": {
			"radius." + SettingEapTlsCertFile: "/tmp/server.crt",
		},
		"cert and key but no CA": {
			"radius." + SettingEapTlsCertFile: "/tmp/server.crt",
			"radius." + SettingEapTlsKeyFile:  "/tmp/server.key",
		},
		"whitespace-only values": {
			"radius." + SettingEapTlsCertFile: "   ",
			"radius." + SettingEapTlsKeyFile:  "\t",
			"radius." + SettingEapTlsCaFile:   " ",
		},
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			provider := NewSettingsTLSConfigProvider(newFakeReader(values))
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
	certFile, keyFile, caFile := writeTestCertFiles(t)
	reader := newFakeReader(map[string]string{
		"radius." + SettingEapTlsCertFile: certFile,
		"radius." + SettingEapTlsKeyFile:  keyFile,
		"radius." + SettingEapTlsCaFile:   caFile,
	})

	cfg, err := NewSettingsTLSConfigProvider(reader)()
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
	certFile, keyFile, caFile := writeTestCertFiles(t)
	base := map[string]string{
		"radius." + SettingEapTlsCertFile: certFile,
		"radius." + SettingEapTlsKeyFile:  keyFile,
		"radius." + SettingEapTlsCaFile:   caFile,
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

			cfg, err := NewSettingsTLSConfigProvider(newFakeReader(values))()
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

func TestNewSettingsTLSConfigProvider_LoadErrors(t *testing.T) {
	certFile, keyFile, caFile := writeTestCertFiles(t)

	t.Run("missing cert file", func(t *testing.T) {
		reader := newFakeReader(map[string]string{
			"radius." + SettingEapTlsCertFile: filepath.Join(t.TempDir(), "nope.crt"),
			"radius." + SettingEapTlsKeyFile:  keyFile,
			"radius." + SettingEapTlsCaFile:   caFile,
		})
		if _, err := NewSettingsTLSConfigProvider(reader)(); err == nil {
			t.Fatal("expected error for missing server certificate")
		}
	})

	t.Run("missing CA file", func(t *testing.T) {
		reader := newFakeReader(map[string]string{
			"radius." + SettingEapTlsCertFile: certFile,
			"radius." + SettingEapTlsKeyFile:  keyFile,
			"radius." + SettingEapTlsCaFile:   filepath.Join(t.TempDir(), "nope-ca.crt"),
		})
		if _, err := NewSettingsTLSConfigProvider(reader)(); err == nil {
			t.Fatal("expected error for missing CA bundle")
		}
	})

	t.Run("CA file without certificates", func(t *testing.T) {
		emptyCA := filepath.Join(t.TempDir(), "empty-ca.crt")
		mustWrite(t, emptyCA, []byte("not a pem certificate"))
		reader := newFakeReader(map[string]string{
			"radius." + SettingEapTlsCertFile: certFile,
			"radius." + SettingEapTlsKeyFile:  keyFile,
			"radius." + SettingEapTlsCaFile:   emptyCA,
		})
		_, err := NewSettingsTLSConfigProvider(reader)()
		if err == nil {
			t.Fatal("expected error for CA bundle with no certificates")
		}
		if !strings.Contains(err.Error(), "no PEM certificates") {
			t.Fatalf("expected 'no PEM certificates' error, got %v", err)
		}
	})
}

func TestNewSettingsPEAPConfigProvider_ValidMaterial(t *testing.T) {
	certFile, keyFile, _ := writeTestCertFiles(t)
	reader := newFakeReader(map[string]string{
		"radius." + SettingEapTlsCertFile:   certFile,
		"radius." + SettingEapTlsKeyFile:    keyFile,
		"radius." + SettingEapTlsMinVersion: "1.3",
	})

	cfg, err := NewSettingsPEAPConfigProvider(reader)()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config for PEAP server material")
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

func TestNewSettingsPEAPConfigProvider_NotConfigured(t *testing.T) {
	cases := map[string]map[string]string{
		"nil reader": nil,
		"all empty":  {},
		"only cert": {
			"radius." + SettingEapTlsCertFile: "server.crt",
		},
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			var reader TLSSettingsReader
			if values != nil {
				reader = newFakeReader(values)
			}
			cfg, err := NewSettingsPEAPConfigProvider(reader)()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg != nil {
				t.Fatalf("expected nil config when PEAP is not configured, got %#v", cfg)
			}
		})
	}
}
