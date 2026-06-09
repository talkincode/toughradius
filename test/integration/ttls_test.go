//go:build integration

package integration

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2759"
	"layeh.com/radius/rfc2865"

	"github.com/talkincode/toughradius/v9/internal/domain"
	eap "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	eaphandlers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/handlers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsfragment"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/microsoft"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// EAP-TTLS inner AVP codes and field sizes (RFC 5281 §10/§11, RFC 2548). The
// handler keeps these as unexported package constants; the integration package
// mirrors them here so the over-the-wire supplicant stays self-contained.
const (
	ttlsAVPUserName         = 1   // RADIUS attribute 1, non-vendor (RFC 5281 §11.2.1)
	ttlsAVPUserPassword     = 2   // RADIUS attribute 2, non-vendor (RFC 5281 §11.2.5)
	ttlsVendorMicrosoft     = 311 // Microsoft SMI Private Enterprise Code (RFC 2548 §1)
	ttlsAVPMSCHAPChallenge  = 11  // MS-CHAP-Challenge AVP (RFC 2548 §2.1)
	ttlsAVPMSCHAP2Response  = 25  // MS-CHAP2-Response AVP (RFC 2548 §2.3.2)
	ttlsAVPMSCHAP2Success   = 26  // MS-CHAP2-Success AVP (RFC 2548 §2.3.3)
	ttlsMSCHAPChallengeLen  = 16
	ttlsMSCHAP2ResponseLen  = 50 // Ident(1)+Flags(1)+PeerChallenge(16)+Reserved(8)+NTResponse(24)
	ttlsImplicitChallengeSz = 17 // 16-octet MS-CHAP-Challenge + 1-octet Ident (RFC 5281 §11.2.4)
)

// ttlsChallengeLabel is the RFC 5705 exporter label EAP-TTLS uses to derive the
// implicit challenge for legacy challenge-based inner methods (RFC 5281 §11.1).
const ttlsChallengeLabel = "ttls challenge"

// TestTTLSEndToEnd drives a real EAP-TTLSv0 supplicant over live RADIUS UDP
// packets against the running auth server, exercising the inner PAP (M9.3) and
// inner MS-CHAP-V2 (M9.4) tunnels landed for TR-F004. It proves the acceptance
// criteria for the EAP-TTLS milestone (M9.6):
//
//   - inner PAP with valid credentials -> Access-Accept + EAP-Success + the
//     MS-MPPE-Send/Recv-Key attributes derived from the TLS tunnel (RFC 2548);
//   - inner PAP with a wrong password -> Access-Reject + EAP-Failure, no keys;
//   - inner MS-CHAP-V2 with valid credentials -> Access-Accept + MPPE keys,
//     after the supplicant validates the tunneled MS-CHAP2-Success AVP and
//     acknowledges with an empty EAP-TTLS frame (RFC 5281 §11.2.4);
//   - inner MS-CHAP-V2 with a wrong password -> Access-Reject + EAP-Failure.
//
// It is intentionally serial (no t.Parallel) because the RADIUS plugin registry,
// dynamic settings, and rate limiter are process-global shared state. TLS 1.2 is
// pinned on the client so the EAP-TLS framing of the outer tunnel is
// deterministic; the TLS 1.3 app-data path is covered by the tlsengine unit
// tests, and EAP-TTLS itself is pinned to TLS 1.2 server-side anyway.
func TestTTLSEndToEnd(t *testing.T) {
	const secret = "it-ttls-secret"
	suffix := uniqueSuffix()
	nasIP := net.ParseIP("10.203.0.1")
	nasID := "it-ttls-nas-" + suffix

	nas := &domain.NetNas{
		ID:         common.UUIDint64(),
		Identifier: nasID,
		Ipaddr:     nasIP.String(),
		Secret:     secret,
		VendorCode: "0",
		Status:     common.ENABLED,
	}
	require.NoError(t, h.appCtx.DB().Create(nas).Error)

	profileID := seedProfile(t, "it-ttls-profile-"+suffix)
	serverAddr := h.radiusServerAddr()

	// Restore the EAP method after this test so later integration cases are not
	// affected by the runtime switch to eap-ttls.
	restoreEapMethod(t)

	ca := newEAPTLSTestCA(t, "IT TTLS Root CA "+suffix)
	serverCert := ca.issueServer(t, "radius.example.com")
	configureTTLS(t, serverCert)

	clientCfg := func() *tls.Config {
		// EAP-TTLS authenticates the server only; the supplicant trusts the test
		// CA and presents no client certificate. TLS 1.2 is pinned for
		// deterministic EAP-TLS framing (see the function doc).
		return &tls.Config{ //nolint:gosec // G402: TLS 1.2 pin is intentional for deterministic test framing
			RootCAs:    ca.pool(),
			ServerName: "radius.example.com",
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS12,
		}
	}

	t.Run("inner PAP valid credentials authenticate and receive MPPE keys", func(t *testing.T) {
		username := "it-ttls-pap-alice-" + suffix
		password := "it-ttls-pap-pass-" + suffix
		seedTTLSUser(t, profileID, username, password)

		sup := &ttlsSupplicant{
			serverAddr:  serverAddr,
			secret:      secret,
			username:    username,
			password:    password,
			nasID:       nasID,
			nasIP:       nasIP,
			clientCfg:   clientCfg(),
			innerMethod: ttlsInnerPAP,
		}
		resp := sup.authenticate(t)
		require.Equalf(t, radius.CodeAccessAccept, resp.Code,
			"valid TTLS-PAP supplicant must authenticate, got %v (%q)", resp.Code, rfc2865.ReplyMessage_GetString(resp))
		assertEAPCode(t, resp, eap.CodeSuccess)
		assertMPPEKeys(t, resp, secret, sup.lastReqAuth)
	})

	t.Run("inner PAP wrong password is rejected", func(t *testing.T) {
		username := "it-ttls-pap-mallory-" + suffix
		password := "it-ttls-pap-correct-" + suffix
		seedTTLSUser(t, profileID, username, password)

		sup := &ttlsSupplicant{
			serverAddr:  serverAddr,
			secret:      secret,
			username:    username,
			password:    password + "-wrong", // inner User-Password AVP will not match
			nasID:       nasID,
			nasIP:       nasIP,
			clientCfg:   clientCfg(),
			innerMethod: ttlsInnerPAP,
		}
		resp := sup.authenticate(t)
		require.Equalf(t, radius.CodeAccessReject, resp.Code,
			"wrong inner PAP password must be rejected, got %v", resp.Code)
		assertEAPCode(t, resp, eap.CodeFailure)
		assertNoMPPEKeys(t, resp, secret, sup.lastReqAuth)
	})

	t.Run("inner MS-CHAP-V2 valid credentials authenticate and receive MPPE keys", func(t *testing.T) {
		username := "it-ttls-mschap-alice-" + suffix
		password := "it-ttls-mschap-pass-" + suffix
		seedTTLSUser(t, profileID, username, password)

		sup := &ttlsSupplicant{
			serverAddr:  serverAddr,
			secret:      secret,
			username:    username,
			password:    password,
			nasID:       nasID,
			nasIP:       nasIP,
			clientCfg:   clientCfg(),
			innerMethod: ttlsInnerMSCHAPv2,
		}
		resp := sup.authenticate(t)
		require.Equalf(t, radius.CodeAccessAccept, resp.Code,
			"valid TTLS-MSCHAPv2 supplicant must authenticate, got %v (%q)", resp.Code, rfc2865.ReplyMessage_GetString(resp))
		assertEAPCode(t, resp, eap.CodeSuccess)
		assertMPPEKeys(t, resp, secret, sup.lastReqAuth)
	})

	t.Run("inner MS-CHAP-V2 wrong password is rejected", func(t *testing.T) {
		username := "it-ttls-mschap-mallory-" + suffix
		password := "it-ttls-mschap-correct-" + suffix
		seedTTLSUser(t, profileID, username, password)

		sup := &ttlsSupplicant{
			serverAddr:  serverAddr,
			secret:      secret,
			username:    username,
			password:    password + "-wrong", // inner NT-Response will not match
			nasID:       nasID,
			nasIP:       nasIP,
			clientCfg:   clientCfg(),
			innerMethod: ttlsInnerMSCHAPv2,
		}
		resp := sup.authenticate(t)
		require.Equalf(t, radius.CodeAccessReject, resp.Code,
			"wrong inner MS-CHAP-V2 password must be rejected, got %v", resp.Code)
		assertEAPCode(t, resp, eap.CodeFailure)
		assertNoMPPEKeys(t, resp, secret, sup.lastReqAuth)
	})
}

// configureTTLS writes the server certificate/key to disk and points the dynamic
// EAP settings at them, switching the server's EAP method to eap-ttls. EAP-TTLS
// is a server-only TLS method, so no client-CA bundle is required (unlike
// EAP-TLS). The settings provider reads these on every handshake, so the live
// server picks up the change without a restart.
func configureTTLS(t *testing.T, serverCert tls.Certificate) {
	t.Helper()
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server-cert.pem")
	keyPath := filepath.Join(dir, "server-key.pem")

	require.NoError(t, os.WriteFile(certPath, certificatePEM(t, serverCert), 0o600))
	require.NoError(t, os.WriteFile(keyPath, privateKeyPEM(t, serverCert.PrivateKey), 0o600))

	cm := h.appCtx.ConfigMgr()
	require.NoError(t, cm.Set("radius", eaphandlers.SettingEapTlsCertFile, certPath))
	require.NoError(t, cm.Set("radius", eaphandlers.SettingEapTlsKeyFile, keyPath))
	require.NoError(t, cm.Set("radius", "EapMethod", "eap-ttls"))
}

// seedTTLSUser creates an enabled RADIUS user whose plaintext password is used to
// validate the inner PAP User-Password AVP and to compute the inner MS-CHAP-V2
// NT-Response. The default password provider returns the stored password as-is,
// so the supplicant and server derive matching credentials.
func seedTTLSUser(t *testing.T, profileID int64, username, password string) {
	t.Helper()
	user := &domain.RadiusUser{
		ID:         common.UUIDint64(),
		ProfileId:  profileID,
		Username:   username,
		Password:   password,
		Status:     common.ENABLED,
		ExpireTime: time.Now().AddDate(1, 0, 0),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(user).Error)
}

// assertMPPEKeys verifies the Access-Accept carries 32-byte MS-MPPE-Send/Recv
// session keys. They are salt-encrypted (RFC 2548 §2.4) with the Request
// Authenticator of the Access-Request that triggered the Accept, not the reply's
// own authenticator, so both are rebound before decryption.
func assertMPPEKeys(t *testing.T, resp *radius.Packet, secret string, reqAuth [16]byte) {
	t.Helper()
	resp.Secret = []byte(secret)
	resp.Authenticator = reqAuth

	recvKey, err := microsoft.MSMPPERecvKey_Lookup(resp)
	require.NoError(t, err, "Access-Accept must carry an MS-MPPE-Recv-Key")
	assert.Lenf(t, recvKey, 32, "MS-MPPE-Recv-Key must be a 32-byte session key, got %d bytes", len(recvKey))

	sendKey, err := microsoft.MSMPPESendKey_Lookup(resp)
	require.NoError(t, err, "Access-Accept must carry an MS-MPPE-Send-Key")
	assert.Lenf(t, sendKey, 32, "MS-MPPE-Send-Key must be a 32-byte session key, got %d bytes", len(sendKey))
}

// assertNoMPPEKeys verifies a rejected handshake leaks no MPPE session keys.
// Both the Send and Recv keys are checked so a reject can never smuggle either
// half of the session key material.
func assertNoMPPEKeys(t *testing.T, resp *radius.Packet, secret string, reqAuth [16]byte) {
	t.Helper()
	resp.Secret = []byte(secret)
	resp.Authenticator = reqAuth
	_, recvErr := microsoft.MSMPPERecvKey_Lookup(resp)
	assert.ErrorIs(t, recvErr, radius.ErrNoAttribute, "rejects must not carry an MS-MPPE-Recv-Key")
	_, sendErr := microsoft.MSMPPESendKey_Lookup(resp)
	assert.ErrorIs(t, sendErr, radius.ErrNoAttribute, "rejects must not carry an MS-MPPE-Send-Key")
}

// --- over-the-wire EAP-TTLS supplicant ------------------------------------

type ttlsInnerMethod int

const (
	ttlsInnerPAP ttlsInnerMethod = iota
	ttlsInnerMSCHAPv2
)

// ttlsSupplicant drives a real crypto/tls client through the EAP-TTLSv0 framing
// of live RADIUS Access-Request packets: it completes the outer server-only TLS
// handshake, keeps the tls.Conn alive, then — because EAP-TTLS phase 2 is
// peer-initiated (RFC 5281 §7.3) — sends its inner authentication AVPs as TLS
// application data. It mirrors the unit-test ttlsPapPeer/ttlsMschapPeer but over
// the UDP transport, reusing the in-memory duplex streams and CA helpers from
// eap_tls_test.go.
type ttlsSupplicant struct {
	serverAddr  string
	secret      string
	username    string
	password    string
	nasID       string
	nasIP       net.IP
	clientCfg   *tls.Config
	innerMethod ttlsInnerMethod

	toClient   *eapStream // server -> client TLS bytes
	fromClient *eapStream // client -> server TLS bytes
	client     *tls.Conn
	clientDone chan error
	finished   bool

	state     []byte
	respIdent uint8
	// lastReqAuth is the Request Authenticator of the most recent Access-Request,
	// needed to decrypt the salt-encrypted MPPE keys on the final Accept.
	lastReqAuth [16]byte
}

// authenticate runs the whole EAP-TTLS exchange and returns the final RADIUS
// reply (Access-Accept or Access-Reject).
func (s *ttlsSupplicant) authenticate(t *testing.T) *radius.Packet {
	t.Helper()

	// Round 1: EAP-Response/Identity -> Access-Challenge carrying EAP-TTLS Start.
	s.sendIdentity(t)

	s.startClient()
	defer s.close()

	flight := s.nextClientFlight(t) // ClientHello
	require.NotEmpty(t, flight, "TLS client must produce a ClientHello")

	for round := 0; round < 64; round++ {
		resp, serverFlight := s.exchange(t, flight)
		switch resp.Code {
		case radius.CodeAccessAccept, radius.CodeAccessReject:
			return resp
		case radius.CodeAccessChallenge:
			// fall through
		default:
			t.Fatalf("unexpected RADIUS code %v during TTLS outer handshake", resp.Code)
		}

		if len(serverFlight) > 0 {
			_, werr := s.toClient.Write(serverFlight)
			require.NoError(t, werr)
		}
		flight = s.nextClientFlight(t)
		if s.finished {
			// The outer tunnel is established; EAP-TTLS phase 2 is
			// peer-initiated, so the supplicant now speaks first with its inner
			// AVP flight.
			return s.runInner(t)
		}
	}
	t.Fatal("TTLS outer handshake did not complete within the round budget")
	return nil
}

// runInner dispatches to the configured inner authentication method.
func (s *ttlsSupplicant) runInner(t *testing.T) *radius.Packet {
	t.Helper()
	switch s.innerMethod {
	case ttlsInnerPAP:
		return s.runInnerPAP(t)
	case ttlsInnerMSCHAPv2:
		return s.runInnerMSCHAPv2(t)
	default:
		t.Fatalf("unknown inner method %d", s.innerMethod)
		return nil
	}
}

// runInnerPAP tunnels a User-Name + User-Password AVP flight (RFC 5281 §11.2.5),
// a single round: the server validates the password and replies Accept/Reject.
func (s *ttlsSupplicant) runInnerPAP(t *testing.T) *radius.Packet {
	t.Helper()
	flight := ttlsEncodePAPFlight(s.username, s.password)
	resp, _ := s.exchange(t, s.encrypt(t, flight))
	return resp
}

// runInnerMSCHAPv2 tunnels User-Name + MS-CHAP-Challenge + MS-CHAP2-Response AVPs
// (RFC 5281 §11.2.4), validating the implicit challenge derived from the TLS
// session. On success the server tunnels an MS-CHAP2-Success AVP (Access-
// Challenge) and the supplicant acknowledges with an empty EAP-TTLS frame so the
// server grants; a wrong password is rejected outright.
func (s *ttlsSupplicant) runInnerMSCHAPv2(t *testing.T) *radius.Packet {
	t.Helper()

	// Both peers derive the 17-octet implicit challenge from the TLS session
	// (RFC 5281 §11.1): a 16-octet MS-CHAP-Challenge and a 1-octet Ident.
	cs := s.client.ConnectionState()
	chal, err := cs.ExportKeyingMaterial(ttlsChallengeLabel, nil, ttlsImplicitChallengeSz)
	require.NoError(t, err)
	authChallenge := chal[:ttlsMSCHAPChallengeLen]
	ident := chal[ttlsMSCHAPChallengeLen]

	flight := ttlsEncodeMSCHAPv2Flight(t, s.username, s.password, authChallenge, ident)
	resp, serverFlight := s.exchange(t, s.encrypt(t, flight))
	if resp.Code == radius.CodeAccessReject {
		return resp
	}
	require.Equalf(t, radius.CodeAccessChallenge, resp.Code,
		"server must tunnel MS-CHAP2-Success before granting, got %v", resp.Code)

	// The server's flight carries the tunneled MS-CHAP2-Success AVP.
	plain := s.decrypt(t, serverFlight)
	require.GreaterOrEqual(t, len(plain), 14, "MS-CHAP2-Success AVP is too short")
	assert.Equalf(t, uint32(ttlsAVPMSCHAP2Success), binary.BigEndian.Uint32(plain[0:4]),
		"server must tunnel an MS-CHAP2-Success AVP")
	assert.Equal(t, uint32(ttlsVendorMicrosoft), binary.BigEndian.Uint32(plain[8:12]),
		"MS-CHAP2-Success must be a Microsoft vendor AVP")
	assert.Equalf(t, byte('S'), plain[13], "MS-CHAP2-Success must carry an authenticator response string")

	// Acknowledge with a zero-length EAP-TTLS frame (RFC 5281 §11.2.4).
	resp2, _ := s.exchange(t, nil)
	return resp2
}

// exchange sends one outer EAP-Response and reassembles the server's reply,
// acknowledging fragments until a full flight (or a terminal reply) is obtained.
func (s *ttlsSupplicant) exchange(t *testing.T, flight []byte) (*radius.Packet, []byte) {
	t.Helper()
	var buf []byte
	for {
		resp := s.sendTTLSFlight(t, flight)
		if resp.Code != radius.CodeAccessChallenge {
			return resp, nil
		}

		msg, err := eap.ParseEAPMessage(resp)
		require.NoError(t, err)
		require.Equalf(t, uint8(eap.TypeTTLS), msg.Type, "expected EAP-TTLS request, got type %d", msg.Type)
		s.respIdent = msg.Identifier
		if st := rfc2865.State_Get(resp); len(st) > 0 {
			s.state = st
		}

		frag, err := tlsfragment.Parse(msg.Data)
		require.NoError(t, err)
		buf = append(buf, frag.Data...)
		if frag.More() {
			flight = nil // ACK this fragment so the server sends the next one
			continue
		}
		return resp, buf
	}
}

// sendIdentity performs the opening EAP-Response/Identity round and captures the
// EAP-TTLS Start challenge's State and identifier.
func (s *ttlsSupplicant) sendIdentity(t *testing.T) {
	t.Helper()
	identity := &eap.EAPMessage{Code: eap.CodeResponse, Identifier: 1, Type: eap.TypeIdentity, Data: []byte(s.username)}
	packet := s.newAccessRequest()
	eap.SetEAPMessageAndAuth(packet, identity.Encode(), s.secret)
	s.lastReqAuth = packet.Authenticator

	resp, err := s.exchangeRaw(packet)
	require.NoError(t, err)
	require.Equalf(t, radius.CodeAccessChallenge, resp.Code, "expected Access-Challenge after identity, got %v", resp.Code)

	challenge, err := eap.ParseEAPMessage(resp)
	require.NoError(t, err)
	require.Equalf(t, uint8(eap.TypeTTLS), challenge.Type, "expected EAP-TTLS challenge, got EAP type %d", challenge.Type)
	require.GreaterOrEqual(t, len(challenge.Data), 1)
	require.NotZerof(t, challenge.Data[0]&tlsfragment.FlagStart, "expected EAP-TTLS Start (S) flag, got flags %#x", challenge.Data[0])

	s.state = rfc2865.State_Get(resp)
	require.NotEmpty(t, s.state, "challenge missing State attribute")
	s.respIdent = challenge.Identifier
}

// sendTTLSFlight wraps a client TLS flight (or an empty/ACK frame when flight is
// nil) in an EAP-Response/EAP-TTLS message and exchanges it with the server,
// recording the Request Authenticator used for MPPE-key decryption.
func (s *ttlsSupplicant) sendTTLSFlight(t *testing.T, flight []byte) *radius.Packet {
	t.Helper()
	var data []byte
	if len(flight) == 0 {
		data = (&tlsfragment.Packet{}).Encode() // empty EAP-TTLS frame: a single zero flags octet
	} else {
		data = (&tlsfragment.Packet{HasLength: true, TLSMessageLength: uint32(len(flight)), Data: flight}).Encode() //nolint:gosec // G115: test TLS flights are far below uint32 max
	}
	eapMsg := &eap.EAPMessage{Code: eap.CodeResponse, Identifier: s.respIdent, Type: eap.TypeTTLS, Data: data}

	packet := s.newAccessRequest()
	require.NoError(t, rfc2865.State_Set(packet, s.state))
	eap.SetEAPMessageAndAuth(packet, eapMsg.Encode(), s.secret)
	s.lastReqAuth = packet.Authenticator

	resp, err := s.exchangeRaw(packet)
	require.NoError(t, err)
	return resp
}

func (s *ttlsSupplicant) newAccessRequest() *radius.Packet {
	packet := radius.New(radius.CodeAccessRequest, []byte(s.secret))
	_ = rfc2865.UserName_SetString(packet, s.username)   //nolint:errcheck
	_ = rfc2865.NASIdentifier_SetString(packet, s.nasID) //nolint:errcheck
	_ = rfc2865.NASIPAddress_Set(packet, s.nasIP)        //nolint:errcheck
	return packet
}

func (s *ttlsSupplicant) exchangeRaw(packet *radius.Packet) (*radius.Packet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return radius.Exchange(ctx, packet, s.serverAddr)
}

// startClient launches the crypto/tls client handshake bound to in-memory duplex
// streams. The connection is kept open after the handshake so the inner AVP
// exchange can flow as application data.
func (s *ttlsSupplicant) startClient() {
	s.toClient = newEAPStream()
	s.fromClient = newEAPStream()
	conn := &eapConn{rd: s.toClient, wr: s.fromClient}
	s.client = tls.Client(conn, s.clientCfg)
	s.clientDone = make(chan error, 1)
	go func() { s.clientDone <- s.client.Handshake() }()
}

func (s *ttlsSupplicant) close() {
	if s.toClient != nil {
		s.toClient.close()
	}
	if s.fromClient != nil {
		s.fromClient.close()
	}
}

// nextClientFlight returns the next batch of TLS bytes the client produced,
// marking the handshake finished once the client handshake completes.
func (s *ttlsSupplicant) nextClientFlight(t *testing.T) []byte {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case err := <-s.clientDone:
			s.finished = true
			require.NoError(t, err)
			return s.fromClient.drain()
		case <-deadline:
			t.Fatal("timed out waiting for TLS client flight")
		default:
		}
		s.toClient.mu.Lock()
		settled := len(s.toClient.buf) == 0 && s.toClient.reading
		s.toClient.mu.Unlock()
		if settled {
			return s.fromClient.drain()
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// decrypt delivers a server application flight to the TLS client and returns the
// decrypted inner plaintext (the tunneled AVPs).
func (s *ttlsSupplicant) decrypt(t *testing.T, flight []byte) []byte {
	t.Helper()
	_, err := s.toClient.Write(flight)
	require.NoError(t, err)
	buf := make([]byte, 4096)
	n, err := s.client.Read(buf)
	require.NoError(t, err)
	return append([]byte(nil), buf[:n]...)
}

// encrypt encrypts an inner AVP flight through the TLS client and returns the
// resulting TLS records bound for the server.
func (s *ttlsSupplicant) encrypt(t *testing.T, inner []byte) []byte {
	t.Helper()
	_, err := s.client.Write(inner)
	require.NoError(t, err)
	return s.fromClient.drain()
}

// --- inner AVP encoders ----------------------------------------------------

// ttlsEncodeAVP serializes a single EAP-TTLS AVP (RFC 5281 §10.1): AVP Code (4) |
// Flags (1) | AVP Length (3) | [Vendor-ID (4)] | Data, zero-padded to the next
// four-octet boundary; the Length excludes the padding.
func ttlsEncodeAVP(code, vendorID uint32, mandatory bool, data []byte) []byte {
	headerLen := 8
	if vendorID != 0 {
		headerLen = 12
	}
	length := headerLen + len(data)
	buf := make([]byte, (length+3)&^3)
	binary.BigEndian.PutUint32(buf[0:4], code)
	var flags byte
	if vendorID != 0 {
		flags |= 0x80 // V (Vendor-Specific)
	}
	if mandatory {
		flags |= 0x40 // M (Mandatory)
	}
	buf[4] = flags
	buf[5] = byte((length >> 16) & 0xFF)
	buf[6] = byte((length >> 8) & 0xFF)
	buf[7] = byte(length & 0xFF)
	off := 8
	if vendorID != 0 {
		binary.BigEndian.PutUint32(buf[8:12], vendorID)
		off = 12
	}
	copy(buf[off:], data)
	return buf
}

// ttlsEncodePAPFlight builds the inner PAP AVP flight: a User-Name AVP followed
// by a NUL-padded User-Password AVP (RFC 5281 §11.2.5).
func ttlsEncodePAPFlight(username, password string) []byte {
	pw := []byte(password)
	if pad := (16 - len(pw)%16) % 16; pad > 0 {
		pw = append(pw, make([]byte, pad)...)
	}
	var buf []byte
	buf = append(buf, ttlsEncodeAVP(ttlsAVPUserName, 0, true, []byte(username))...)
	buf = append(buf, ttlsEncodeAVP(ttlsAVPUserPassword, 0, true, pw)...)
	return buf
}

// ttlsEncodeMSCHAPv2Flight builds the inner MS-CHAP-V2 AVP flight: a User-Name
// AVP, an MS-CHAP-Challenge AVP echoing the implicitly-derived challenge, and an
// MS-CHAP2-Response AVP carrying the NT-Response computed from the supplicant's
// password (RFC 5281 §11.2.4, RFC 2548 §2.3.2, RFC 2759 §4).
func ttlsEncodeMSCHAPv2Flight(t *testing.T, username, password string, authChallenge []byte, ident uint8) []byte {
	t.Helper()
	peerChallenge, err := eap.GenerateRandomBytes(ttlsMSCHAPChallengeLen)
	require.NoError(t, err)
	nt, err := rfc2759.GenerateNTResponse(authChallenge, peerChallenge, []byte(username), []byte(password))
	require.NoError(t, err)

	respVal := make([]byte, ttlsMSCHAP2ResponseLen)
	respVal[0] = ident
	respVal[1] = 0 // Flags
	copy(respVal[2:18], peerChallenge)
	copy(respVal[26:50], nt)

	var buf []byte
	buf = append(buf, ttlsEncodeAVP(ttlsAVPUserName, 0, true, []byte(username))...)
	buf = append(buf, ttlsEncodeAVP(ttlsAVPMSCHAPChallenge, ttlsVendorMicrosoft, true, authChallenge)...)
	buf = append(buf, ttlsEncodeAVP(ttlsAVPMSCHAP2Response, ttlsVendorMicrosoft, true, respVal)...)
	return buf
}
