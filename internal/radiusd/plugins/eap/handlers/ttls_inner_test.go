package handlers

import (
	"context"
	"crypto/tls"
	"errors"
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
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// ttlsPapPeer emulates an EAP-TTLSv0 supplicant performing inner PAP. It drives
// the outer TLS handshake to completion (keeping the tls.Client alive) and then,
// because EAP-TTLS phase 2 is peer-initiated (RFC 5281 §7.3), sends its inner
// User-Name / User-Password AVP flight as TLS application data so the full TTLS
// inner state machine (handleInnerAVP) is exercised end to end.
//
// Like peapPeer it pins TLS 1.2 for deterministic EAP framing; the TLS 1.3
// application-data/key-export path is covered by tlsengine/appdata_test.go.
type ttlsPapPeer struct {
	t        *testing.T
	h        *TTLSHandler
	sm       eap.EAPStateManager
	stateID  string
	secret   string
	username string
	password string
	user     *domain.RadiusUser
	pwd      eap.PasswordProvider
	// verifier, when set, is exposed to the handler as an active external
	// credential authority (LDAP); inner PAP then binds through it instead of
	// comparing against pwd.
	verifier eap.CredentialVerifier

	// avpFlight overrides the inner AVP flight the peer sends; when nil it sends
	// a standard User-Name + User-Password PAP flight.
	avpFlight []byte

	client     *tls.Conn
	toClient   *hsStream
	fromClient *hsStream
	clientDone chan error
	hsReturned bool

	ident          uint8
	acceptResponse *radius.Packet
}

func newTTLSPapPeer(t *testing.T, h *TTLSHandler, ca *hsTestCA, sm eap.EAPStateManager, stateID, secret, username, password string) *ttlsPapPeer {
	t.Helper()
	toClient := newHSStream()
	fromClient := newHSStream()
	conn := &hsConn{rd: toClient, wr: fromClient}
	client := tls.Client(conn, ttlsTunnelClientCfg(ca))

	done := make(chan error, 1)
	go func() { done <- client.Handshake() }()

	return &ttlsPapPeer{
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

// run drives the whole EAP-TTLS PAP conversation and reports whether the handler
// granted access.
func (p *ttlsPapPeer) run() (bool, error) {
	respData := p.clientFlight() // ClientHello
	require.NotEmpty(p.t, respData, "client should produce a ClientHello")

	for round := 0; round < 64; round++ {
		if p.hsReturned {
			// Phase 2 is peer-initiated: send the inner PAP AVP flight as a
			// single application-data record.
			flight := p.avpFlight
			if flight == nil {
				flight = encodeTTLSPAP(p.username, p.password)
			}
			appData := p.encrypt(flight)
			granted, _, err := p.exchange(appData)
			return granted, err
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

// exchange sends one outer EAP-Response and reassembles the server's reply,
// acknowledging fragments until a full flight (or a grant) is obtained.
func (p *ttlsPapPeer) exchange(respData []byte) (granted bool, flight []byte, err error) {
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

func (p *ttlsPapPeer) responseCtx(writer *mockResponseWriter, tlsData []byte) *eap.EAPContext {
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
		Verifier:       p.verifier,
		EAPMessage:     &eap.EAPMessage{Code: eap.CodeResponse, Identifier: p.ident, Type: eap.TypeTTLS, Data: data},
	}
}

// clientFlight returns the next batch of TLS bytes the client produced, marking
// hsReturned once the client handshake completes.
func (p *ttlsPapPeer) clientFlight() []byte {
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

// encrypt encrypts an inner application record through the TLS client and
// returns the resulting TLS records.
func (p *ttlsPapPeer) encrypt(plaintext []byte) []byte {
	_, err := p.client.Write(plaintext)
	require.NoError(p.t, err)
	return p.fromClient.drain()
}

func ttlsParseChallenge(t *testing.T, resp *radius.Packet) *tlsfragment.Packet {
	t.Helper()
	eapData, err := rfc2869.EAPMessage_Lookup(resp)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(eapData), 5)
	require.Equal(t, uint8(eap.TypeTTLS), eapData[4])
	frag, err := tlsfragment.Parse(eapData[5:])
	require.NoError(t, err)
	return frag
}

// --- end-to-end inner PAP tests -------------------------------------------

func TestTTLSHandler_FullPAPAuth_Succeeds(t *testing.T) {
	ca := newHSTestCA(t, "TTLS PAP Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	peer := newTTLSPapPeer(t, h, ca, sm, stateID, "secret", "ttlsuser", "S3cr3t!")
	success, err := peer.run()
	require.NoError(t, err)
	assert.True(t, success, "EAP-TTLS inner PAP with a correct password must authenticate")

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

func TestTTLSHandler_FullPAPAuth_Fragmented(t *testing.T) {
	ca := newHSTestCA(t, "TTLS PAP Fragment Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })
	h.maxFragment = 48 // force the outer handshake records to fragment

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	peer := newTTLSPapPeer(t, h, ca, sm, stateID, "secret", "ttlsuser", "S3cr3t!")
	success, err := peer.run()
	require.NoError(t, err)
	assert.True(t, success, "fragmented EAP-TTLS outer handshake must still authenticate inner PAP")
}

func TestTTLSHandler_PAPAuth_WrongPasswordRejected(t *testing.T) {
	ca := newHSTestCA(t, "TTLS PAP WrongPwd Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	// The server validates against "correct"; the client offers "wrong".
	peer := newTTLSPapPeer(t, h, ca, sm, stateID, "secret", "ttlsuser", "wrong")
	peer.pwd = &mockPasswordProvider{password: "correct"}
	success, err := peer.run()
	assert.False(t, success, "wrong inner PAP password must not authenticate")
	require.Error(t, err)
	assert.ErrorIs(t, err, eap.ErrPasswordMismatch)

	state, err := sm.GetState(stateID)
	require.NoError(t, err)
	assert.False(t, state.Success)
}

func TestTTLSHandler_PAPAuth_NoUserPasswordRejected(t *testing.T) {
	ca := newHSTestCA(t, "TTLS PAP NoPwd Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	peer := newTTLSPapPeer(t, h, ca, sm, stateID, "secret", "ttlsuser", "S3cr3t!")
	// Send only a User-Name AVP: an inner method other than PAP (no
	// User-Password AVP), which M9.3 does not yet support.
	peer.avpFlight = buildTTLSAVP(ttlsAVPCodeUserName, true, 0, []byte("ttlsuser"))
	success, err := peer.run()
	assert.False(t, success, "a non-PAP inner method must not authenticate in M9.3")
	require.Error(t, err)
	assert.ErrorIs(t, err, eap.ErrTTLSInnerNotImplemented)
}

// fakeTTLSVerifier stands in for the LDAP credential backend (M14.2): when
// active, inner PAP must bind through it instead of comparing against the local
// password provider.
type fakeTTLSVerifier struct {
	active           bool
	err              error
	gotUser, gotPass string
	calls            int
}

func (f *fakeTTLSVerifier) Active() bool { return f.active }

func (f *fakeTTLSVerifier) VerifyCleartext(_ context.Context, username, password string) error {
	f.calls++
	f.gotUser, f.gotPass = username, password
	return f.err
}

func TestTTLSHandler_PAPAuth_LDAPBindSucceeds(t *testing.T) {
	ca := newHSTestCA(t, "TTLS PAP LDAP OK Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	peer := newTTLSPapPeer(t, h, ca, sm, stateID, "secret", "ttlsuser", "S3cr3t!")
	// The local password provider would REJECT (it holds a different password);
	// an active LDAP verifier must be consulted instead, and it accepts.
	peer.pwd = &mockPasswordProvider{password: "local-would-be-wrong"}
	v := &fakeTTLSVerifier{active: true}
	peer.verifier = v

	success, err := peer.run()
	require.NoError(t, err)
	assert.True(t, success, "inner PAP must authenticate via the active LDAP bind")

	// The verifier received the user's identity and the cleartext tunnel password,
	// proving the local password was bypassed.
	assert.Equal(t, 1, v.calls)
	assert.Equal(t, "ttlsuser", v.gotUser)
	assert.Equal(t, "S3cr3t!", v.gotPass)

	require.NotNil(t, peer.acceptResponse)
	_, rerr := microsoft.MSMPPERecvKey_Lookup(peer.acceptResponse)
	assert.NoError(t, rerr, "MS-MPPE keys must still be derived from the TLS session on LDAP success")
}

func TestTTLSHandler_PAPAuth_LDAPBindRejected(t *testing.T) {
	ca := newHSTestCA(t, "TTLS PAP LDAP Reject Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	peer := newTTLSPapPeer(t, h, ca, sm, stateID, "secret", "ttlsuser", "S3cr3t!")
	// The local password provider WOULD accept (it matches the presented
	// password); the active LDAP verifier rejects, and its verdict must win.
	// If the verifier branch were removed this test would falsely succeed.
	peer.pwd = &mockPasswordProvider{password: "S3cr3t!"}
	rejectErr := errors.New("ldap rejected")
	v := &fakeTTLSVerifier{active: true, err: rejectErr}
	peer.verifier = v

	success, err := peer.run()
	assert.False(t, success, "an active LDAP rejection must override a matching local password")
	require.Error(t, err)
	assert.ErrorIs(t, err, rejectErr)
	assert.Equal(t, 1, v.calls)

	state, err := sm.GetState(stateID)
	require.NoError(t, err)
	assert.False(t, state.Success)
}

// TestTTLSHandler_PAPAuth_InactiveVerifierUsesLocal proves an inactive verifier
// is ignored and the local password path is preserved unchanged.
func TestTTLSHandler_PAPAuth_InactiveVerifierUsesLocal(t *testing.T) {
	ca := newHSTestCA(t, "TTLS PAP LDAP Inactive Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	peer := newTTLSPapPeer(t, h, ca, sm, stateID, "secret", "ttlsuser", "S3cr3t!")
	peer.pwd = &mockPasswordProvider{password: "S3cr3t!"}
	v := &fakeTTLSVerifier{active: false, err: errors.New("must not be called")}
	peer.verifier = v

	success, err := peer.run()
	require.NoError(t, err)
	assert.True(t, success, "inactive verifier must fall back to the local password compare")
	assert.Equal(t, 0, v.calls, "an inactive verifier must never be consulted")
}
