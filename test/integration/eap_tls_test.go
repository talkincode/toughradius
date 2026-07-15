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
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"strings"
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
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsfragment"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// TestEAPTLSEndToEnd drives a full EAP-TLS handshake over real RADIUS UDP packets
// against the running auth server, exercising the dynamic certificate
// configuration wired in milestone M1.5 (TR-F004 / TR-F014). It asserts the
// success path (a trusted client certificate whose identity matches the
// User-Name authenticates) and the failure paths required by the acceptance
// criteria (RFC 5216 §2.2/§5.3 trust validation and §5.2 identity binding):
//   - an untrusted client certificate is rejected with an explicit reason;
//   - a trusted certificate whose identity does not match the User-Name is
//     rejected with an explicit reason.
//
// Milestone M10.1 (TR-F004 / RFC 9190) adds the TLS 1.3 cases: a pinned TLS 1.3
// client must receive and decrypt the §2.1.1 protected success indication
// (application data 0x00) before EAP-Success, and a pinned TLS 1.2 client must
// still authenticate through the legacy RFC 5216 flow without one.
//
// It is intentionally serial (no t.Parallel) because the RADIUS plugin registry,
// dynamic settings, and rate limiter are process-global shared state.
func TestEAPTLSEndToEnd(t *testing.T) {
	const secret = "it-eaptls-secret"
	suffix := uniqueSuffix()
	nasIP := net.ParseIP("10.201.0.1")
	nasID := "it-eaptls-nas-" + suffix

	nas := &domain.NetNas{
		ID:         common.UUIDint64(),
		Identifier: nasID,
		Ipaddr:     nasIP.String(),
		Secret:     secret,
		VendorCode: "0",
		Status:     common.ENABLED,
	}
	require.NoError(t, h.appCtx.DB().Create(nas).Error)

	profileID := seedProfile(t, "it-eaptls-profile-"+suffix)
	serverAddr := h.radiusServerAddr()

	// Restore the EAP method after this test so later integration cases are not
	// affected by the runtime switch to eap-tls.
	restoreEapMethod(t)

	t.Run("trusted client certificate authenticates", func(t *testing.T) {
		ca := newEAPTLSTestCA(t, "IT EAP-TLS Root CA "+suffix)
		serverCert := ca.issueServer(t, "radius.example.com")
		username := "it-eaptls-alice-" + suffix + "@example.com"
		clientCert := ca.issueClient(t, "alice", username)
		seedEAPTLSUser(t, profileID, username)

		configureEAPTLS(t, serverCert, ca.certPEM())

		sup := &eapTLSSupplicant{
			serverAddr: serverAddr,
			secret:     secret,
			username:   username,
			nasID:      nasID,
			nasIP:      nasIP,
			clientCfg:  clientTLSConfig(ca, clientCert),
		}
		resp := sup.authenticate(t)
		assert.Equalf(t, radius.CodeAccessAccept, resp.Code,
			"trusted EAP-TLS client must authenticate, got %v (%q)", resp.Code, rfc2865.ReplyMessage_GetString(resp))
		assertEAPCode(t, resp, eap.CodeSuccess)
	})

	t.Run("untrusted client certificate is rejected", func(t *testing.T) {
		serverCA := newEAPTLSTestCA(t, "IT EAP-TLS Server CA "+suffix)
		rogueCA := newEAPTLSTestCA(t, "IT EAP-TLS Rogue CA "+suffix)
		serverCert := serverCA.issueServer(t, "radius.example.com")
		username := "it-eaptls-mallory-" + suffix + "@example.com"
		// The rogue certificate chains to a CA the server does not trust.
		rogueCert := rogueCA.issueClient(t, "mallory", username)
		seedEAPTLSUser(t, profileID, username)

		// The server trusts only serverCA, so the rogue client cert fails the
		// chain validation during the handshake (RFC 5216 §2.2/§5.3).
		configureEAPTLS(t, serverCert, serverCA.certPEM())

		sup := &eapTLSSupplicant{
			serverAddr: serverAddr,
			secret:     secret,
			username:   username,
			nasID:      nasID,
			nasIP:      nasIP,
			// The client still trusts the server certificate so the rejection is
			// driven by the server distrusting the client, not vice versa.
			clientCfg: clientTLSConfig(serverCA, rogueCert),
		}
		resp := sup.authenticate(t)
		assert.Equalf(t, radius.CodeAccessReject, resp.Code,
			"untrusted EAP-TLS client must be rejected, got %v", resp.Code)
		assertEAPCode(t, resp, eap.CodeFailure)
		assert.Containsf(t, strings.ToLower(rfc2865.ReplyMessage_GetString(resp)), "handshake failed",
			"reject must carry an explicit EAP-TLS failure reason")
	})

	t.Run("certificate identity mismatch is rejected", func(t *testing.T) {
		ca := newEAPTLSTestCA(t, "IT EAP-TLS Mismatch CA "+suffix)
		serverCert := ca.issueServer(t, "radius.example.com")
		// The certificate identity (SAN email) differs from the RADIUS User-Name.
		certIdentity := "it-eaptls-cert-" + suffix + "@example.com"
		username := "it-eaptls-named-" + suffix + "@example.com"
		clientCert := ca.issueClient(t, "named", certIdentity)
		seedEAPTLSUser(t, profileID, username)

		configureEAPTLS(t, serverCert, ca.certPEM())

		sup := &eapTLSSupplicant{
			serverAddr: serverAddr,
			secret:     secret,
			username:   username,
			nasID:      nasID,
			nasIP:      nasIP,
			clientCfg:  clientTLSConfig(ca, clientCert),
		}
		resp := sup.authenticate(t)
		assert.Equalf(t, radius.CodeAccessReject, resp.Code,
			"EAP-TLS identity mismatch must be rejected, got %v", resp.Code)
		assertEAPCode(t, resp, eap.CodeFailure)
		assert.Containsf(t, strings.ToLower(rfc2865.ReplyMessage_GetString(resp)), "identity mismatch",
			"reject must carry an explicit identity-mismatch reason")
	})

	t.Run("tls 1.3 sends protected success indication", func(t *testing.T) {
		// RFC 9190 §2.1.1: with TLS 1.3 the server commits to success by
		// sending one octet of 0x00 as protected application data after its
		// final handshake message; EAP-Success follows the peer's ACK. The
		// pinned TLS 1.3 client decrypts and verifies the indication itself.
		ca := newEAPTLSTestCA(t, "IT EAP-TLS 13 CA "+suffix)
		serverCert := ca.issueServer(t, "radius.example.com")
		username := "it-eaptls-t13-" + suffix + "@example.com"
		clientCert := ca.issueClient(t, "t13", username)
		seedEAPTLSUser(t, profileID, username)

		configureEAPTLS(t, serverCert, ca.certPEM())

		cfg := clientTLSConfig(ca, clientCert)
		cfg.MinVersion = tls.VersionTLS13
		cfg.MaxVersion = tls.VersionTLS13

		sup := &eapTLSSupplicant{
			serverAddr: serverAddr,
			secret:     secret,
			username:   username,
			nasID:      nasID,
			nasIP:      nasIP,
			clientCfg:  cfg,
			clientRun: func(client *tls.Conn) error {
				if err := client.Handshake(); err != nil {
					return err
				}
				if v := client.ConnectionState().Version; v != tls.VersionTLS13 {
					return fmt.Errorf("negotiated version %#x, want TLS 1.3", v)
				}
				buf := make([]byte, 1)
				if _, err := io.ReadFull(client, buf); err != nil {
					return fmt.Errorf("read protected success indication: %w", err)
				}
				if buf[0] != 0x00 {
					return fmt.Errorf("protected success indication byte %#x, want 0x00", buf[0])
				}
				return nil
			},
		}
		resp := sup.authenticate(t)
		assert.Equalf(t, radius.CodeAccessAccept, resp.Code,
			"TLS 1.3 EAP-TLS client must authenticate, got %v (%q)", resp.Code, rfc2865.ReplyMessage_GetString(resp))
		assertEAPCode(t, resp, eap.CodeSuccess)
		require.NoError(t, sup.clientErr,
			"client must decrypt the 0x00 protected success indication before EAP-Success")
	})

	t.Run("tls 1.2 pinned client falls back", func(t *testing.T) {
		// A peer that cannot speak TLS 1.3 negotiates 1.2 and follows the
		// legacy RFC 5216 flow with no protected success indication.
		ca := newEAPTLSTestCA(t, "IT EAP-TLS 12 CA "+suffix)
		serverCert := ca.issueServer(t, "radius.example.com")
		username := "it-eaptls-t12-" + suffix + "@example.com"
		clientCert := ca.issueClient(t, "t12", username)
		seedEAPTLSUser(t, profileID, username)

		configureEAPTLS(t, serverCert, ca.certPEM())

		cfg := clientTLSConfig(ca, clientCert)
		cfg.MaxVersion = tls.VersionTLS12

		sup := &eapTLSSupplicant{
			serverAddr: serverAddr,
			secret:     secret,
			username:   username,
			nasID:      nasID,
			nasIP:      nasIP,
			clientCfg:  cfg,
			clientRun: func(client *tls.Conn) error {
				if err := client.Handshake(); err != nil {
					return err
				}
				if v := client.ConnectionState().Version; v != tls.VersionTLS12 {
					return fmt.Errorf("negotiated version %#x, want TLS 1.2", v)
				}
				return nil
			},
		}
		resp := sup.authenticate(t)
		assert.Equalf(t, radius.CodeAccessAccept, resp.Code,
			"TLS 1.2 pinned EAP-TLS client must authenticate, got %v (%q)", resp.Code, rfc2865.ReplyMessage_GetString(resp))
		assertEAPCode(t, resp, eap.CodeSuccess)
		require.NoError(t, sup.clientErr)
	})
}

// radiusServerAddr returns the auth server's UDP address.
func (hr *harness) radiusServerAddr() string {
	return fmt.Sprintf("127.0.0.1:%d", hr.cfg.Radiusd.AuthPort)
}

// seedEAPTLSUser creates an enabled RADIUS user. EAP-TLS authenticates by
// certificate, so the stored password is unused, but the user row must exist and
// be valid because the auth pipeline loads it before dispatching EAP.
func seedEAPTLSUser(t *testing.T, profileID int64, username string) {
	t.Helper()
	user := &domain.RadiusUser{
		ID:         common.UUIDint64(),
		ProfileId:  profileID,
		Username:   username,
		Password:   "eap-tls-unused",
		Status:     common.ENABLED,
		ExpireTime: time.Now().AddDate(1, 0, 0),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(user).Error)
}

// configureEAPTLS stores the server certificate/key and client CA bundle as
// managed certificates (domain.SysCert) and points the dynamic EAP-TLS settings
// at them by name, switching the server's EAP method to eap-tls. This mirrors an
// operator selecting certificates on the Certificates page; the settings
// provider resolves the managed certificates on every handshake, so the live
// server picks up the change without a restart (milestone M1.5).
func configureEAPTLS(t *testing.T, serverCert tls.Certificate, caPEMs ...[]byte) {
	t.Helper()
	serverName := seedManagedServerCert(t, serverCert)
	caName := seedManagedCA(t, caPEMs...)

	cm := h.appCtx.ConfigMgr()
	require.NoError(t, cm.Set("radius", eaphandlers.SettingEapTlsServerCert, serverName))
	require.NoError(t, cm.Set("radius", eaphandlers.SettingEapTlsClientCa, caName))
	require.NoError(t, cm.Set("radius", "EapMethod", "eap-tls"))
}

// seedManagedServerCert stores the EAP server certificate/key as a managed
// SysCert row and returns its unique local name. The CertStore resolver loads it
// by name on every handshake, so EAP-TLS/PEAP/TTLS material comes solely from the
// managed certificate store (no on-disk file paths).
func seedManagedServerCert(t *testing.T, serverCert tls.Certificate) string {
	t.Helper()
	name := fmt.Sprintf("it-eap-server-%d", common.UUIDint64())
	rec := &domain.SysCert{
		ID:         common.UUIDint64(),
		Name:       name,
		CertType:   "server",
		Cert:       string(certificatePEM(t, serverCert)),
		PrivateKey: string(privateKeyPEM(t, serverCert.PrivateKey)),
		NotBefore:  time.Now().Add(-time.Hour),
		NotAfter:   time.Now().Add(24 * time.Hour),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(rec).Error)
	return name
}

// seedManagedCA stores a CA bundle as a managed SysCert row and returns its
// unique local name.
func seedManagedCA(t *testing.T, caPEMs ...[]byte) string {
	t.Helper()
	bundle := make([]byte, 0)
	for _, p := range caPEMs {
		bundle = append(bundle, p...)
	}
	name := fmt.Sprintf("it-eap-ca-%d", common.UUIDint64())
	rec := &domain.SysCert{
		ID:        common.UUIDint64(),
		Name:      name,
		CertType:  "ca",
		Cert:      string(bundle),
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(rec).Error)
	return name
}

// restoreEapMethod captures the current EAP method and restores it when the test
// finishes, keeping the runtime switch to eap-tls scoped to this test.
func restoreEapMethod(t *testing.T) {
	t.Helper()
	cm := h.appCtx.ConfigMgr()
	prev := cm.GetString("radius", "EapMethod")
	t.Cleanup(func() {
		if prev == "" {
			prev = "eap-md5"
		}
		_ = cm.Set("radius", "EapMethod", prev)
	})
}

// assertEAPCode verifies the RADIUS reply carries an EAP-Message whose code field
// (EAP-Success or EAP-Failure) matches the expectation. EAP-Success/Failure are
// 4-octet headers without a Type field, so the raw attribute is inspected
// directly rather than via ParseEAPMessage (which expects a typed request).
func assertEAPCode(t *testing.T, resp *radius.Packet, code uint8) {
	t.Helper()
	eapAttr := rfc2869.EAPMessage_Get(resp)
	require.GreaterOrEqualf(t, len(eapAttr), 1, "response missing EAP-Message attribute")
	assert.Equalf(t, code, eapAttr[0], "expected EAP code %d, got %d", code, eapAttr[0])
}

// --- test CA / certificate helpers ----------------------------------------

type eapTLSTestCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	der  []byte
}

func newEAPTLSTestCA(t *testing.T, cn string) *eapTLSTestCA {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
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
	return &eapTLSTestCA{cert: cert, key: key, der: der}
}

func (ca *eapTLSTestCA) issueServer(t *testing.T, dnsName string) tls.Certificate {
	t.Helper()
	return ca.issue(t, dnsName, func(c *x509.Certificate) {
		c.DNSNames = []string{dnsName}
		c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	})
}

func (ca *eapTLSTestCA) issueClient(t *testing.T, cn, email string) tls.Certificate {
	t.Helper()
	return ca.issue(t, cn, func(c *x509.Certificate) {
		c.EmailAddresses = []string{email}
		c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	})
}

func (ca *eapTLSTestCA) issue(t *testing.T, cn string, customize func(*x509.Certificate)) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
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

// certPEM returns the CA certificate as PEM so it can be written into the
// server's client-CA bundle.
func (ca *eapTLSTestCA) certPEM() []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.der})
}

// pool returns a cert pool trusting only this CA, used by the test TLS client to
// validate the server certificate.
func (ca *eapTLSTestCA) pool() *x509.CertPool {
	p := x509.NewCertPool()
	p.AddCert(ca.cert)
	return p
}

func clientTLSConfig(serverCA *eapTLSTestCA, clientCert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      serverCA.pool(),
		ServerName:   "radius.example.com",
		MinVersion:   tls.VersionTLS12,
	}
}

func certificatePEM(t *testing.T, cert tls.Certificate) []byte {
	t.Helper()
	var out []byte
	for _, der := range cert.Certificate {
		out = append(out, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})...)
	}
	return out
}

func privateKeyPEM(t *testing.T, key any) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

// --- over-the-wire EAP-TLS supplicant -------------------------------------

// eapTLSSupplicant drives a real crypto/tls client through the EAP-TLS framing of
// live RADIUS Access-Request packets. It reassembles fragmented server flights
// and acknowledges each one with an empty EAP-TLS response (RFC 5216 §2.1.5),
// bridging the in-memory TLS handshake to the UDP RADIUS transport.
type eapTLSSupplicant struct {
	serverAddr string
	secret     string
	username   string
	nasID      string
	nasIP      net.IP
	clientCfg  *tls.Config

	toClient   *eapStream // server -> client TLS bytes
	fromClient *eapStream // client -> server TLS bytes
	clientDone chan error
	finished   bool
	// clientErr is the TLS client goroutine's result, captured when it exits.
	clientErr error
	// clientRun overrides the TLS client behavior (default: handshake only).
	// The TLS 1.3 test uses it to decrypt and verify the RFC 9190 §2.1.1
	// protected success indication.
	clientRun func(client *tls.Conn) error
}

// authenticate runs the whole EAP-TLS exchange and returns the final RADIUS
// reply (Access-Accept or Access-Reject).
func (s *eapTLSSupplicant) authenticate(t *testing.T) *radius.Packet {
	t.Helper()

	// Round 1: EAP-Response/Identity -> Access-Challenge carrying EAP-TLS Start.
	challenge, state := s.sendIdentity(t)
	require.Equalf(t, uint8(eap.TypeTLS), challenge.Type, "expected EAP-TLS challenge, got EAP type %d", challenge.Type)
	require.GreaterOrEqual(t, len(challenge.Data), 1)
	require.NotZerof(t, challenge.Data[0]&tlsfragment.FlagStart, "expected EAP-TLS Start (S) flag, got flags %#x", challenge.Data[0])

	s.startClient()
	defer s.close()

	respIdent := challenge.Identifier
	flight := s.nextClientFlight(t) // ClientHello
	require.NotEmpty(t, flight, "TLS client must produce a ClientHello")

	var serverBuf []byte
	for round := 0; round < 64; round++ {
		resp := s.sendTLSFlight(t, state, respIdent, flight)
		switch resp.Code {
		case radius.CodeAccessAccept, radius.CodeAccessReject:
			return resp
		case radius.CodeAccessChallenge:
			// fall through to process the next server flight
		default:
			t.Fatalf("unexpected RADIUS code %v during EAP-TLS handshake", resp.Code)
		}

		chMsg, err := eap.ParseEAPMessage(resp)
		require.NoError(t, err)
		require.Equalf(t, uint8(eap.TypeTLS), chMsg.Type, "expected EAP-TLS request, got type %d", chMsg.Type)
		respIdent = chMsg.Identifier
		if st := rfc2865.State_Get(resp); len(st) > 0 {
			state = st
		}

		frag, err := tlsfragment.Parse(chMsg.Data)
		require.NoError(t, err)
		serverBuf = append(serverBuf, frag.Data...)
		if frag.More() {
			// Acknowledge this fragment so the server sends the next one.
			flight = nil
			continue
		}

		// A complete server flight is reassembled; hand it to the TLS client.
		full := serverBuf
		serverBuf = nil
		if !s.finished && len(full) > 0 {
			_, werr := s.toClient.Write(full)
			require.NoError(t, werr)
		}
		flight = s.nextClientFlight(t) // may be empty (pure ACK) once finished
	}
	t.Fatal("EAP-TLS handshake did not complete within the round budget")
	return nil
}

// sendIdentity performs the opening EAP-Response/Identity round.
func (s *eapTLSSupplicant) sendIdentity(t *testing.T) (*eap.EAPMessage, []byte) {
	t.Helper()
	identity := &eap.EAPMessage{Code: eap.CodeResponse, Identifier: 1, Type: eap.TypeIdentity, Data: []byte(s.username)}
	packet := s.newAccessRequest()
	eap.SetEAPMessageAndAuth(packet, identity.Encode(), s.secret)

	resp, err := s.exchange(packet)
	require.NoError(t, err)
	require.Equalf(t, radius.CodeAccessChallenge, resp.Code, "expected Access-Challenge after identity, got %v", resp.Code)

	challenge, err := eap.ParseEAPMessage(resp)
	require.NoError(t, err)
	state := rfc2865.State_Get(resp)
	require.NotEmpty(t, state, "challenge missing State attribute")
	return challenge, state
}

// sendTLSFlight wraps a client TLS flight (or an ACK when flight is empty) in an
// EAP-Response/EAP-TLS message and exchanges it with the server.
func (s *eapTLSSupplicant) sendTLSFlight(t *testing.T, state []byte, ident uint8, flight []byte) *radius.Packet {
	t.Helper()
	var data []byte
	if len(flight) == 0 {
		data = (&tlsfragment.Packet{}).Encode() // ACK: a single zero flags octet
	} else {
		data = (&tlsfragment.Packet{HasLength: true, TLSMessageLength: uint32(len(flight)), Data: flight}).Encode() //nolint:gosec // G115: test TLS flights are far below uint32 max
	}
	eapMsg := &eap.EAPMessage{Code: eap.CodeResponse, Identifier: ident, Type: eap.TypeTLS, Data: data}

	packet := s.newAccessRequest()
	require.NoError(t, rfc2865.State_Set(packet, state))
	eap.SetEAPMessageAndAuth(packet, eapMsg.Encode(), s.secret)

	resp, err := s.exchange(packet)
	require.NoError(t, err)
	return resp
}

func (s *eapTLSSupplicant) newAccessRequest() *radius.Packet {
	packet := radius.New(radius.CodeAccessRequest, []byte(s.secret))
	_ = rfc2865.UserName_SetString(packet, s.username)   //nolint:errcheck
	_ = rfc2865.NASIdentifier_SetString(packet, s.nasID) //nolint:errcheck
	_ = rfc2865.NASIPAddress_Set(packet, s.nasIP)        //nolint:errcheck
	return packet
}

func (s *eapTLSSupplicant) exchange(packet *radius.Packet) (*radius.Packet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return radius.Exchange(ctx, packet, s.serverAddr)
}

// startClient launches the crypto/tls client bound to in-memory duplex streams
// that the supplicant bridges to the RADIUS transport. The default behavior
// runs the handshake only and deliberately does not Close the conn: a
// close_notify alert would surface as an extra non-ACK EAP-TLS round after the
// server's pending-success point (most visibly with a TLS 1.2 pinned client,
// where the client finishes last). Tests may override clientRun to also read
// application data, e.g. the TLS 1.3 protected success indication.
func (s *eapTLSSupplicant) startClient() {
	s.toClient = newEAPStream()
	s.fromClient = newEAPStream()
	conn := &eapConn{rd: s.toClient, wr: s.fromClient}
	client := tls.Client(conn, s.clientCfg)
	s.clientDone = make(chan error, 1)
	run := s.clientRun
	if run == nil {
		run = func(c *tls.Conn) error { return c.Handshake() }
	}
	go func() { s.clientDone <- run(client) }()
}

func (s *eapTLSSupplicant) close() {
	if s.toClient != nil {
		s.toClient.close()
	}
	if s.fromClient != nil {
		s.fromClient.close()
	}
}

// waitSettled blocks until the TLS client is parked waiting for more server bytes
// (its current flight fully written) or until the handshake has finished. Once
// the client goroutine has exited it returns immediately: later rounds (such as
// the server's TLS 1.3 protected success indication) produce no client bytes
// and are answered with a bare ACK.
func (s *eapTLSSupplicant) waitSettled(t *testing.T) {
	t.Helper()
	if s.finished {
		return
	}
	deadline := time.After(5 * time.Second)
	for {
		select {
		case err := <-s.clientDone:
			s.finished = true
			s.clientErr = err
			return
		case <-deadline:
			t.Fatal("timed out waiting for TLS client to settle")
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
			s.clientErr = err
			return
		case <-time.After(2 * time.Millisecond):
		}
	}
}

// nextClientFlight returns the next batch of TLS bytes the client produced.
func (s *eapTLSSupplicant) nextClientFlight(t *testing.T) []byte {
	t.Helper()
	s.waitSettled(t)
	return s.fromClient.drain()
}

// --- in-memory duplex conn for the test TLS client ------------------------

type eapStream struct {
	mu      sync.Mutex
	cond    *sync.Cond
	buf     []byte
	closed  bool
	reading bool
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

// drain returns and clears all buffered bytes.
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
