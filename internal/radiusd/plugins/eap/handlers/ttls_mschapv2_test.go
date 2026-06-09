package handlers

import (
	"bytes"
	"crypto/tls"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsengine"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsfragment"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/microsoft"
	"layeh.com/radius"
	"layeh.com/radius/rfc2759"
	"layeh.com/radius/rfc2865"
)

// ttlsMschapPeer emulates an EAP-TTLSv0 supplicant performing inner MS-CHAP-V2
// (RFC 5281 §11.2.4). It drives the outer TLS handshake to completion (keeping
// the tls.Client alive) and then, because EAP-TTLS phase 2 is peer-initiated,
// derives the implicit challenge from the TLS session, sends its User-Name /
// MS-CHAP-Challenge / MS-CHAP2-Response AVP flight, validates the server's
// MS-CHAP2-Success AVP, and acknowledges with an empty EAP-TTLS frame, so the
// full two-round inner state machine is exercised end to end.
//
// Like peapPeer it pins TLS 1.2 for deterministic EAP framing; the TLS 1.3
// application-data/key-export path is covered by tlsengine/appdata_test.go.
type ttlsMschapPeer struct {
	t        *testing.T
	h        *TTLSHandler
	sm       eap.EAPStateManager
	stateID  string
	secret   string
	username string
	password string
	user     *domain.RadiusUser
	pwd      eap.PasswordProvider

	// tamperChallenge sends an MS-CHAP-Challenge AVP that does not match the
	// implicitly-derived challenge, exercising the server's binding check.
	tamperChallenge bool

	client     *tls.Conn
	toClient   *hsStream
	fromClient *hsStream
	clientDone chan error
	hsReturned bool

	ident          uint8
	acceptResponse *radius.Packet
}

func newTTLSMschapPeer(t *testing.T, h *TTLSHandler, ca *hsTestCA, sm eap.EAPStateManager, stateID, secret, username, password string) *ttlsMschapPeer {
	t.Helper()
	toClient := newHSStream()
	fromClient := newHSStream()
	conn := &hsConn{rd: toClient, wr: fromClient}
	client := tls.Client(conn, ttlsTunnelClientCfg(ca))

	done := make(chan error, 1)
	go func() { done <- client.Handshake() }()

	return &ttlsMschapPeer{
		t: t, h: h, sm: sm, stateID: stateID, secret: secret,
		username: username, password: password,
		user:       &domain.RadiusUser{Username: username},
		pwd:        &mockPasswordProvider{password: password},
		client:     client,
		toClient:   toClient,
		fromClient: fromClient,
		clientDone: done,
		ident:      20,
	}
}

// run drives the whole EAP-TTLS MS-CHAP-V2 conversation and reports whether the
// handler granted access.
func (p *ttlsMschapPeer) run() (bool, error) {
	respData := p.clientFlight() // ClientHello
	require.NotEmpty(p.t, respData, "client should produce a ClientHello")

	for round := 0; round < 64; round++ {
		if p.hsReturned {
			return p.runInner()
		}
		granted, flight, err := p.exchange(respData)
		if err != nil {
			return false, err
		}
		if granted {
			return true, nil
		}
		if _, werr := p.toClient.Write(flight); werr != nil {
			return false, werr
		}
		respData = p.clientFlight()
	}
	p.t.Fatal("outer handshake did not complete within the round budget")
	return false, nil
}

// runInner carries the two-round inner MS-CHAP-V2 exchange: send the credential
// flight, validate the server's MS-CHAP2-Success, then acknowledge.
func (p *ttlsMschapPeer) runInner() (bool, error) {
	granted, serverFlight, err := p.exchange(p.encrypt(p.mschapFlight()))
	if err != nil {
		return false, err
	}
	if granted {
		return true, nil // a grant before the acknowledgement would be a server bug
	}

	p.assertSuccessAVP(serverFlight)

	// RFC 5281 §11.2.4: acknowledge with a zero-length EAP-TTLS frame.
	granted, _, err = p.exchange(nil)
	return granted, err
}

// mschapFlight builds the inner credential AVP flight: User-Name,
// MS-CHAP-Challenge (echoing the derived challenge) and MS-CHAP2-Response.
func (p *ttlsMschapPeer) mschapFlight() []byte {
	authChallenge, ident := p.implicitChallenge()

	// A deterministic peer challenge keeps the test reproducible; MS-CHAP-V2
	// accepts any 16-octet value.
	peerChallenge := bytes.Repeat([]byte{0xAB}, 16)
	nt, err := rfc2759.GenerateNTResponse(authChallenge, peerChallenge, []byte(p.username), []byte(p.password))
	require.NoError(p.t, err)

	respVal := make([]byte, ttlsMSCHAP2ResponseLen)
	respVal[0] = ident
	respVal[1] = 0 // Flags
	copy(respVal[2:18], peerChallenge)
	copy(respVal[26:50], nt)

	sentChallenge := authChallenge
	if p.tamperChallenge {
		sentChallenge = bytes.Repeat([]byte{0x00}, 16)
	}

	var flight []byte
	flight = append(flight, buildTTLSAVP(ttlsAVPCodeUserName, true, 0, []byte(p.username))...)
	flight = append(flight, buildTTLSAVP(ttlsMSCHAPChallengeCode, true, ttlsVendorMicrosoft, sentChallenge)...)
	flight = append(flight, buildTTLSAVP(ttlsMSCHAP2ResponseCode, true, ttlsVendorMicrosoft, respVal)...)
	return flight
}

// implicitChallenge derives the 17-octet EAP-TTLS implicit challenge from the
// peer's TLS session, exactly as the server does (RFC 5281 §11.1/§11.2.4).
func (p *ttlsMschapPeer) implicitChallenge() (authChallenge []byte, ident uint8) {
	cs := p.client.ConnectionState()
	chal, err := cs.ExportKeyingMaterial(ttlsChallengeLabel, nil, ttlsMSCHAPv2ChallengeLen)
	require.NoError(p.t, err)
	return chal[:16], chal[16]
}

// assertSuccessAVP decrypts the server's flight and asserts it tunneled a
// well-formed MS-CHAP2-Success AVP whose authenticator response the peer accepts.
func (p *ttlsMschapPeer) assertSuccessAVP(serverFlight []byte) {
	plain := p.decrypt(serverFlight)
	avps, err := parseTTLSAVPs(plain)
	require.NoError(p.t, err)
	val, ok := findTTLSVendorAVP(avps, ttlsVendorMicrosoft, ttlsMSCHAP2SuccessCode)
	require.True(p.t, ok, "server must tunnel an MS-CHAP2-Success AVP")
	require.GreaterOrEqual(p.t, len(val), 3)
	assert.Equal(p.t, byte('S'), val[1], "MS-CHAP2-Success must carry an authenticator response string")
}

// exchange sends one outer EAP-Response and reassembles the server's reply,
// acknowledging fragments until a full flight (or a grant) is obtained.
func (p *ttlsMschapPeer) exchange(respData []byte) (granted bool, flight []byte, err error) {
	var buf []byte
	for {
		writer := &mockResponseWriter{}
		ctx := p.responseCtx(writer, respData)
		ok, herr := p.h.HandleResponse(ctx)
		if herr != nil {
			return false, nil, herr
		}
		if ok {
			p.acceptResponse = ctx.Response
			return true, nil, nil
		}
		require.NotNil(p.t, writer.response, "handler must answer with a challenge")
		require.Equal(p.t, radius.CodeAccessChallenge, writer.response.Code)

		frag := ttlsParseChallenge(p.t, writer.response)
		buf = append(buf, frag.Data...)
		if frag.More() {
			respData = nil // ACK
			continue
		}
		return false, buf, nil
	}
}

func (p *ttlsMschapPeer) responseCtx(writer *mockResponseWriter, tlsData []byte) *eap.EAPContext {
	packet := radius.New(radius.CodeAccessRequest, []byte(p.secret))
	require.NoError(p.t, rfc2865.State_SetString(packet, p.stateID))
	p.ident++
	var data []byte
	if len(tlsData) == 0 {
		data = (&tlsfragment.Packet{}).Encode()
	} else {
		data = (&tlsfragment.Packet{HasLength: true, TLSMessageLength: uint32(len(tlsData)), Data: tlsData}).Encode() //nolint:gosec // G115: test payloads are small
	}
	req := &radius.Request{Packet: packet}
	return &eap.EAPContext{
		Request:        req,
		Response:       req.Response(radius.CodeAccessAccept),
		ResponseWriter: writer,
		StateManager:   p.sm,
		Secret:         p.secret,
		User:           p.user,
		PwdProvider:    p.pwd,
		EAPMessage:     &eap.EAPMessage{Code: eap.CodeResponse, Identifier: p.ident, Type: eap.TypeTTLS, Data: data},
	}
}

// clientFlight returns the next batch of TLS bytes the client produced, marking
// hsReturned once the client handshake completes.
func (p *ttlsMschapPeer) clientFlight() []byte {
	deadline := time.After(5 * time.Second)
	for {
		select {
		case err := <-p.clientDone:
			p.hsReturned = true
			require.NoError(p.t, err)
			return p.fromClient.drain()
		case <-deadline:
			p.t.Fatal("timed out waiting for TLS client flight")
		default:
		}
		p.toClient.mu.Lock()
		settled := len(p.toClient.buf) == 0 && p.toClient.reading
		p.toClient.mu.Unlock()
		if settled {
			return p.fromClient.drain()
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// decrypt delivers a server application flight to the TLS client and returns the
// decrypted inner plaintext (the tunneled AVPs).
func (p *ttlsMschapPeer) decrypt(flight []byte) []byte {
	_, err := p.toClient.Write(flight)
	require.NoError(p.t, err)
	buf := make([]byte, 4096)
	n, err := p.client.Read(buf)
	require.NoError(p.t, err)
	return append([]byte(nil), buf[:n]...)
}

// encrypt encrypts an inner AVP flight through the TLS client and returns the
// resulting TLS records.
func (p *ttlsMschapPeer) encrypt(plaintext []byte) []byte {
	_, err := p.client.Write(plaintext)
	require.NoError(p.t, err)
	return p.fromClient.drain()
}

// --- end-to-end inner MS-CHAP-V2 tests ------------------------------------

func TestTTLSHandler_FullMSCHAPv2Auth_Succeeds(t *testing.T) {
	ca := newHSTestCA(t, "TTLS MSCHAPv2 Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	peer := newTTLSMschapPeer(t, h, ca, sm, stateID, "secret", "ttlsuser", "S3cr3t!")
	success, err := peer.run()
	require.NoError(t, err)
	assert.True(t, success, "EAP-TTLS inner MS-CHAP-V2 with a correct password must authenticate")

	state, err := sm.GetState(stateID)
	require.NoError(t, err)
	assert.True(t, state.Success)
	assert.Equal(t, "ttlsuser", getString(state, stateKeyInnerIdentity),
		"the inner User-Name AVP must be captured on the state")

	require.NotNil(t, peer.acceptResponse)
	_, rerr := microsoft.MSMPPERecvKey_Lookup(peer.acceptResponse)
	assert.NoError(t, rerr, "MS-MPPE-Recv-Key must be present on the Access-Accept")
	_, serr := microsoft.MSMPPESendKey_Lookup(peer.acceptResponse)
	assert.NoError(t, serr, "MS-MPPE-Send-Key must be present on the Access-Accept")
}

func TestTTLSHandler_FullMSCHAPv2Auth_Fragmented(t *testing.T) {
	ca := newHSTestCA(t, "TTLS MSCHAPv2 Fragment Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })
	h.maxFragment = 48 // force the outer handshake and inner MS-CHAP2-Success to fragment

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	peer := newTTLSMschapPeer(t, h, ca, sm, stateID, "secret", "ttlsuser", "S3cr3t!")
	success, err := peer.run()
	require.NoError(t, err)
	assert.True(t, success, "fragmented EAP-TTLS outer handshake must still authenticate inner MS-CHAP-V2")
}

func TestTTLSHandler_MSCHAPv2Auth_WrongPasswordRejected(t *testing.T) {
	ca := newHSTestCA(t, "TTLS MSCHAPv2 WrongPwd Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	// The client computes its NT-Response with "wrong"; the server checks "correct".
	peer := newTTLSMschapPeer(t, h, ca, sm, stateID, "secret", "ttlsuser", "wrong")
	peer.pwd = &mockPasswordProvider{password: "correct"}
	success, err := peer.run()
	assert.False(t, success, "a wrong inner MS-CHAP-V2 password must not authenticate")
	require.Error(t, err)
	assert.ErrorIs(t, err, eap.ErrPasswordMismatch)

	state, err := sm.GetState(stateID)
	require.NoError(t, err)
	assert.False(t, state.Success)
}

func TestTTLSHandler_MSCHAPv2Auth_TamperedChallengeRejected(t *testing.T) {
	ca := newHSTestCA(t, "TTLS MSCHAPv2 BadChallenge Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	peer := newTTLSMschapPeer(t, h, ca, sm, stateID, "secret", "ttlsuser", "S3cr3t!")
	peer.tamperChallenge = true
	success, err := peer.run()
	assert.False(t, success, "an MS-CHAP-Challenge AVP that does not match the derived challenge must be rejected")
	require.Error(t, err)
	assert.ErrorIs(t, err, eap.ErrTTLSInnerProtocol)
}
