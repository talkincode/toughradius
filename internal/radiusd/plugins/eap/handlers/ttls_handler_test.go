package handlers

import (
	"context"
	"crypto/tls"
	"crypto/x509"
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

func TestNewTTLSHandler(t *testing.T) {
	assert.NotNil(t, NewTTLSHandler())
}

func TestTTLSHandler_Name(t *testing.T) {
	assert.Equal(t, "eap-ttls", NewTTLSHandler().Name())
}

func TestTTLSHandler_EAPType(t *testing.T) {
	assert.Equal(t, uint8(eap.TypeTTLS), NewTTLSHandler().EAPType())
	assert.Equal(t, uint8(21), NewTTLSHandler().EAPType(), "EAP-TTLS is IANA EAP method type 21")
}

func TestTTLSHandler_CanHandle(t *testing.T) {
	h := NewTTLSHandler()

	tests := []struct {
		name     string
		ctx      *eap.EAPContext
		expected bool
	}{
		{
			name:     "can handle TTLS type",
			ctx:      &eap.EAPContext{EAPMessage: &eap.EAPMessage{Type: eap.TypeTTLS}},
			expected: true,
		},
		{
			name:     "cannot handle PEAP type",
			ctx:      &eap.EAPContext{EAPMessage: &eap.EAPMessage{Type: eap.TypePEAP}},
			expected: false,
		},
		{
			name:     "cannot handle EAP-TLS type",
			ctx:      &eap.EAPContext{EAPMessage: &eap.EAPMessage{Type: eap.TypeTLS}},
			expected: false,
		},
		{
			name:     "cannot handle nil message",
			ctx:      &eap.EAPContext{EAPMessage: nil},
			expected: false,
		},
		{
			name:     "cannot handle nil context",
			ctx:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, h.CanHandle(tt.ctx))
		})
	}
}

// TestTTLSHandler_HandleIdentity verifies an EAP-TTLSv0 Start (S bit set,
// version 0, no data) is sent in an Access-Challenge with handshake state stored
// (RFC 5281 §9.2, framing per RFC 5216 §3.1).
func TestTTLSHandler_HandleIdentity(t *testing.T) {
	h := NewTTLSHandler()
	stateManager := statemanager.NewMemoryStateManager()
	writer := &mockResponseWriter{}

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	require.NoError(t, rfc2865.UserName_SetString(packet, "ttlsuser"))

	ctx := &eap.EAPContext{
		Context:        context.Background(),
		Request:        &radius.Request{Packet: packet},
		ResponseWriter: writer,
		EAPMessage:     &eap.EAPMessage{Code: eap.CodeResponse, Identifier: 7, Type: eap.TypeIdentity},
		StateManager:   stateManager,
		Secret:         "secret",
	}

	handled, err := h.HandleIdentity(ctx)
	require.NoError(t, err)
	assert.True(t, handled)
	require.NotNil(t, writer.response)
	assert.Equal(t, radius.CodeAccessChallenge, writer.response.Code)

	// A State attribute must be present so the client can echo it.
	stateID := rfc2865.State_GetString(writer.response)
	require.NotEmpty(t, stateID)
	storedState, err := stateManager.GetState(stateID)
	require.NoError(t, err)
	assert.Equal(t, EAPMethodTTLS, storedState.Method)
	assert.Equal(t, "ttlsuser", storedState.Username)
	assert.False(t, storedState.Success)

	// Verify the EAP-TTLSv0 Start payload.
	eapData, err := rfc2869.EAPMessage_Lookup(writer.response)
	require.NoError(t, err)
	require.Len(t, eapData, 6)
	assert.Equal(t, byte(eap.CodeRequest), eapData[0])
	assert.Equal(t, byte(7), eapData[1])
	assert.Equal(t, byte(eap.TypeTTLS), eapData[4])
	assert.Equal(t, byte(tlsfragment.FlagStart), eapData[5], "Start (S) bit must be set, version 0")
	assert.Zero(t, eapData[5]&tlsfragment.FlagLengthIncluded, "L bit must be clear for a Start with no data")
	assert.Zero(t, eapData[5]&0x07, "EAP-TTLSv0 version bits (RFC 5281 §9.1) must be 0")
}

func TestTTLSHandler_buildStartRequest(t *testing.T) {
	result := NewTTLSHandler().buildStartRequest(3)

	require.Len(t, result, 6)
	assert.Equal(t, byte(eap.CodeRequest), result[0], "Code should be Request")
	assert.Equal(t, byte(3), result[1], "Identifier should match")
	actualLen := (int(result[2]) << 8) | int(result[3])
	assert.Equal(t, 6, actualLen, "Length should be 6")
	assert.Equal(t, byte(eap.TypeTTLS), result[4], "Type should be TTLS")
	assert.Equal(t, byte(tlsfragment.FlagStart), result[5], "Flags should have only the Start bit set (version 0)")
}

// TestTTLSHandler_HandleResponse_NeverAuthenticatesWithoutConfig ensures the
// EAP-TTLS handler rejects safely before server certificate/key material is
// configured: without a TLS config provider the outer tunnel cannot be driven,
// so it can never grant access.
func TestTTLSHandler_HandleResponse_NeverAuthenticatesWithoutConfig(t *testing.T) {
	h := NewTTLSHandler()

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	require.NoError(t, rfc2865.State_SetString(packet, "ttls-state-1"))
	sm := statemanager.NewMemoryStateManager()
	require.NoError(t, sm.SetState("ttls-state-1", &eap.EAPState{StateID: "ttls-state-1", Method: EAPMethodTTLS}))

	ctx := &eap.EAPContext{
		Context:      context.Background(),
		Request:      &radius.Request{Packet: packet},
		EAPMessage:   &eap.EAPMessage{Code: eap.CodeResponse, Identifier: 8, Type: eap.TypeTTLS, Data: []byte{0x00, 0x16, 0x03, 0x01}},
		StateManager: sm,
		Secret:       "secret",
	}

	success, err := h.HandleResponse(ctx)
	assert.False(t, success, "EAP-TTLS must never authenticate without an outer TLS server certificate")
	require.Error(t, err)
	assert.ErrorIs(t, err, eap.ErrTLSNotConfigured)
}

func ttlsServerEngineConfig(t *testing.T, serverCA *hsTestCA) *tlsengine.Config {
	t.Helper()
	serverCert := serverCA.issue(t, "radius.example.com", func(c *x509.Certificate) {
		c.DNSNames = []string{"radius.example.com"}
	})
	return &tlsengine.Config{
		ServerCertificate: serverCert,
		ServerOnly:        true,
		MinVersion:        tls.VersionTLS12,
		HandshakeTimeout:  5 * time.Second,
	}
}

// ttlsTunnelClientCfg builds a server-only TLS client config (no client
// certificate): EAP-TTLS authenticates the peer with inner AVPs, so the outer
// handshake only verifies the server (RFC 5281 §7).
func ttlsTunnelClientCfg(serverCA *hsTestCA) *tls.Config {
	return &tls.Config{
		RootCAs:    serverCA.pool,
		ServerName: "radius.example.com",
		MinVersion: tls.VersionTLS12,
	}
}

// TestTTLSHandler_FullHandshake_EstablishesTunnelThenRejects drives a real
// server-only TLS handshake through the EAP-TTLS outer tunnel. Reaching the
// handshake-complete callback proves the tunnel and fragmentation framing
// (RFC 5281 §7-§9, RFC 5216 §2.1.5/§3.1) work end to end; the M9.2 milestone
// then rejects with eap.ErrTTLSInnerNotImplemented because the inner AVP
// authentication (M9.3+) is not yet implemented, so the tunnel never grants.
func TestTTLSHandler_FullHandshake_EstablishesTunnelThenRejects(t *testing.T) {
	ca := newHSTestCA(t, "TTLS Outer Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	sup := newSupplicantForType(t, h, eap.TypeTTLS, sm, stateID, "secret", ttlsTunnelClientCfg(ca))

	success, err := sup.run()
	assert.False(t, success, "EAP-TTLS M9.2 outer tunnel must not grant access on its own")
	require.Error(t, err)
	assert.ErrorIs(t, err, eap.ErrTTLSInnerNotImplemented)

	state, err := sm.GetState(stateID)
	require.NoError(t, err)
	assert.False(t, state.Success, "state must not be marked successful")
}

// TestTTLSHandler_FullHandshake_Fragmented forces a small fragment size so the
// outer handshake spans multiple EAP-TLS fragments, exercising reassembly and
// the ACK exchange (RFC 5216 §2.1.5). It must still reach the
// inner-not-implemented rejection without granting.
func TestTTLSHandler_FullHandshake_Fragmented(t *testing.T) {
	ca := newHSTestCA(t, "TTLS Outer Fragment Root CA")
	cfg := ttlsServerEngineConfig(t, ca)
	h := NewTTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })
	h.maxFragment = 48 // force the handshake records to fragment

	sm := statemanager.NewMemoryStateManager()
	defer sm.Close()

	stateID := startHandshake(t, h, sm, "ttlsuser", "secret")
	sup := newSupplicantForType(t, h, eap.TypeTTLS, sm, stateID, "secret", ttlsTunnelClientCfg(ca))

	success, err := sup.run()
	assert.False(t, success)
	require.Error(t, err)
	assert.ErrorIs(t, err, eap.ErrTTLSInnerNotImplemented)
}
