package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsengine"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

func TestNewTLSHandler(t *testing.T) {
	h := NewTLSHandler()
	assert.NotNil(t, h)
}

func TestTLSHandler_Name(t *testing.T) {
	h := NewTLSHandler()
	assert.Equal(t, "eap-tls", h.Name())
}

func TestTLSHandler_EAPType(t *testing.T) {
	h := NewTLSHandler()
	assert.Equal(t, uint8(eap.TypeTLS), h.EAPType())
}

func TestTLSHandler_CanHandle(t *testing.T) {
	h := NewTLSHandler()

	tests := []struct {
		name     string
		ctx      *eap.EAPContext
		expected bool
	}{
		{
			name:     "can handle TLS type",
			ctx:      &eap.EAPContext{EAPMessage: &eap.EAPMessage{Type: eap.TypeTLS}},
			expected: true,
		},
		{
			name:     "cannot handle other type",
			ctx:      &eap.EAPContext{EAPMessage: &eap.EAPMessage{Type: eap.TypeMD5Challenge}},
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

// TestTLSHandler_HandleIdentity verifies an EAP-TLS Start (S bit set, no data)
// is sent in an Access-Challenge with state stored (RFC 5216 §2.1.1/§3.1).
func TestTLSHandler_HandleIdentity(t *testing.T) {
	h := NewTLSHandler()
	stateManager := statemanager.NewMemoryStateManager()
	writer := &mockResponseWriter{}

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	require.NoError(t, rfc2865.UserName_SetString(packet, "testuser"))

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
	assert.NotEmpty(t, stateID)
	storedState, err := stateManager.GetState(stateID)
	require.NoError(t, err)
	assert.Equal(t, EAPMethodTLS, storedState.Method)
	assert.False(t, storedState.Success)

	// Verify the EAP-TLS Start payload.
	eapData, err := rfc2869.EAPMessage_Lookup(writer.response)
	require.NoError(t, err)
	require.Len(t, eapData, 6)
	assert.Equal(t, byte(eap.CodeRequest), eapData[0])
	assert.Equal(t, byte(7), eapData[1])
	assert.Equal(t, byte(eap.TypeTLS), eapData[4])
	assert.Equal(t, byte(TLSFlagStart), eapData[5], "Start (S) bit must be set")
	assert.Zero(t, eapData[5]&TLSFlagLengthIncluded, "L bit must be clear for a Start with no data")
}

func TestTLSHandler_buildStartRequest(t *testing.T) {
	h := NewTLSHandler()

	result := h.buildStartRequest(3)

	require.Len(t, result, 6)
	assert.Equal(t, byte(eap.CodeRequest), result[0], "Code should be Request")
	assert.Equal(t, byte(3), result[1], "Identifier should match")
	actualLen := (int(result[2]) << 8) | int(result[3])
	assert.Equal(t, 6, actualLen, "Length should be 6")
	assert.Equal(t, byte(eap.TypeTLS), result[4], "Type should be EAP-TLS")
	assert.Equal(t, byte(TLSFlagStart), result[5], "Flags should have only the Start bit set")
}

// TestTLSHandler_HandleResponse_NeverAuthenticatesWithoutConfig ensures that
// even after a complete TLS message is reassembled, a handler with no TLS
// material configured never grants access: it rejects with
// eap.ErrTLSNotConfigured (RFC 5216 §2.2 requires a verified peer certificate).
func TestTLSHandler_HandleResponse_NeverAuthenticatesWithoutConfig(t *testing.T) {
	h := NewTLSHandler()

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	require.NoError(t, rfc2865.State_SetString(packet, "state-tls-1"))

	sm := statemanager.NewMemoryStateManager()
	require.NoError(t, sm.SetState("state-tls-1", &eap.EAPState{StateID: "state-tls-1", Method: EAPMethodTLS}))

	// A single, complete (no M bit) EAP-TLS fragment carrying TLS data.
	ctx := &eap.EAPContext{
		Request:      &radius.Request{Packet: packet},
		StateManager: sm,
		EAPMessage:   &eap.EAPMessage{Type: eap.TypeTLS, Data: []byte{0x00, 0x16, 0x03, 0x01}},
	}

	success, err := h.HandleResponse(ctx)
	assert.False(t, success, "the EAP-TLS handler must never authenticate without configured trust anchors")
	assert.ErrorIs(t, err, eap.ErrTLSNotConfigured)
}

// TestTLSHandler_HandleResponse_MalformedFragment ensures a truncated framing
// (L flag set without the TLS Message Length field) is rejected without
// authenticating.
func TestTLSHandler_HandleResponse_MalformedFragment(t *testing.T) {
	h := NewTLSHandler()

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	require.NoError(t, rfc2865.State_SetString(packet, "state-tls-bad"))

	sm := statemanager.NewMemoryStateManager()
	require.NoError(t, sm.SetState("state-tls-bad", &eap.EAPState{StateID: "state-tls-bad", Method: EAPMethodTLS}))

	ctx := &eap.EAPContext{
		Request:      &radius.Request{Packet: packet},
		StateManager: sm,
		EAPMessage:   &eap.EAPMessage{Type: eap.TypeTLS, Data: []byte{TLSFlagLengthIncluded}},
	}

	success, err := h.HandleResponse(ctx)
	assert.False(t, success)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, eap.ErrTLSNotConfigured)
}

func TestTLSHandler_HandleResponse_PendingSuccessRequiresACK(t *testing.T) {
	h := NewTLSHandler()
	sm := statemanager.NewMemoryStateManager()
	const stateID = "state-tls-pending"
	require.NoError(t, sm.SetState(stateID, &eap.EAPState{
		StateID: stateID,
		Method:  EAPMethodTLS,
		Data: map[string]interface{}{
			stateKeyPendingSuccess: true,
		},
	}))

	writer := &mockResponseWriter{}
	ctx := newTLSResponseCtx(t, stateID, writer, sm, 10, []byte{TLSFlagStart})
	success, err := h.HandleResponse(ctx)
	assert.False(t, success)
	assert.ErrorIs(t, err, eap.ErrTLSUnexpectedFragment)
	assert.Nil(t, writer.response)
}

func TestTLSHandler_HandleResponse_EmptyInitialTLSRoundRejected(t *testing.T) {
	ca := newHSTestCA(t, "Test Root CA")
	cfg := serverEngineConfig(t, ca, ca)
	h := NewTLSHandlerWithConfig(func() (*tlsengine.Config, error) { return cfg, nil })

	sm := statemanager.NewMemoryStateManager()
	const stateID = "state-tls-empty"
	require.NoError(t, sm.SetState(stateID, &eap.EAPState{StateID: stateID, Method: EAPMethodTLS}))

	writer := &mockResponseWriter{}
	ctx := newTLSResponseCtx(t, stateID, writer, sm, 10, []byte{0x00})
	success, err := h.HandleResponse(ctx)
	assert.False(t, success)
	assert.ErrorIs(t, err, eap.ErrTLSUnexpectedFragment)
	assert.Nil(t, writer.response)
}

// TestTLSHandler_HandleResponse_FragmentReassembly drives a fragmented inbound
// TLS message: the first two fragments (M bit set) must each be acknowledged
// with an EAP-TLS fragment ACK in an Access-Challenge, the reassembly buffer
// must persist across rounds, and the final fragment must yield the complete
// message (here rejected safely until the TLS engine lands).
func TestTLSHandler_HandleResponse_FragmentReassembly(t *testing.T) {
	h := NewTLSHandler()
	sm := statemanager.NewMemoryStateManager()
	const stateID = "state-tls-frag"
	require.NoError(t, sm.SetState(stateID, &eap.EAPState{StateID: stateID, Method: EAPMethodTLS}))

	// Fragment 1: L + M, declared length 6, data "ab".
	frag1 := []byte{TLSFlagLengthIncluded | TLSFlagMoreFragments, 0x00, 0x00, 0x00, 0x06, 'a', 'b'}
	// Fragment 2: M, data "cd".
	frag2 := []byte{TLSFlagMoreFragments, 'c', 'd'}
	// Fragment 3: final, data "ef".
	frag3 := []byte{0x00, 'e', 'f'}

	// Fragment 1 -> expect a fragment ACK challenge.
	writer := &mockResponseWriter{}
	ctx := newTLSResponseCtx(t, stateID, writer, sm, 10, frag1)
	success, err := h.HandleResponse(ctx)
	require.NoError(t, err)
	assert.False(t, success)
	assertFragmentACK(t, writer, stateID, 11)

	// Fragment 2 -> another fragment ACK challenge.
	writer = &mockResponseWriter{}
	ctx = newTLSResponseCtx(t, stateID, writer, sm, 12, frag2)
	success, err = h.HandleResponse(ctx)
	require.NoError(t, err)
	assert.False(t, success)
	assertFragmentACK(t, writer, stateID, 13)

	// Verify the reassembly buffer persisted across rounds.
	state, err := sm.GetState(stateID)
	require.NoError(t, err)
	buf, ok := state.Data[stateKeyRxBuf].([]byte)
	require.True(t, ok)
	assert.Equal(t, []byte("abcd"), buf)

	// Fragment 3 (final) -> complete message; with no TLS material configured
	// the handler rejects safely with eap.ErrTLSNotConfigured.
	writer = &mockResponseWriter{}
	ctx = newTLSResponseCtx(t, stateID, writer, sm, 14, frag3)
	success, err = h.HandleResponse(ctx)
	assert.False(t, success)
	assert.ErrorIs(t, err, eap.ErrTLSNotConfigured)
	assert.Nil(t, writer.response, "no challenge is written once the message is complete")
}

// newTLSResponseCtx builds an EAPContext for an EAP-Response/EAP-TLS fragment.
func newTLSResponseCtx(t *testing.T, stateID string, writer *mockResponseWriter, sm eap.EAPStateManager, identifier uint8, fragData []byte) *eap.EAPContext {
	t.Helper()
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	require.NoError(t, rfc2865.State_SetString(packet, stateID))
	return &eap.EAPContext{
		Request:        &radius.Request{Packet: packet},
		ResponseWriter: writer,
		StateManager:   sm,
		Secret:         "secret",
		EAPMessage:     &eap.EAPMessage{Code: eap.CodeResponse, Identifier: identifier, Type: eap.TypeTLS, Data: fragData},
	}
}

// assertFragmentACK verifies that writer holds an Access-Challenge carrying an
// EAP-Request/EAP-TLS fragment ACK (single flags octet, all flags clear) with
// the expected (incremented) identifier and the echoed State attribute.
func assertFragmentACK(t *testing.T, writer *mockResponseWriter, stateID string, wantIdentifier uint8) {
	t.Helper()
	require.NotNil(t, writer.response)
	assert.Equal(t, radius.CodeAccessChallenge, writer.response.Code)
	assert.Equal(t, stateID, rfc2865.State_GetString(writer.response))

	eapData, err := rfc2869.EAPMessage_Lookup(writer.response)
	require.NoError(t, err)
	require.Len(t, eapData, 6, "EAP header (4) + Type (1) + Flags (1)")
	assert.Equal(t, byte(eap.CodeRequest), eapData[0])
	assert.Equal(t, wantIdentifier, eapData[1], "fragment ACK identifier must be incremented")
	assert.Equal(t, byte(eap.TypeTLS), eapData[4])
	assert.Zero(t, eapData[5], "fragment ACK flags octet must be all zero")
}

func TestTLSHandler_HandleResponse_NoState(t *testing.T) {
	h := NewTLSHandler()

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	ctx := &eap.EAPContext{
		Request:      &radius.Request{Packet: packet},
		StateManager: statemanager.NewMemoryStateManager(),
		EAPMessage:   &eap.EAPMessage{Type: eap.TypeTLS},
	}

	success, err := h.HandleResponse(ctx)
	assert.False(t, success)
	assert.ErrorIs(t, err, eap.ErrStateNotFound)
}

func TestTLSHandler_ImplementsInterface(t *testing.T) {
	var _ eap.EAPHandler = (*TLSHandler)(nil)
}
