//go:build integration

package integration

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"net"
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

// Inner EAP-MSCHAPv2 opcodes and field sizes (RFC 2759). The handler keeps these
// as unexported package constants; the integration package mirrors them here so
// the over-the-wire supplicant stays self-contained.
const (
	peapMSCHAPv2Challenge   = 1
	peapMSCHAPv2Response    = 2
	peapMSCHAPv2Success     = 3
	peapMSCHAPv2ChallengeSz = 16
	peapMSCHAPv2ResponseSz  = 49 // PeerChallenge(16) + Reserved(8) + NTResponse(24) + Flags(1)
)

// TestPEAPMSCHAPv2EndToEnd drives a real PEAPv0 / EAP-MSCHAPv2 supplicant over
// live RADIUS UDP packets against the running auth server, exercising the inner
// EAP-MSCHAPv2 tunnel landed in milestone M8.3b (TR-F004). It proves the
// acceptance criteria for the PEAP milestone (M8.5):
//
//   - valid credentials produce an Access-Accept carrying EAP-Success and the
//     MS-MPPE-Send/Recv-Key attributes derived from the TLS tunnel (RFC 2548);
//   - a wrong inner password produces an Access-Reject carrying EAP-Failure,
//     with no MPPE keys leaked.
//
// It is intentionally serial (no t.Parallel) because the RADIUS plugin registry,
// dynamic settings, and rate limiter are process-global shared state. TLS 1.2 is
// pinned on the client so the EAP-TLS framing of the outer tunnel is
// deterministic (no post-handshake NewSessionTicket round); the TLS 1.3 app-data
// path is covered by the tlsengine unit tests.
func TestPEAPMSCHAPv2EndToEnd(t *testing.T) {
	const secret = "it-peap-secret"
	suffix := uniqueSuffix()
	nasIP := net.ParseIP("10.202.0.1")
	nasID := "it-peap-nas-" + suffix

	nas := &domain.NetNas{
		ID:         common.UUIDint64(),
		Identifier: nasID,
		Ipaddr:     nasIP.String(),
		Secret:     secret,
		VendorCode: "0",
		Status:     common.ENABLED,
	}
	require.NoError(t, h.appCtx.DB().Create(nas).Error)

	profileID := seedProfile(t, "it-peap-profile-"+suffix)
	serverAddr := h.radiusServerAddr()

	// Restore the EAP method after this test so later integration cases are not
	// affected by the runtime switch to eap-peap.
	restoreEapMethod(t)

	ca := newEAPTLSTestCA(t, "IT PEAP Root CA "+suffix)
	serverCert := ca.issueServer(t, "radius.example.com")
	configurePEAP(t, serverCert)

	clientCfg := func() *tls.Config {
		// PEAP authenticates the server only; the supplicant trusts the test CA
		// and presents no client certificate. TLS 1.2 is pinned for deterministic
		// EAP-TLS framing (see the function doc).
		return &tls.Config{ //nolint:gosec // G402: TLS 1.2 pin is intentional for deterministic test framing
			RootCAs:    ca.pool(),
			ServerName: "radius.example.com",
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS12,
		}
	}

	t.Run("valid credentials authenticate and receive MPPE keys", func(t *testing.T) {
		username := "it-peap-alice-" + suffix
		password := "it-peap-pass-" + suffix
		seedPEAPUser(t, profileID, username, password)

		sup := &peapSupplicant{
			serverAddr: serverAddr,
			secret:     secret,
			username:   username,
			password:   password,
			nasID:      nasID,
			nasIP:      nasIP,
			clientCfg:  clientCfg(),
		}
		resp := sup.authenticate(t)
		require.Equalf(t, radius.CodeAccessAccept, resp.Code,
			"valid PEAP-MSCHAPv2 supplicant must authenticate, got %v (%q)", resp.Code, rfc2865.ReplyMessage_GetString(resp))
		assertEAPCode(t, resp, eap.CodeSuccess)

		// MS-MPPE-Send/Recv-Key are salt-encrypted (RFC 2548 §2.4) with the
		// Request Authenticator of the Access-Request that triggered the Accept,
		// not the reply's own authenticator. Rebind both before decrypting.
		resp.Secret = []byte(secret)
		resp.Authenticator = sup.lastReqAuth

		recvKey, err := microsoft.MSMPPERecvKey_Lookup(resp)
		require.NoError(t, err, "Access-Accept must carry an MS-MPPE-Recv-Key")
		assert.Lenf(t, recvKey, 32, "MS-MPPE-Recv-Key must be a 32-byte session key, got %d bytes", len(recvKey))

		sendKey, err := microsoft.MSMPPESendKey_Lookup(resp)
		require.NoError(t, err, "Access-Accept must carry an MS-MPPE-Send-Key")
		assert.Lenf(t, sendKey, 32, "MS-MPPE-Send-Key must be a 32-byte session key, got %d bytes", len(sendKey))
	})

	t.Run("wrong password is rejected", func(t *testing.T) {
		username := "it-peap-mallory-" + suffix
		password := "it-peap-correct-" + suffix
		seedPEAPUser(t, profileID, username, password)

		sup := &peapSupplicant{
			serverAddr: serverAddr,
			secret:     secret,
			username:   username,
			password:   password + "-wrong", // inner NT-Response will not match
			nasID:      nasID,
			nasIP:      nasIP,
			clientCfg:  clientCfg(),
		}
		resp := sup.authenticate(t)
		require.Equalf(t, radius.CodeAccessReject, resp.Code,
			"wrong inner password must be rejected, got %v", resp.Code)
		assertEAPCode(t, resp, eap.CodeFailure)

		// A rejected handshake must not leak MPPE session keys.
		resp.Secret = []byte(secret)
		resp.Authenticator = sup.lastReqAuth
		_, err := microsoft.MSMPPERecvKey_Lookup(resp)
		assert.ErrorIs(t, err, radius.ErrNoAttribute, "rejects must not carry MPPE keys")
	})
}

// configurePEAP stores the server certificate/key as a managed certificate and
// points the dynamic EAP settings at it by name, switching the server's EAP
// method to eap-peap. PEAP is a server-only TLS method, so no client-CA bundle
// is required (unlike EAP-TLS). The settings provider resolves the managed
// certificate on every handshake, so the live server picks up the change without
// a restart.
func configurePEAP(t *testing.T, serverCert tls.Certificate) {
	t.Helper()
	serverName := seedManagedServerCert(t, serverCert)

	cm := h.appCtx.ConfigMgr()
	require.NoError(t, cm.Set("radius", eaphandlers.SettingEapTlsServerCert, serverName))
	require.NoError(t, cm.Set("radius", "EapMethod", "eap-peap"))
}

// seedPEAPUser creates an enabled RADIUS user whose plaintext password is used to
// compute the inner EAP-MSCHAPv2 NT-Response. The default password provider
// returns the stored password as-is, so the supplicant and server derive matching
// challenge responses (RFC 2759 §4).
func seedPEAPUser(t *testing.T, profileID int64, username, password string) {
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

// --- over-the-wire PEAP-MSCHAPv2 supplicant -------------------------------

// peapSupplicant drives a real crypto/tls client through the PEAPv0 framing of
// live RADIUS Access-Request packets: it completes the outer EAP-TLS handshake,
// keeps the tls.Conn alive, then carries the inner EAP-MSCHAPv2 exchange as TLS
// application data. It mirrors the unit-test peapPeer but over the UDP transport,
// reusing the in-memory duplex streams and CA helpers from eap_tls_test.go.
type peapSupplicant struct {
	serverAddr string
	secret     string
	username   string
	password   string
	nasID      string
	nasIP      net.IP
	clientCfg  *tls.Config

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

// authenticate runs the whole PEAP-MSCHAPv2 exchange and returns the final RADIUS
// reply (Access-Accept or Access-Reject).
func (s *peapSupplicant) authenticate(t *testing.T) *radius.Packet {
	t.Helper()

	// Round 1: EAP-Response/Identity -> Access-Challenge carrying PEAP Start.
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
			t.Fatalf("unexpected RADIUS code %v during PEAP outer handshake", resp.Code)
		}

		if s.finished {
			// The outer tunnel is established; serverFlight carries the first
			// inner EAP request as TLS application data.
			return s.runInner(t, serverFlight)
		}
		if len(serverFlight) > 0 {
			_, werr := s.toClient.Write(serverFlight)
			require.NoError(t, werr)
		}
		flight = s.nextClientFlight(t)
	}
	t.Fatal("PEAP outer handshake did not complete within the round budget")
	return nil
}

// runInner runs the tunneled EAP-MSCHAPv2 exchange starting from the server's
// first inner request flight.
func (s *peapSupplicant) runInner(t *testing.T, flight []byte) *radius.Packet {
	t.Helper()
	for round := 0; round < 32; round++ {
		innerReq := s.decrypt(t, flight)
		innerResp := s.respondInner(t, innerReq)
		out := s.encrypt(t, innerResp)

		resp, next := s.exchange(t, out)
		switch resp.Code {
		case radius.CodeAccessAccept, radius.CodeAccessReject:
			return resp
		case radius.CodeAccessChallenge:
			// continue the inner exchange
		default:
			t.Fatalf("unexpected RADIUS code %v during PEAP inner exchange", resp.Code)
		}
		flight = next
	}
	t.Fatal("PEAP inner exchange did not complete within the round budget")
	return nil
}

// exchange sends one outer EAP-Response and reassembles the server's reply,
// acknowledging fragments until a full flight (or a terminal reply) is obtained.
func (s *peapSupplicant) exchange(t *testing.T, flight []byte) (*radius.Packet, []byte) {
	t.Helper()
	var buf []byte
	for {
		resp := s.sendPEAPFlight(t, flight)
		if resp.Code != radius.CodeAccessChallenge {
			return resp, nil
		}

		msg, err := eap.ParseEAPMessage(resp)
		require.NoError(t, err)
		require.Equalf(t, uint8(eap.TypePEAP), msg.Type, "expected EAP-PEAP request, got type %d", msg.Type)
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
// PEAP Start challenge's State and identifier.
func (s *peapSupplicant) sendIdentity(t *testing.T) {
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
	require.Equalf(t, uint8(eap.TypePEAP), challenge.Type, "expected EAP-PEAP challenge, got EAP type %d", challenge.Type)
	require.GreaterOrEqual(t, len(challenge.Data), 1)
	require.NotZerof(t, challenge.Data[0]&tlsfragment.FlagStart, "expected PEAP Start (S) flag, got flags %#x", challenge.Data[0])

	s.state = rfc2865.State_Get(resp)
	require.NotEmpty(t, s.state, "challenge missing State attribute")
	s.respIdent = challenge.Identifier
}

// sendPEAPFlight wraps a client TLS flight (or an ACK when flight is empty) in an
// EAP-Response/EAP-PEAP message and exchanges it with the server, recording the
// Request Authenticator used for MPPE-key decryption.
func (s *peapSupplicant) sendPEAPFlight(t *testing.T, flight []byte) *radius.Packet {
	t.Helper()
	var data []byte
	if len(flight) == 0 {
		data = (&tlsfragment.Packet{}).Encode() // ACK: a single zero flags octet
	} else {
		data = (&tlsfragment.Packet{HasLength: true, TLSMessageLength: uint32(len(flight)), Data: flight}).Encode() //nolint:gosec // G115: test TLS flights are far below uint32 max
	}
	eapMsg := &eap.EAPMessage{Code: eap.CodeResponse, Identifier: s.respIdent, Type: eap.TypePEAP, Data: data}

	packet := s.newAccessRequest()
	require.NoError(t, rfc2865.State_Set(packet, s.state))
	eap.SetEAPMessageAndAuth(packet, eapMsg.Encode(), s.secret)
	s.lastReqAuth = packet.Authenticator

	resp, err := s.exchangeRaw(packet)
	require.NoError(t, err)
	return resp
}

func (s *peapSupplicant) newAccessRequest() *radius.Packet {
	packet := radius.New(radius.CodeAccessRequest, []byte(s.secret))
	_ = rfc2865.UserName_SetString(packet, s.username)   //nolint:errcheck
	_ = rfc2865.NASIdentifier_SetString(packet, s.nasID) //nolint:errcheck
	_ = rfc2865.NASIPAddress_Set(packet, s.nasIP)        //nolint:errcheck
	return packet
}

func (s *peapSupplicant) exchangeRaw(packet *radius.Packet) (*radius.Packet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return radius.Exchange(ctx, packet, s.serverAddr)
}

// startClient launches the crypto/tls client handshake bound to in-memory duplex
// streams. Unlike the EAP-TLS supplicant, the connection is kept open after the
// handshake so the inner EAP-MSCHAPv2 exchange can flow as application data.
func (s *peapSupplicant) startClient() {
	s.toClient = newEAPStream()
	s.fromClient = newEAPStream()
	conn := &eapConn{rd: s.toClient, wr: s.fromClient}
	s.client = tls.Client(conn, s.clientCfg)
	s.clientDone = make(chan error, 1)
	go func() { s.clientDone <- s.client.Handshake() }()
}

func (s *peapSupplicant) close() {
	if s.toClient != nil {
		s.toClient.close()
	}
	if s.fromClient != nil {
		s.fromClient.close()
	}
}

// nextClientFlight returns the next batch of TLS bytes the client produced,
// marking the handshake finished once the client handshake completes.
func (s *peapSupplicant) nextClientFlight(t *testing.T) []byte {
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
// decrypted inner EAP packet.
func (s *peapSupplicant) decrypt(t *testing.T, flight []byte) []byte {
	t.Helper()
	_, err := s.toClient.Write(flight)
	require.NoError(t, err)
	buf := make([]byte, 4096)
	n, err := s.client.Read(buf)
	require.NoError(t, err)
	return append([]byte(nil), buf[:n]...)
}

// encrypt encrypts an inner EAP packet through the TLS client and returns the
// resulting TLS records bound for the server.
func (s *peapSupplicant) encrypt(t *testing.T, inner []byte) []byte {
	t.Helper()
	_, err := s.client.Write(inner)
	require.NoError(t, err)
	return s.fromClient.drain()
}

// respondInner produces the client's inner EAP response to a server inner EAP
// request, implementing the supplicant side of EAP-MSCHAPv2 (RFC 2759).
func (s *peapSupplicant) respondInner(t *testing.T, req []byte) []byte {
	t.Helper()
	msg := decodeInnerEAP(t, req)
	require.Equal(t, uint8(eap.CodeRequest), msg.Code)

	switch msg.Type {
	case eap.TypeIdentity:
		return (&eap.EAPMessage{
			Code:       eap.CodeResponse,
			Identifier: msg.Identifier,
			Type:       eap.TypeIdentity,
			Data:       []byte(s.username),
		}).Encode()
	case eap.TypeMSCHAPv2:
		require.NotEmpty(t, msg.Data)
		switch msg.Data[0] {
		case peapMSCHAPv2Challenge:
			return s.buildClientChallengeResponse(t, msg)
		case peapMSCHAPv2Success:
			// Acknowledge the server's success with a bare Success opcode.
			return (&eap.EAPMessage{
				Code:       eap.CodeResponse,
				Identifier: msg.Identifier,
				Type:       eap.TypeMSCHAPv2,
				Data:       []byte{peapMSCHAPv2Success},
			}).Encode()
		default:
			t.Fatalf("unexpected inner MSCHAPv2 opcode: %d", msg.Data[0])
		}
	default:
		t.Fatalf("unexpected inner EAP type: %d", msg.Type)
	}
	return nil
}

// buildClientChallengeResponse computes the EAP-MSCHAPv2 Response for a server
// Challenge using the supplicant's configured password (RFC 2759 §4).
func (s *peapSupplicant) buildClientChallengeResponse(t *testing.T, challenge *eap.EAPMessage) []byte {
	t.Helper()
	// Challenge layout after the EAP Type byte:
	// OpCode(1) MS-ID(1) MS-Length(2) Value-Size(1) Challenge(16) Name(...)
	require.GreaterOrEqual(t, len(challenge.Data), 5+peapMSCHAPv2ChallengeSz)
	msID := challenge.Data[1]
	authChallenge := challenge.Data[5 : 5+peapMSCHAPv2ChallengeSz]

	peerChallenge, err := eap.GenerateRandomBytes(peapMSCHAPv2ChallengeSz)
	require.NoError(t, err)

	ntResponse, err := rfc2759.GenerateNTResponse(authChallenge, peerChallenge, []byte(s.username), []byte(s.password))
	require.NoError(t, err)

	value := make([]byte, peapMSCHAPv2ResponseSz) // PeerChallenge(16) Reserved(8) NTResponse(24) Flags(1)
	copy(value[0:16], peerChallenge)
	copy(value[24:48], ntResponse)

	name := []byte(s.username)
	msLen := 5 + peapMSCHAPv2ResponseSz + len(name) // OpCode+MSID+MSLength(2)+ValueSize(1)+value+name
	data := make([]byte, msLen)
	data[0] = peapMSCHAPv2Response
	data[1] = msID
	binary.BigEndian.PutUint16(data[2:4], uint16(msLen)) //nolint:gosec // G115: bounded by EAP packet size
	data[4] = peapMSCHAPv2ResponseSz
	copy(data[5:], value)
	copy(data[5+peapMSCHAPv2ResponseSz:], name)

	return (&eap.EAPMessage{
		Code:       eap.CodeResponse,
		Identifier: challenge.Identifier,
		Type:       eap.TypeMSCHAPv2,
		Data:       data,
	}).Encode()
}

// decodeInnerEAP parses a raw inner EAP packet (carried as TLS application data)
// into its header fields. Inner requests in this exchange always carry a Type
// byte (Identity or MSCHAPv2), so a 5-octet minimum is enforced.
func decodeInnerEAP(t *testing.T, data []byte) *eap.EAPMessage {
	t.Helper()
	require.GreaterOrEqualf(t, len(data), 5, "inner EAP packet too short: %d bytes", len(data))
	declared := binary.BigEndian.Uint16(data[2:4])
	require.Equalf(t, len(data), int(declared), "inner EAP length %d != actual %d", declared, len(data))
	return &eap.EAPMessage{
		Code:       data[0],
		Identifier: data[1],
		Length:     declared,
		Type:       data[4],
		Data:       append([]byte(nil), data[5:]...),
	}
}
