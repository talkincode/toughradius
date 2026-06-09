package handlers

import (
	"crypto/tls"
	"encoding/binary"
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
	"layeh.com/radius/rfc2869"
)

// peapPeer emulates a PEAPv0 / EAP-MSCHAPv2 supplicant. It drives the outer TLS
// handshake to completion (keeping the tls.Client alive, unlike supplicant) and
// then carries the inner EAP-MSCHAPv2 exchange as TLS application data, so the
// full PEAP inner state machine (handleInnerEAP) can be exercised end to end.
//
// The harness forces TLS 1.2 so the EAP-TLS framing is deterministic (no
// post-handshake NewSessionTicket round); the tlsengine appdata_test.go already
// proves application data and key export are identical under TLS 1.3.
type peapPeer struct {
	t        *testing.T
	h        *PEAPHandler
	sm       eap.EAPStateManager
	stateID  string
	secret   string
	username string
	password string
	user     *domain.RadiusUser
	pwd      eap.PasswordProvider

	client     *tls.Conn
	toClient   *hsStream
	fromClient *hsStream
	clientDone chan error
	hsReturned bool

	ident          uint8
	acceptResponse *radius.Packet
}

func newPeapPeerWithCA(t *testing.T, h *PEAPHandler, ca *hsTestCA, sm eap.EAPStateManager, stateID, secret, username, password string) *peapPeer {
	t.Helper()
	toClient := newHSStream()
	fromClient := newHSStream()
	conn := &hsConn{rd: toClient, wr: fromClient}
	// Pin TLS 1.2 for deterministic EAP framing: TLS 1.3 adds a post-handshake
	// NewSessionTicket round that complicates the inner-EAP state machine. The
	// TLS 1.3 app-data/key-export path is covered by tlsengine/appdata_test.go.
	clientCfg := &tls.Config{ //nolint:gosec // G402: TLS 1.2 pin is intentional for deterministic test framing
		RootCAs:    ca.pool,
		ServerName: "radius.example.com",
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS12,
	}
	client := tls.Client(conn, clientCfg)

	done := make(chan error, 1)
	go func() { done <- client.Handshake() }()

	return &peapPeer{
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

// run drives the whole PEAP conversation and returns whether the handler granted
// access.
func (p *peapPeer) run() (bool, error) {
	respData := p.clientFlight() // ClientHello
	require.NotEmpty(p.t, respData, "client should produce a ClientHello")

	for round := 0; round < 64; round++ {
		granted, flight, err := p.exchange(respData)
		if err != nil {
			return false, err
		}
		if granted {
			return true, nil
		}
		if p.hsReturned {
			// The outer tunnel is established; `flight` carries the first inner
			// EAP request as TLS application data.
			return p.runInner(flight)
		}
		if _, werr := p.toClient.Write(flight); werr != nil {
			return false, werr
		}
		respData = p.clientFlight()
	}
	p.t.Fatal("outer handshake did not complete within the round budget")
	return false, nil
}

// runInner runs the tunneled EAP-MSCHAPv2 exchange starting from the server's
// first inner request flight.
func (p *peapPeer) runInner(flight []byte) (bool, error) {
	for round := 0; round < 16; round++ {
		innerReq := p.decrypt(flight)
		innerResp, err := p.respondInner(innerReq)
		if err != nil {
			return false, err
		}
		respData := p.encrypt(innerResp)

		granted, next, eerr := p.exchange(respData)
		if eerr != nil {
			return false, eerr
		}
		if granted {
			return true, nil
		}
		flight = next
	}
	p.t.Fatal("inner exchange did not complete within the round budget")
	return false, nil
}

// exchange sends one outer EAP-Response and reassembles the server's reply,
// acknowledging fragments until a full flight (or a grant) is obtained.
func (p *peapPeer) exchange(respData []byte) (granted bool, flight []byte, err error) {
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

		frag := peapParseChallenge(p.t, writer.response)
		buf = append(buf, frag.Data...)
		if frag.More() {
			respData = nil // ACK
			continue
		}
		return false, buf, nil
	}
}

func (p *peapPeer) responseCtx(writer *mockResponseWriter, tlsData []byte) *eap.EAPContext {
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
		EAPMessage:     &eap.EAPMessage{Code: eap.CodeResponse, Identifier: p.ident, Type: eap.TypePEAP, Data: data},
	}
}

// clientFlight returns the next batch of TLS bytes the client produced, marking
// hsReturned once the client handshake completes.
func (p *peapPeer) clientFlight() []byte {
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
// decrypted inner EAP packet.
func (p *peapPeer) decrypt(flight []byte) []byte {
	_, err := p.toClient.Write(flight)
	require.NoError(p.t, err)
	buf := make([]byte, 4096)
	n, err := p.client.Read(buf)
	require.NoError(p.t, err)
	return append([]byte(nil), buf[:n]...)
}

// encrypt encrypts an inner EAP packet through the TLS client and returns the
// resulting TLS records.
func (p *peapPeer) encrypt(inner []byte) []byte {
	_, err := p.client.Write(inner)
	require.NoError(p.t, err)
	return p.fromClient.drain()
}

// respondInner produces the client's inner EAP response to a server inner EAP
// request, implementing the supplicant side of EAP-MSCHAPv2 (RFC 2759).
func (p *peapPeer) respondInner(req []byte) ([]byte, error) {
	msg, err := parseInnerEAP(req)
	require.NoError(p.t, err)
	require.Equal(p.t, uint8(eap.CodeRequest), msg.Code)

	switch msg.Type {
	case eap.TypeIdentity:
		return (&eap.EAPMessage{
			Code:       eap.CodeResponse,
			Identifier: msg.Identifier,
			Type:       eap.TypeIdentity,
			Data:       []byte(p.username),
		}).Encode(), nil
	case eap.TypeMSCHAPv2:
		require.NotEmpty(p.t, msg.Data)
		switch msg.Data[0] {
		case MSCHAPv2Challenge:
			return p.buildClientChallengeResponse(msg)
		case MSCHAPv2Success:
			// Acknowledge the server's success with a bare Success opcode.
			return (&eap.EAPMessage{
				Code:       eap.CodeResponse,
				Identifier: msg.Identifier,
				Type:       eap.TypeMSCHAPv2,
				Data:       []byte{MSCHAPv2Success},
			}).Encode(), nil
		default:
			p.t.Fatalf("unexpected inner MSCHAPv2 opcode: %d", msg.Data[0])
		}
	default:
		p.t.Fatalf("unexpected inner EAP type: %d", msg.Type)
	}
	return nil, nil
}

// buildClientChallengeResponse computes the EAP-MSCHAPv2 Response for a server
// Challenge using the peer's configured password (RFC 2759 §4).
func (p *peapPeer) buildClientChallengeResponse(challenge *eap.EAPMessage) ([]byte, error) {
	// Challenge layout after the EAP Type byte:
	// OpCode(1) MS-ID(1) MS-Length(2) Value-Size(1) Challenge(16) Name(...)
	require.GreaterOrEqual(p.t, len(challenge.Data), 5+MSCHAPChallengeSize)
	msID := challenge.Data[1]
	authChallenge := challenge.Data[5 : 5+MSCHAPChallengeSize]

	peerChallenge, err := eap.GenerateRandomBytes(MSCHAPChallengeSize)
	require.NoError(p.t, err)

	ntResponse, err := rfc2759.GenerateNTResponse(authChallenge, peerChallenge, []byte(p.username), []byte(p.password))
	require.NoError(p.t, err)

	value := make([]byte, MSCHAPResponseSize) // PeerChallenge(16) Reserved(8) NTResponse(24) Flags(1)
	copy(value[0:16], peerChallenge)
	copy(value[24:48], ntResponse)

	name := []byte(p.username)
	msLen := 5 + MSCHAPResponseSize + len(name) // OpCode+MSID+MSLength(2)+ValueSize(1)+value+name
	data := make([]byte, msLen)
	data[0] = MSCHAPv2Response
	data[1] = msID
	binary.BigEndian.PutUint16(data[2:4], uint16(msLen)) //nolint:gosec // G115: bounded by EAP packet size
	data[4] = MSCHAPResponseSize
	copy(data[5:], value)
	copy(data[5+MSCHAPResponseSize:], name)

	return (&eap.EAPMessage{
		Code:       eap.CodeResponse,
		Identifier: challenge.Identifier,
		Type:       eap.TypeMSCHAPv2,
		Data:       data,
	}).Encode(), nil
}

func peapParseChallenge(t *testing.T, resp *radius.Packet) *tlsfragment.Packet {
	t.Helper()
	eapData, err := rfc2869.EAPMessage_Lookup(resp)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(eapData), 5)
	require.Equal(t, uint8(eap.TypePEAP), eapData[4])
	frag, err := tlsfragment.Parse(eapData[5:])
	require.NoError(t, err)
	return frag
}

// --- end-to-end inner-auth tests ------------------------------------------

func TestPEAPHandler_FullInnerAuth_Succeeds(t *testing.T) {
	ca := newHSTestCA(t, "PEAP Inner Root CA")
	cfg := peapServerEngineConfig(t, ca)
	h := NewPEAPHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "peapuser", "secret")
	peer := newPeapPeerWithCA(t, h, ca, sm, stateID, "secret", "peapuser", "S3cr3t!")
	success, err := peer.run()
	require.NoError(t, err)
	assert.True(t, success, "PEAP inner EAP-MSCHAPv2 with a correct password must authenticate")

	state, err := sm.GetState(stateID)
	require.NoError(t, err)
	assert.True(t, state.Success)

	require.NotNil(t, peer.acceptResponse)
	_, rerr := microsoft.MSMPPERecvKey_Lookup(peer.acceptResponse)
	assert.NoError(t, rerr, "MS-MPPE-Recv-Key must be present on the Access-Accept")
	_, serr := microsoft.MSMPPESendKey_Lookup(peer.acceptResponse)
	assert.NoError(t, serr, "MS-MPPE-Send-Key must be present on the Access-Accept")
}

func TestPEAPHandler_FullInnerAuth_Fragmented(t *testing.T) {
	ca := newHSTestCA(t, "PEAP Inner Fragment Root CA")
	cfg := peapServerEngineConfig(t, ca)
	h := NewPEAPHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })
	h.maxFragment = 48 // force both handshake and inner records to fragment

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "peapuser", "secret")
	peer := newPeapPeerWithCA(t, h, ca, sm, stateID, "secret", "peapuser", "S3cr3t!")
	success, err := peer.run()
	require.NoError(t, err)
	assert.True(t, success, "fragmented PEAP inner exchange must still authenticate")
}

func TestPEAPHandler_InnerAuth_WrongPasswordRejected(t *testing.T) {
	ca := newHSTestCA(t, "PEAP Inner WrongPwd Root CA")
	cfg := peapServerEngineConfig(t, ca)
	h := NewPEAPHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "peapuser", "secret")
	// Server validates against "correct"; the client offers "wrong".
	peer := newPeapPeerWithCA(t, h, ca, sm, stateID, "secret", "peapuser", "wrong")
	peer.pwd = &mockPasswordProvider{password: "correct"}
	success, err := peer.run()
	assert.False(t, success, "wrong inner password must not authenticate")
	require.Error(t, err)
	assert.ErrorIs(t, err, eap.ErrPasswordMismatch)
}
