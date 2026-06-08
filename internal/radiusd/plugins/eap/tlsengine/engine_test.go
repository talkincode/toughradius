package tlsengine

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"sync"
	"testing"
	"time"
)

// --- in-memory test certificate authority ---------------------------------

type testCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	pool *x509.CertPool
}

func newTestCA(t *testing.T, commonName string) *testCA {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create CA cert: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse CA cert: %v", err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	return &testCA{cert: cert, key: key, pool: pool}
}

// issue creates a leaf certificate signed by the CA. The customize hook may set
// SAN fields (EmailAddresses / DNSNames) and key usage.
func (ca *testCA) issue(t *testing.T, commonName string, customize func(*x509.Certificate)) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate leaf key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	if customize != nil {
		customize(tmpl)
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		t.Fatalf("create leaf cert: %v", err)
	}
	return tls.Certificate{
		Certificate: [][]byte{der},
		PrivateKey:  key,
		Leaf:        mustParse(t, der),
	}
}

func mustParse(t *testing.T, der []byte) *x509.Certificate {
	t.Helper()
	c, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	return c
}

// --- buffered in-memory duplex pipe for the test client -------------------

type bufConn struct {
	rd *bufStream
	wr *bufStream
}

func (c *bufConn) Read(p []byte) (int, error)       { return c.rd.Read(p) }
func (c *bufConn) Write(p []byte) (int, error)      { return c.wr.Write(p) }
func (c *bufConn) Close() error                     { c.rd.close(); c.wr.close(); return nil }
func (c *bufConn) LocalAddr() net.Addr              { return memAddr{} }
func (c *bufConn) RemoteAddr() net.Addr             { return memAddr{} }
func (c *bufConn) SetDeadline(time.Time) error      { return nil }
func (c *bufConn) SetReadDeadline(time.Time) error  { return nil }
func (c *bufConn) SetWriteDeadline(time.Time) error { return nil }

type bufStream struct {
	mu     sync.Mutex
	cond   *sync.Cond
	buf    []byte
	closed bool
}

func newBufStream() *bufStream {
	s := &bufStream{}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *bufStream) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for len(s.buf) == 0 && !s.closed {
		s.cond.Wait()
	}
	if len(s.buf) == 0 && s.closed {
		return 0, net.ErrClosed
	}
	n := copy(p, s.buf)
	s.buf = s.buf[n:]
	return n, nil
}

func (s *bufStream) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, net.ErrClosed
	}
	s.buf = append(s.buf, p...)
	s.cond.Broadcast()
	return len(p), nil
}

func (s *bufStream) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.cond.Broadcast()
}

// newClientPipe returns a net.Conn for the test's TLS client and the two
// streams the driver loop uses to shuttle bytes to/from the server Engine.
func newClientPipe() (clientConn *bufConn, toClient *bufStream, fromClient *bufStream) {
	toClient = newBufStream()   // server -> client
	fromClient = newBufStream() // client -> server
	clientConn = &bufConn{rd: toClient, wr: fromClient}
	return
}

// driveHandshake runs a real tls.Client against the server Engine, shuttling
// flights between them until the server handshake finishes or errors.
func driveHandshake(t *testing.T, eng *Engine, clientCfg *tls.Config) (clientErr error, serverErr error) {
	t.Helper()
	clientConn, toClient, fromClient := newClientPipe()
	client := tls.Client(clientConn, clientCfg)

	clientDone := make(chan error, 1)
	go func() {
		err := client.Handshake()
		_ = client.Close()
		clientDone <- err
	}()

	buf := make([]byte, 16*1024)
	var firstInput []byte // opening Process call has no input
	in := firstInput
	deadline := time.After(5 * time.Second)
	for {
		out, done, err := eng.Process(in)
		if len(out) > 0 {
			if _, werr := toClient.Write(out); werr != nil {
				serverErr = werr
				break
			}
		}
		if done {
			break
		}
		if err != nil {
			serverErr = err
			break
		}
		// Read the next client flight.
		select {
		case <-deadline:
			t.Fatal("handshake timed out")
		default:
		}
		n, rerr := fromClient.Read(buf)
		if rerr != nil {
			serverErr = rerr
			break
		}
		in = append([]byte(nil), buf[:n]...)
	}

	select {
	case clientErr = <-clientDone:
	case <-time.After(2 * time.Second):
		t.Fatal("client handshake did not return")
	}
	return clientErr, serverErr
}

func serverConfig(t *testing.T, serverCA *testCA, clientCA *testCA) *Config {
	t.Helper()
	serverCert := serverCA.issue(t, "radius.example.com", func(c *x509.Certificate) {
		c.DNSNames = []string{"radius.example.com"}
	})
	return &Config{
		ServerCertificate: serverCert,
		ClientCAs:         clientCA.pool,
		MinVersion:        tls.VersionTLS12,
		HandshakeTimeout:  5 * time.Second,
	}
}

func clientTLSConfig(serverCA *testCA, clientCert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      serverCA.pool,
		ServerName:   "radius.example.com",
		MinVersion:   tls.VersionTLS12,
	}
}

// --- tests ----------------------------------------------------------------

func TestEngine_New_RequiresClientCAs(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("expected error for nil config")
	}
	if _, err := New(&Config{}); err == nil {
		t.Fatal("expected error for missing ClientCAs")
	}
}

func TestEngine_New_RequiresServerCertificate(t *testing.T) {
	ca := newTestCA(t, "Root")
	if _, err := New(&Config{ClientCAs: ca.pool}); err != ErrNoServerCertificate {
		t.Fatalf("expected ErrNoServerCertificate, got %v", err)
	}
}

func TestEngine_TrustedClientHandshakeSucceeds(t *testing.T) {
	ca := newTestCA(t, "Test Root CA")
	clientCert := ca.issue(t, "alice", func(c *x509.Certificate) {
		c.EmailAddresses = []string{"alice@example.com"}
	})

	eng, err := New(serverConfig(t, ca, ca))
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	defer func() { _ = eng.Close() }()

	clientErr, serverErr := driveHandshake(t, eng, clientTLSConfig(ca, clientCert))
	if clientErr != nil {
		t.Fatalf("client handshake error: %v", clientErr)
	}
	if serverErr != nil {
		t.Fatalf("server handshake error: %v", serverErr)
	}

	id, err := eng.Identity()
	if err != nil {
		t.Fatalf("identity: %v", err)
	}
	if id.Name != "alice@example.com" || id.Source != SourceSANEmail {
		t.Fatalf("unexpected identity: %+v", id)
	}
	if !id.Matches("alice@example.com") {
		t.Fatal("expected identity to match SAN email")
	}
	if id.Matches("alice") {
		t.Fatal("did not expect SAN identity to match subject CN")
	}
	if id.Matches("bob") {
		t.Fatal("did not expect match for unrelated name")
	}
}

func TestEngine_UntrustedClientCertRejected(t *testing.T) {
	serverCA := newTestCA(t, "Server Root CA")
	otherCA := newTestCA(t, "Rogue CA")
	// Client cert signed by a CA the server does not trust.
	clientCert := otherCA.issue(t, "mallory", func(c *x509.Certificate) {
		c.EmailAddresses = []string{"mallory@evil.example"}
	})

	eng, err := New(serverConfig(t, serverCA, serverCA))
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	defer func() { _ = eng.Close() }()

	// Note: with TLS 1.3 the client can locally complete its handshake before
	// the server verifies (and rejects) the client certificate, so only the
	// server-side outcome is asserted here.
	_, serverErr := driveHandshake(t, eng, clientTLSConfig(serverCA, clientCert))
	if serverErr == nil {
		t.Fatal("expected server to reject untrusted client certificate")
	}
	if _, err := eng.Identity(); err == nil {
		t.Fatal("expected Identity to fail after rejected handshake")
	}
}

func TestEngine_IdentityPrefersSANThenCN(t *testing.T) {
	ca := newTestCA(t, "Root")

	// DNS SAN only -> san-dns.
	dnsCert := ca.issue(t, "device-1", func(c *x509.Certificate) {
		c.DNSNames = []string{"device-1.example.com"}
	})
	engDNS, err := New(serverConfig(t, ca, ca))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = engDNS.Close() }()
	if _, serverErr := driveHandshake(t, engDNS, clientTLSConfig(ca, dnsCert)); serverErr != nil {
		t.Fatalf("dns handshake: %v", serverErr)
	}
	id, err := engDNS.Identity()
	if err != nil {
		t.Fatal(err)
	}
	if id.Name != "device-1.example.com" || id.Source != SourceSANDNS {
		t.Fatalf("expected san-dns identity, got %+v", id)
	}

	// CN only -> subject-cn.
	cnCert := ca.issue(t, "cn-only-user", nil)
	engCN, err := New(serverConfig(t, ca, ca))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = engCN.Close() }()
	if _, serverErr := driveHandshake(t, engCN, clientTLSConfig(ca, cnCert)); serverErr != nil {
		t.Fatalf("cn handshake: %v", serverErr)
	}
	id, err = engCN.Identity()
	if err != nil {
		t.Fatal(err)
	}
	if id.Name != "cn-only-user" || id.Source != SourceSubject {
		t.Fatalf("expected subject-cn identity, got %+v", id)
	}
	if !id.Matches("cn-only-user") {
		t.Fatal("expected CN-only identity to match subject CN")
	}
}

func TestEngine_IdentityBeforeHandshakeFails(t *testing.T) {
	ca := newTestCA(t, "Root")
	eng, err := New(serverConfig(t, ca, ca))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = eng.Close() }()
	if _, err := eng.Identity(); err == nil {
		t.Fatal("expected error before handshake completes")
	}
}
