//go:build integration

package integration

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"

	"github.com/talkincode/toughradius/v9/internal/domain"
	eap "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	eaphandlers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/handlers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsengine"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsfragment"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// TestRadiusEAPTLSHandshake drives a *complete* EAP-TLS handshake end-to-end
// against the live UDP RADIUS auth server (backed by PostgreSQL), proving the
// whole stack works for certificate authentication: UDP transport, RADIUS
// framing, the EAP coordinator, EAP-TLS Start/ACK fragmentation (RFC 5216
// §2.1.5), the server-side TLS engine, certificate-to-User-Name identity
// binding (RFC 5216 §5.2), and the database user lookup that turns a verified
// peer into an Access-Accept.
//
// Production registers the EAP-TLS handler without TLS material (certificate
// runtime wiring is a later milestone), so it safely rejects every handshake.
// To exercise the success path this test temporarily registers a
// certificate-configured handler into the shared registry and restores the
// original afterwards. The test is intentionally serial (no t.Parallel): the
// plugin registry and the configured EAP method are process-global state.
func TestRadiusEAPTLSHandshake(t *testing.T) {
	const secret = "it-eaptls-secret"
	suffix := uniqueSuffix()
	nasIP := "10.201.0.1"
	nasID := "it-eaptls-nas-" + suffix

	nas := &domain.NetNas{
		ID:         common.UUIDint64(),
		Identifier: nasID,
		Ipaddr:     nasIP,
		Secret:     secret,
		VendorCode: "0",
		Status:     common.ENABLED,
	}
	require.NoError(t, h.appCtx.DB().Create(nas).Error)

	profileID := seedProfile(t, "it-eaptls-profile-"+suffix)

	// The EAP identity (User-Name), the certificate identity, and the database
	// user must all agree for the handshake to end in Access-Accept.
	identity := "it-eaptls-" + suffix + "@example.com"
	user := &domain.RadiusUser{
		ID:         common.UUIDint64(),
		ProfileId:  profileID,
		Username:   identity,
		Password:   "unused-by-eap-tls",
		Status:     common.ENABLED,
		ExpireTime: time.Now().AddDate(1, 0, 0),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(user).Error)

	// Build an in-memory PKI: one CA that issues both the server certificate and
	// the (trusted) client certificate.
	ca := newEAPTLSCA(t, "IT EAP-TLS Root CA")
	serverCert := ca.issue(t, "radius.example.com", func(c *x509.Certificate) {
		c.DNSNames = []string{"radius.example.com"}
	})
	clientCert := ca.issue(t, "eap-tls-client", func(c *x509.Certificate) {
		c.EmailAddresses = []string{identity}
	})

	engineCfg := &tlsengine.Config{
		ServerCertificate: serverCert,
		ClientCAs:         ca.pool,
		MinVersion:        tls.VersionTLS12,
		HandshakeTimeout:  5 * time.Second,
	}

	// Swap in a certificate-configured EAP-TLS handler for the duration of this
	// test, then restore whatever was registered at suite startup.
	original, hadOriginal := registry.GetEAPHandler(eap.TypeTLS)
	registry.RegisterEAPHandler(eaphandlers.NewTLSHandlerWithConfig(
		func() (*tlsengine.Config, error) { return engineCfg, nil },
	))
	t.Cleanup(func() {
		if hadOriginal {
			registry.RegisterEAPHandler(original)
		} else {
			registry.RegisterEAPHandler(eaphandlers.NewTLSHandler())
		}
	})

	// Select EAP-TLS as the server's preferred method, then restore the prior
	// value so other tests are unaffected.
	cfgMgr := h.appCtx.ConfigMgr()
	prevMethod := cfgMgr.GetString("radius", "EapMethod")
	require.NoError(t, cfgMgr.Set("radius", "EapMethod", "eap-tls"))
	t.Cleanup(func() { _ = cfgMgr.Set("radius", "EapMethod", prevMethod) })

	serverAddr := fmt.Sprintf("127.0.0.1:%d", h.cfg.Radiusd.AuthPort)

	t.Run("accept trusted client certificate", func(t *testing.T) {
		sup := newEAPTLSSupplicant(t, serverAddr, secret, identity, nasID, nasIP,
			eapTLSClientConfig(ca, clientCert))
		defer sup.close()

		resp := sup.run(t)
		assert.Equalf(t, radius.CodeAccessAccept, resp.Code,
			"trusted client certificate must authenticate, got %v", resp.Code)
		assertEAPSuccess(t, resp)
		h.radiusSvc.ReleaseAuthRateLimit(identity)
	})

	t.Run("reject untrusted client certificate", func(t *testing.T) {
		// A client certificate from a CA the server does not trust must fail the
		// handshake and yield Access-Reject, never Access-Accept.
		rogueCA := newEAPTLSCA(t, "IT Rogue CA")
		rogueCert := rogueCA.issue(t, "eap-tls-client", func(c *x509.Certificate) {
			c.EmailAddresses = []string{identity}
		})

		sup := newEAPTLSSupplicant(t, serverAddr, secret, identity, nasID, nasIP,
			eapTLSClientConfig(ca, rogueCert))
		defer sup.close()

		resp := sup.run(t)
		assert.Equalf(t, radius.CodeAccessReject, resp.Code,
			"untrusted client certificate must be rejected, got %v", resp.Code)
		h.radiusSvc.ReleaseAuthRateLimit(identity)
	})
}

// assertEAPSuccess verifies the RADIUS reply carries an EAP-Success. EAP
// Success/Failure messages are only 4 bytes (Code, Identifier, Length) with no
// Type field, so they are read directly rather than via ParseEAPMessage (which
// requires a Type octet).
func assertEAPSuccess(t *testing.T, resp *radius.Packet) {
	t.Helper()
	eapData := rfc2869.EAPMessage_Get(resp)
	require.GreaterOrEqual(t, len(eapData), 1, "reply must carry an EAP-Message")
	assert.Equalf(t, byte(eap.CodeSuccess), eapData[0],
		"expected EAP-Success (code %d), got code %d", eap.CodeSuccess, eapData[0])
}

// --- in-memory test PKI ----------------------------------------------------

type eapTLSCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	pool *x509.CertPool
}

func newEAPTLSCA(t *testing.T, cn string) *eapTLSCA {
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
	return &eapTLSCA{cert: cert, key: key, pool: pool}
}

func (ca *eapTLSCA) issue(t *testing.T, cn string, customize func(*x509.Certificate)) tls.Certificate {
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

func eapTLSClientConfig(serverCA *eapTLSCA, clientCert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      serverCA.pool,
		ServerName:   "radius.example.com",
		MinVersion:   tls.VersionTLS12,
	}
}

// --- in-memory duplex pipe for the supplicant's TLS client -----------------

// eapStream is a blocking, buffered byte pipe connecting the supplicant driver
// to its crypto/tls client. The driver separates client flights using a quiet
// period on the client->server direction (see collectFlight) rather than any
// read-state heuristic, which is robust to the TLS client consuming a server
// flight in several Read calls.
type eapStream struct {
	mu     sync.Mutex
	cond   *sync.Cond
	buf    []byte
	closed bool
}

func newEAPStream() *eapStream {
	s := &eapStream{}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *eapStream) Read(p []byte) (int, error) {
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

func (s *eapStream) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, net.ErrClosed
	}
	s.buf = append(s.buf, p...)
	s.cond.Broadcast()
	return len(p), nil
}

func (s *eapStream) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.cond.Broadcast()
}

func (s *eapStream) drain() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.buf) == 0 {
		return nil
	}
	out := s.buf
	s.buf = nil
	return out
}

type eapAddr struct{}

func (eapAddr) Network() string { return "eap-tls-it" }
func (eapAddr) String() string  { return "eap-tls-it" }

type eapConn struct {
	rd *eapStream
	wr *eapStream
}

func (c *eapConn) Read(p []byte) (int, error)       { return c.rd.Read(p) }
func (c *eapConn) Write(p []byte) (int, error)      { return c.wr.Write(p) }
func (c *eapConn) Close() error                     { c.rd.close(); c.wr.close(); return nil }
func (c *eapConn) LocalAddr() net.Addr              { return eapAddr{} }
func (c *eapConn) RemoteAddr() net.Addr             { return eapAddr{} }
func (c *eapConn) SetDeadline(time.Time) error      { return nil }
func (c *eapConn) SetReadDeadline(time.Time) error  { return nil }
func (c *eapConn) SetWriteDeadline(time.Time) error { return nil }

// --- EAP-TLS supplicant over UDP -------------------------------------------

// eapTLSSupplicant emulates a NAS-side EAP-TLS peer: it bridges a real
// crypto/tls client to the RADIUS server by wrapping each TLS flight in EAP-TLS
// framing and exchanging it over UDP, reassembling inbound server fragments and
// acknowledging each one per RFC 5216 §2.1.5.
type eapTLSSupplicant struct {
	serverAddr string
	secret     string
	username   string
	nasID      string
	nasIP      string

	toClient   *eapStream // server -> client TLS bytes
	fromClient *eapStream // client -> server TLS bytes
	clientDone chan error
	finished   bool

	state []byte // RADIUS State attribute echoed across rounds
	ident uint8  // EAP identifier echoed back to the server
}

func newEAPTLSSupplicant(t *testing.T, serverAddr, secret, username, nasID, nasIP string, clientCfg *tls.Config) *eapTLSSupplicant {
	t.Helper()
	toClient := newEAPStream()
	fromClient := newEAPStream()
	client := tls.Client(&eapConn{rd: toClient, wr: fromClient}, clientCfg)

	done := make(chan error, 1)
	go func() {
		err := client.Handshake()
		_ = client.Close()
		done <- err
	}()

	return &eapTLSSupplicant{
		serverAddr: serverAddr,
		secret:     secret,
		username:   username,
		nasID:      nasID,
		nasIP:      nasIP,
		toClient:   toClient,
		fromClient: fromClient,
		clientDone: done,
	}
}

func (s *eapTLSSupplicant) close() {
	s.toClient.close()
	s.fromClient.close()
}

// run performs the identity round, then drives the EAP-TLS exchange until the
// server returns a terminal Access-Accept or Access-Reject. It fails the test
// (rather than hanging) on timeout via bounded round-trips and a round budget.
func (s *eapTLSSupplicant) run(t *testing.T) *radius.Packet {
	t.Helper()

	// Round 1: EAP-Response/Identity -> Access-Challenge carrying EAP-TLS Start.
	challenge := s.sendIdentity(t)
	require.Equalf(t, byte(eap.TypeTLS), challenge.Type,
		"expected EAP-TLS challenge, got EAP type %d", challenge.Type)
	start, err := tlsfragment.Parse(challenge.Data)
	require.NoError(t, err)
	require.Truef(t, start.Start(), "first EAP-TLS challenge must set the Start (S) flag, data=%x", challenge.Data)
	s.ident = challenge.Identifier

	// The opening client flight (ClientHello).
	respData := s.nextClientFlight(t)
	require.NotEmpty(t, respData, "client should produce a ClientHello")

	var serverBuf []byte
	for round := 0; round < 64; round++ {
		resp := s.sendTLS(t, respData)
		switch resp.Code {
		case radius.CodeAccessAccept, radius.CodeAccessReject:
			return resp
		case radius.CodeAccessChallenge:
			// fall through to process the next server fragment
		default:
			t.Fatalf("unexpected RADIUS reply code %v", resp.Code)
		}

		if st := rfc2865.State_Get(resp); len(st) > 0 {
			s.state = st
		}
		eapMsg, err := eap.ParseEAPMessage(resp)
		require.NoError(t, err)
		require.Equalf(t, byte(eap.TypeTLS), eapMsg.Type, "expected EAP-TLS, got type %d", eapMsg.Type)
		s.ident = eapMsg.Identifier

		frag, err := tlsfragment.Parse(eapMsg.Data)
		require.NoError(t, err)
		serverBuf = append(serverBuf, frag.Data...)
		if frag.More() {
			// Acknowledge this fragment with an empty EAP-TLS response so the
			// server sends the next one (RFC 5216 §2.1.5).
			respData = nil
			continue
		}

		// A full server flight has been reassembled; hand it to the TLS client.
		flight := serverBuf
		serverBuf = nil
		if !s.finished && len(flight) > 0 {
			_, werr := s.toClient.Write(flight)
			require.NoError(t, werr)
		}
		respData = s.nextClientFlight(t) // may be empty (pure ACK) once finished
	}

	t.Fatal("EAP-TLS exchange did not complete within the round budget")
	return nil
}

// nextClientFlight blocks until the TLS client has produced its next flight (or
// finished) and returns the bytes it wrote. A flight is a contiguous burst of
// writes followed by the client parking to wait for the server's reply; the
// burst is detected by a short quiet period with no new bytes, which (unlike a
// read-state heuristic) is immune to the TLS client consuming the previous
// server flight in multiple Read calls. The server round-trip latency is far
// larger than the quiet window, so flights never run together.
func (s *eapTLSSupplicant) nextClientFlight(t *testing.T) []byte {
	t.Helper()
	const quiet = 50 * time.Millisecond
	deadline := time.After(10 * time.Second)
	var collected []byte
	lastData := time.Now()
	for {
		select {
		case <-s.clientDone:
			s.finished = true
			return append(collected, s.fromClient.drain()...)
		default:
		}
		if chunk := s.fromClient.drain(); len(chunk) > 0 {
			collected = append(collected, chunk...)
			lastData = time.Now()
		} else if len(collected) > 0 && time.Since(lastData) >= quiet {
			return collected
		}
		select {
		case <-deadline:
			if len(collected) > 0 {
				return collected
			}
			t.Fatal("timed out waiting for TLS client flight")
		case <-time.After(2 * time.Millisecond):
		}
	}
}

// newAccessRequest builds a fresh Access-Request with the mandatory identity
// attributes shared by every round.
func (s *eapTLSSupplicant) newAccessRequest() *radius.Packet {
	packet := radius.New(radius.CodeAccessRequest, []byte(s.secret))
	_ = rfc2865.UserName_SetString(packet, s.username)
	_ = rfc2865.NASIdentifier_SetString(packet, s.nasID)
	_ = rfc2865.NASIPAddress_Set(packet, net.ParseIP(s.nasIP))
	return packet
}

// sendIdentity performs the EAP-Response/Identity round and returns the server's
// EAP-TLS Start challenge.
func (s *eapTLSSupplicant) sendIdentity(t *testing.T) *eap.EAPMessage {
	t.Helper()
	id := &eap.EAPMessage{
		Code:       eap.CodeResponse,
		Identifier: 1,
		Type:       eap.TypeIdentity,
		Data:       []byte(s.username),
	}
	packet := s.newAccessRequest()
	eap.SetEAPMessageAndAuth(packet, id.Encode(), s.secret)

	resp := s.exchange(t, packet)
	require.Equalf(t, radius.CodeAccessChallenge, resp.Code,
		"expected Access-Challenge after identity, got %v", resp.Code)
	s.state = rfc2865.State_Get(resp)
	require.NotEmpty(t, s.state, "challenge must carry a State attribute")

	challenge, err := eap.ParseEAPMessage(resp)
	require.NoError(t, err)
	return challenge
}

// sendTLS wraps a TLS flight (or an empty ACK) in EAP-TLS framing and exchanges
// it with the server.
func (s *eapTLSSupplicant) sendTLS(t *testing.T, tlsData []byte) *radius.Packet {
	t.Helper()
	// The server's EAP-TLS engine drives a long-lived crypto/tls handshake
	// goroutine that persists across RADIUS round-trips. Real supplicants have
	// network RTT between flights, which lets that goroutine settle into its
	// blocked Read before the next flight arrives. This loopback test has no
	// such latency, so a small pause here keeps the server-side turn detection
	// deterministic without touching production code.
	time.Sleep(15 * time.Millisecond)
	var frame []byte
	if len(tlsData) == 0 {
		frame = (&tlsfragment.Packet{}).Encode() // ACK: single zero flags octet
	} else {
		frame = (&tlsfragment.Packet{
			HasLength:        true,
			TLSMessageLength: uint32(len(tlsData)), //nolint:gosec // G115: test flights are far below uint32 max
			Data:             tlsData,
		}).Encode()
	}
	eapMsg := &eap.EAPMessage{
		Code:       eap.CodeResponse,
		Identifier: s.ident,
		Type:       eap.TypeTLS,
		Data:       frame,
	}
	packet := s.newAccessRequest()
	_ = rfc2865.State_Set(packet, s.state)
	eap.SetEAPMessageAndAuth(packet, eapMsg.Encode(), s.secret)
	return s.exchange(t, packet)
}

// exchange performs one bounded RADIUS round-trip so a stuck server fails the
// test fast instead of hanging.
func (s *eapTLSSupplicant) exchange(t *testing.T, packet *radius.Packet) *radius.Packet {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := radius.Exchange(ctx, packet, s.serverAddr)
	require.NoError(t, err)
	return resp
}
