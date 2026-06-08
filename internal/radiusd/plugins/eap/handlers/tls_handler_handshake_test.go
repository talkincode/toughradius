package handlers

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsengine"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsfragment"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// --- test CA / certificate helpers ----------------------------------------

type hsTestCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	pool *x509.CertPool
}

func newHSTestCA(t *testing.T, cn string) *hsTestCA {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	return &hsTestCA{cert: cert, key: key, pool: pool}
}

func (ca *hsTestCA) issue(t *testing.T, cn string, customize func(*x509.Certificate)) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	if customize != nil {
		customize(tmpl)
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, &key.PublicKey, ca.key)
	require.NoError(t, err)
	leaf, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}
}

func serverEngineConfig(t *testing.T, serverCA, clientCA *hsTestCA) *tlsengine.Config {
	t.Helper()
	serverCert := serverCA.issue(t, "radius.example.com", func(c *x509.Certificate) {
		c.DNSNames = []string{"radius.example.com"}
	})
	return &tlsengine.Config{
		ServerCertificate: serverCert,
		ClientCAs:         clientCA.pool,
		MinVersion:        tls.VersionTLS12,
		HandshakeTimeout:  5 * time.Second,
	}
}

// --- in-memory duplex conn for the test TLS client ------------------------

type hsStream struct {
	mu      sync.Mutex
	cond    *sync.Cond
	buf     []byte
	closed  bool
	reading bool
}

func newHSStream() *hsStream {
	s := &hsStream{}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *hsStream) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for len(s.buf) == 0 && !s.closed {
		s.reading = true
		s.cond.Broadcast()
		s.cond.Wait()
	}
	s.reading = false
	if len(s.buf) == 0 && s.closed {
		return 0, net.ErrClosed
	}
	n := copy(p, s.buf)
	s.buf = s.buf[n:]
	return n, nil
}

func (s *hsStream) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, net.ErrClosed
	}
	s.buf = append(s.buf, p...)
	s.cond.Broadcast()
	return len(p), nil
}

func (s *hsStream) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.cond.Broadcast()
}

// drain returns and clears all buffered bytes.
func (s *hsStream) drain() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.buf) == 0 {
		return nil
	}
	out := s.buf
	s.buf = nil
	return out
}

type hsAddr struct{}

func (hsAddr) Network() string { return "eap-tls-test" }
func (hsAddr) String() string  { return "eap-tls-test" }

type hsConn struct {
	rd *hsStream
	wr *hsStream
}

func (c *hsConn) Read(p []byte) (int, error)       { return c.rd.Read(p) }
func (c *hsConn) Write(p []byte) (int, error)      { return c.wr.Write(p) }
func (c *hsConn) Close() error                     { c.rd.close(); c.wr.close(); return nil }
func (c *hsConn) LocalAddr() net.Addr              { return hsAddr{} }
func (c *hsConn) RemoteAddr() net.Addr             { return hsAddr{} }
func (c *hsConn) SetDeadline(time.Time) error      { return nil }
func (c *hsConn) SetReadDeadline(time.Time) error  { return nil }
func (c *hsConn) SetWriteDeadline(time.Time) error { return nil }

// supplicant emulates the EAP-TLS peer (NAS-side supplicant): it bridges between
// EAP-TLS framing and a real crypto/tls client, reassembling inbound server
// fragments and acknowledging each one per RFC 5216 §2.1.5.
type supplicant struct {
	t          *testing.T
	h          *TLSHandler
	sm         eap.EAPStateManager
	stateID    string
	secret     string
	toClient   *hsStream // server -> client TLS bytes
	fromClient *hsStream // client -> server TLS bytes
	clientDone chan error
	ident      uint8
	finished   bool
}

// newSupplicant starts a TLS client handshake bound to the handler's state.
func newSupplicant(t *testing.T, h *TLSHandler, sm eap.EAPStateManager, stateID, secret string, clientCfg *tls.Config) *supplicant {
	t.Helper()
	toClient := newHSStream()
	fromClient := newHSStream()
	conn := &hsConn{rd: toClient, wr: fromClient}
	client := tls.Client(conn, clientCfg)

	done := make(chan error, 1)
	go func() {
		err := client.Handshake()
		_ = client.Close()
		done <- err
	}()

	return &supplicant{
		t: t, h: h, sm: sm, stateID: stateID, secret: secret,
		toClient: toClient, fromClient: fromClient, clientDone: done, ident: 20,
	}
}

// waitSettled blocks until the TLS client is blocked waiting for more server
// bytes (its current flight is fully written) or until it has finished.
func (s *supplicant) waitSettled() {
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-s.clientDone:
			s.finished = true
			return
		case <-deadline:
			s.t.Fatal("timed out waiting for TLS client to settle")
		default:
		}
		s.toClient.mu.Lock()
		settled := len(s.toClient.buf) == 0 && s.toClient.reading
		s.toClient.mu.Unlock()
		if settled {
			return
		}
		select {
		case err := <-s.clientDone:
			s.finished = true
			_ = err
			return
		case <-time.After(2 * time.Millisecond):
		}
	}
}

// nextClientFlight returns the next batch of TLS bytes the client produced.
func (s *supplicant) nextClientFlight() []byte {
	s.waitSettled()
	return s.fromClient.drain()
}

// run drives the full EAP-TLS exchange until the handler authenticates, rejects,
// or the round budget is exhausted.
func (s *supplicant) run() (success bool, err error) {
	// The opening client flight (ClientHello).
	respData := s.nextClientFlight()
	require.NotEmpty(s.t, respData, "client should produce a ClientHello")

	var serverBuf []byte
	for round := 0; round < 64; round++ {
		writer := &mockResponseWriter{}
		ctx := s.responseCtx(writer, respData)
		ok, herr := s.h.HandleResponse(ctx)
		if herr != nil {
			return false, herr
		}
		if ok {
			return true, nil
		}
		require.NotNil(s.t, writer.response, "handler must answer with a challenge")
		require.Equal(s.t, radius.CodeAccessChallenge, writer.response.Code)

		frag := s.parseChallenge(writer.response)
		serverBuf = append(serverBuf, frag.Data...)
		if frag.More() {
			// Acknowledge this fragment and let the server send the next one.
			respData = nil // empty EAP-TLS response = ACK
			continue
		}

		// A full server flight has been reassembled.
		flight := serverBuf
		serverBuf = nil
		if !s.finished && len(flight) > 0 {
			_, werr := s.toClient.Write(flight)
			require.NoError(s.t, werr)
		}
		next := s.nextClientFlight()
		respData = next // may be empty (pure ACK) when the client has finished
	}
	s.t.Fatal("EAP-TLS exchange did not complete within the round budget")
	return false, nil
}

func (s *supplicant) responseCtx(writer *mockResponseWriter, tlsData []byte) *eap.EAPContext {
	packet := radius.New(radius.CodeAccessRequest, []byte(s.secret))
	require.NoError(s.t, rfc2865.State_SetString(packet, s.stateID))
	s.ident++
	var data []byte
	if len(tlsData) == 0 {
		data = (&tlsfragment.Packet{}).Encode() // ACK: single zero flags octet
	} else {
		data = (&tlsfragment.Packet{HasLength: true, TLSMessageLength: uint32(len(tlsData)), Data: tlsData}).Encode() //nolint:gosec // G115: test TLS payloads are far below uint32 max
	}
	return &eap.EAPContext{
		Request:        &radius.Request{Packet: packet},
		ResponseWriter: writer,
		StateManager:   s.sm,
		Secret:         s.secret,
		EAPMessage:     &eap.EAPMessage{Code: eap.CodeResponse, Identifier: s.ident, Type: eap.TypeTLS, Data: data},
	}
}

func (s *supplicant) parseChallenge(resp *radius.Packet) *tlsfragment.Packet {
	eapData, err := rfc2869.EAPMessage_Lookup(resp)
	require.NoError(s.t, err)
	require.GreaterOrEqual(s.t, len(eapData), 5)
	require.Equal(s.t, byte(eap.TypeTLS), eapData[4])
	frag, err := tlsfragment.Parse(eapData[5:])
	require.NoError(s.t, err)
	return frag
}

// startHandshake runs HandleIdentity and returns the issued state ID.
func startHandshake(t *testing.T, h *TLSHandler, sm eap.EAPStateManager, username, secret string) string {
	t.Helper()
	packet := radius.New(radius.CodeAccessRequest, []byte(secret))
	if username != "" {
		require.NoError(t, rfc2865.UserName_SetString(packet, username))
	}
	writer := &mockResponseWriter{}
	ctx := &eap.EAPContext{
		Request:        &radius.Request{Packet: packet},
		ResponseWriter: writer,
		StateManager:   sm,
		Secret:         secret,
		EAPMessage:     &eap.EAPMessage{Code: eap.CodeResponse, Identifier: 1, Type: eap.TypeIdentity},
	}
	handled, err := h.HandleIdentity(ctx)
	require.NoError(t, err)
	require.True(t, handled)
	return rfc2865.State_GetString(writer.response)
}

func clientCfg(serverCA *hsTestCA, clientCert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      serverCA.pool,
		ServerName:   "radius.example.com",
		MinVersion:   tls.VersionTLS12,
	}
}

// --- end-to-end handshake tests -------------------------------------------

func TestTLSHandler_FullHandshake_Succeeds(t *testing.T) {
	ca := newHSTestCA(t, "Test Root CA")
	clientCert := ca.issue(t, "alice", func(c *x509.Certificate) {
		c.EmailAddresses = []string{"alice@example.com"}
	})

	cfg := serverEngineConfig(t, ca, ca)
	h := NewTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	// EAP identity matches the certificate SAN email.
	stateID := startHandshake(t, h, sm, "alice@example.com", "secret")

	sup := newSupplicant(t, h, sm, stateID, "secret", clientCfg(ca, clientCert))
	success, err := sup.run()
	require.NoError(t, err)
	assert.True(t, success, "EAP-TLS handshake with a trusted client cert must authenticate")

	state, err := sm.GetState(stateID)
	require.NoError(t, err)
	assert.True(t, state.Success)
}

func TestTLSHandler_FullHandshake_UntrustedCARejected(t *testing.T) {
	serverCA := newHSTestCA(t, "Server Root CA")
	rogueCA := newHSTestCA(t, "Rogue CA")
	clientCert := rogueCA.issue(t, "mallory", func(c *x509.Certificate) {
		c.EmailAddresses = []string{"mallory@example.com"}
	})

	cfg := serverEngineConfig(t, serverCA, serverCA) // trusts only serverCA
	h := NewTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "mallory@example.com", "secret")
	sup := newSupplicant(t, h, sm, stateID, "secret", clientCfg(serverCA, clientCert))
	success, err := sup.run()
	assert.False(t, success)
	require.Error(t, err)
	assert.ErrorIs(t, err, eap.ErrTLSHandshakeFailed)
}

func TestTLSHandler_FullHandshake_IdentityMismatchRejected(t *testing.T) {
	ca := newHSTestCA(t, "Test Root CA")
	clientCert := ca.issue(t, "alice", func(c *x509.Certificate) {
		c.EmailAddresses = []string{"alice@example.com"}
	})

	cfg := serverEngineConfig(t, ca, ca)
	h := NewTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	// EAP identity (User-Name) does NOT match the certificate identity.
	stateID := startHandshake(t, h, sm, "bob@example.com", "secret")
	sup := newSupplicant(t, h, sm, stateID, "secret", clientCfg(ca, clientCert))
	success, err := sup.run()
	assert.False(t, success)
	require.Error(t, err)
	assert.ErrorIs(t, err, eap.ErrTLSIdentityMismatch)
}

func TestTLSHandler_FullHandshake_OutboundFragmentation(t *testing.T) {
	ca := newHSTestCA(t, "Test Root CA")
	clientCert := ca.issue(t, "carol", func(c *x509.Certificate) {
		c.EmailAddresses = []string{"carol@example.com"}
	})

	cfg := serverEngineConfig(t, ca, ca)
	h := NewTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })
	// Force the server's certificate flight to be split across many EAP-TLS
	// fragments, exercising the ACK-driven outbound fragmentation path.
	h.maxFragment = 64

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "carol@example.com", "secret")
	sup := newSupplicant(t, h, sm, stateID, "secret", clientCfg(ca, clientCert))
	success, err := sup.run()
	require.NoError(t, err)
	assert.True(t, success, "fragmented EAP-TLS handshake must still authenticate")
}
