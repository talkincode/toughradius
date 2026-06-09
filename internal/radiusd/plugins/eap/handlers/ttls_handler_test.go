package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager"
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

// TestTTLSHandler_HandleResponse_NeverAuthenticates ensures the M9.1 skeleton
// rejects every challenge response with eap.ErrTTLSNotImplemented and never
// grants access before the outer TLS tunnel (M9.2) is implemented.
func TestTTLSHandler_HandleResponse_NeverAuthenticates(t *testing.T) {
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
	assert.False(t, success, "EAP-TTLS skeleton must never authenticate")
	require.Error(t, err)
	assert.ErrorIs(t, err, eap.ErrTTLSNotImplemented)
}
