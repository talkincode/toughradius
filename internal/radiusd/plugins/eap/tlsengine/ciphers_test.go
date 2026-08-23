package tlsengine

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"testing"
	"time"
)

func TestResolveCipherSuites(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		profile string
		custom  string
		want    []uint16
		wantErr error
	}{
		{name: "empty is modern", profile: "", want: nil},
		{name: "modern", profile: "modern", want: nil},
		{name: "modern case", profile: "Modern", want: nil},
		{name: "legacy copies the CBC list", profile: CipherProfileLegacyRSACBC, want: LegacyRSACBCSuites},
		{
			name:    "custom by name",
			profile: CipherProfileCustom,
			custom:  "TLS_RSA_WITH_AES_128_CBC_SHA256, TLS_RSA_WITH_AES_256_CBC_SHA",
			want:    []uint16{tls.TLS_RSA_WITH_AES_128_CBC_SHA256, tls.TLS_RSA_WITH_AES_256_CBC_SHA},
		},
		{
			name:    "custom by hex id",
			profile: CipherProfileCustom,
			custom:  "0x003c",
			want:    []uint16{tls.TLS_RSA_WITH_AES_128_CBC_SHA256},
		},
		{name: "unknown profile", profile: "weak", wantErr: ErrUnknownCipherProfile},
		{name: "custom empty", profile: CipherProfileCustom, custom: "  ", wantErr: ErrEmptyCustomCipherSuites},
		{name: "custom rc4 rejected", profile: CipherProfileCustom, custom: "TLS_RSA_WITH_RC4_128_SHA", wantErr: ErrForbiddenCipherSuite},
		{name: "custom 3des hex rejected", profile: CipherProfileCustom, custom: "0x000a", wantErr: ErrForbiddenCipherSuite},
		{name: "custom unknown name", profile: CipherProfileCustom, custom: "TLS_DHE_RSA_WITH_AES_128_CBC_SHA", wantErr: ErrUnknownCipherSuite},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveCipherSuites(tc.profile, tc.custom)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("len=%d want %d (%v)", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("suites[%d]=%#x want %#x", i, got[i], tc.want[i])
				}
			}
			if tc.profile == CipherProfileLegacyRSACBC && len(got) > 0 {
				got[0] = 0
				if LegacyRSACBCSuites[0] == 0 {
					t.Fatal("ResolveCipherSuites must copy the legacy list")
				}
			}
		})
	}
}

func TestValidateCipherPolicy(t *testing.T) {
	t.Parallel()

	rsaCert := issueRSALeaf(t)
	ecdsaCert := issueECDSALeaf(t)

	if err := ValidateCipherPolicy(ecdsaCert, tls.VersionTLS12, nil); err != nil {
		t.Fatalf("modern + ECDSA should pass: %v", err)
	}
	if err := ValidateCipherPolicy(rsaCert, tls.VersionTLS12, LegacyRSACBCSuites); err != nil {
		t.Fatalf("legacy + RSA should pass: %v", err)
	}
	if err := ValidateCipherPolicy(ecdsaCert, tls.VersionTLS12, LegacyRSACBCSuites); !errors.Is(err, ErrLegacyNeedsRSACertificate) {
		t.Fatalf("legacy + ECDSA: got %v, want %v", err, ErrLegacyNeedsRSACertificate)
	}
	if err := ValidateCipherPolicy(rsaCert, tls.VersionTLS13, LegacyRSACBCSuites); !errors.Is(err, ErrLegacyIncompatibleTLS13) {
		t.Fatalf("legacy + TLS1.3: got %v, want %v", err, ErrLegacyIncompatibleTLS13)
	}
}

func TestEngine_LegacyRSAHandshakeWithCBCOnlyClient(t *testing.T) {
	ca, rsaCert := newRSATestCA(t)
	clientCfg := &tls.Config{ //nolint:gosec // G402: TLS 1.2 + CBC pin reproduces issue #598
		RootCAs:      ca.pool,
		ServerName:   "radius.example.com",
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS12,
		CipherSuites: LegacyRSACBCSuites,
	}

	t.Run("modern rejects CBC-only ClientHello", func(t *testing.T) {
		eng, err := New(&Config{
			ServerCertificate: rsaCert,
			ServerOnly:        true,
			MinVersion:        tls.VersionTLS12,
			HandshakeTimeout:  5 * time.Second,
		})
		if err != nil {
			t.Fatalf("new engine: %v", err)
		}
		defer func() { _ = eng.Close() }()

		clientErr, serverErr := driveHandshake(t, eng, clientCfg)
		if serverErr == nil {
			t.Fatal("modern profile must reject a CBC-only ClientHello")
		}
		if clientErr == nil {
			t.Fatal("CBC-only client must fail against the modern profile")
		}
	})

	t.Run("legacy-rsa-cbc accepts CBC-only ClientHello", func(t *testing.T) {
		eng, err := New(&Config{
			ServerCertificate: rsaCert,
			ServerOnly:        true,
			MinVersion:        tls.VersionTLS12,
			MaxVersion:        tls.VersionTLS12,
			CipherSuites:      LegacyRSACBCSuites,
			HandshakeTimeout:  5 * time.Second,
		})
		if err != nil {
			t.Fatalf("new engine: %v", err)
		}
		defer func() { _ = eng.Close() }()

		clientErr, serverErr := driveHandshake(t, eng, clientCfg)
		if clientErr != nil || serverErr != nil {
			t.Fatalf("legacy handshake: client=%v server=%v", clientErr, serverErr)
		}
		state := eng.conn.ConnectionState()
		if state.CipherSuite != tls.TLS_RSA_WITH_AES_128_CBC_SHA256 &&
			state.CipherSuite != tls.TLS_RSA_WITH_AES_256_CBC_SHA &&
			state.CipherSuite != tls.TLS_RSA_WITH_AES_128_CBC_SHA {
			t.Fatalf("negotiated unexpected suite %#x", state.CipherSuite)
		}
	})
}

func issueECDSALeaf(t *testing.T) tls.Certificate {
	t.Helper()
	ca := newTestCA(t, "ECDSA policy CA")
	return ca.issue(t, "radius.example.com", func(c *x509.Certificate) {
		c.DNSNames = []string{"radius.example.com"}
	})
}

func issueRSALeaf(t *testing.T) tls.Certificate {
	t.Helper()
	_, cert := newRSATestCA(t)
	return cert
}

func newRSATestCA(t *testing.T) (*testCA, tls.Certificate) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA CA key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "RSA policy CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create RSA CA: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse RSA CA: %v", err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	ca := &testCA{cert: cert, key: nil, pool: pool}

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA leaf: %v", err)
	}
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "radius.example.com"},
		DNSNames:     []string{"radius.example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, cert, &leafKey.PublicKey, key)
	if err != nil {
		t.Fatalf("create RSA leaf: %v", err)
	}
	leaf, err := x509.ParseCertificate(leafDER)
	if err != nil {
		t.Fatalf("parse RSA leaf: %v", err)
	}
	return ca, tls.Certificate{
		Certificate: [][]byte{leafDER},
		PrivateKey:  leafKey,
		Leaf:        leaf,
	}
}
