package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager"
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

// TestTLSHandler_HandleResponse_NeverAuthenticates ensures the M1.1 skeleton can
// never grant access before the handshake is implemented.
func TestTLSHandler_HandleResponse_NeverAuthenticates(t *testing.T) {
	h := NewTLSHandler()

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	require.NoError(t, rfc2865.State_SetString(packet, "state-tls-1"))

	sm := statemanager.NewMemoryStateManager()
	require.NoError(t, sm.SetState("state-tls-1", &eap.EAPState{StateID: "state-tls-1", Method: EAPMethodTLS}))

	ctx := &eap.EAPContext{
		Request:      &radius.Request{Packet: packet},
		StateManager: sm,
		EAPMessage:   &eap.EAPMessage{Type: eap.TypeTLS, Data: []byte{TLSFlagLengthIncluded}},
	}

	success, err := h.HandleResponse(ctx)
	assert.False(t, success, "the EAP-TLS skeleton must never authenticate")
	assert.ErrorIs(t, err, eap.ErrTLSHandshakeNotImplemented)
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
