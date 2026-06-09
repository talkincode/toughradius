package handlers

import (
	"context"
	"errors"
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

func TestNewPEAPHandler(t *testing.T) {
	assert.NotNil(t, NewPEAPHandler())
}

func TestPEAPHandler_Name(t *testing.T) {
	assert.Equal(t, "eap-peap", NewPEAPHandler().Name())
}

func TestPEAPHandler_EAPType(t *testing.T) {
	assert.Equal(t, uint8(eap.TypePEAP), NewPEAPHandler().EAPType())
	assert.Equal(t, uint8(25), NewPEAPHandler().EAPType(), "PEAP is IANA EAP method type 25")
}

func TestPEAPHandler_CanHandle(t *testing.T) {
	h := NewPEAPHandler()

	tests := []struct {
		name     string
		ctx      *eap.EAPContext
		expected bool
	}{
		{
			name:     "can handle PEAP type",
			ctx:      &eap.EAPContext{EAPMessage: &eap.EAPMessage{Type: eap.TypePEAP}},
			expected: true,
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

// TestPEAPHandler_HandleIdentity verifies a PEAPv0 Start (S bit set, version 0,
// no data) is sent in an Access-Challenge with handshake state stored
// (Microsoft [MS-PEAP] §3.1.5.1, framing per RFC 5216 §3.1).
func TestPEAPHandler_HandleIdentity(t *testing.T) {
	h := NewPEAPHandler()
	stateManager := statemanager.NewMemoryStateManager()
	writer := &mockResponseWriter{}

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	require.NoError(t, rfc2865.UserName_SetString(packet, "peapuser"))

	ctx := &eap.EAPContext{
		Context:        context.Background(),
		Request:        &radius.Request{Packet: packet},
		ResponseWriter: writer,
		EAPMessage:     &eap.EAPMessage{Code: eap.CodeResponse, Identifier: 9, Type: eap.TypeIdentity},
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
	assert.Equal(t, EAPMethodPEAP, storedState.Method)
	assert.Equal(t, "peapuser", storedState.Username)
	assert.False(t, storedState.Success)

	// Verify the PEAPv0 Start payload.
	eapData, err := rfc2869.EAPMessage_Lookup(writer.response)
	require.NoError(t, err)
	require.Len(t, eapData, 6)
	assert.Equal(t, byte(eap.CodeRequest), eapData[0])
	assert.Equal(t, byte(9), eapData[1])
	assert.Equal(t, byte(eap.TypePEAP), eapData[4])
	assert.Equal(t, byte(tlsfragment.FlagStart), eapData[5], "Start (S) bit must be set, version 0")
	assert.Zero(t, eapData[5]&tlsfragment.FlagLengthIncluded, "L bit must be clear for a Start with no data")
	assert.Zero(t, eapData[5]&0x03, "PEAPv0 version bits must be 0")
}

func TestPEAPHandler_buildStartRequest(t *testing.T) {
	result := NewPEAPHandler().buildStartRequest(4)

	require.Len(t, result, 6)
	assert.Equal(t, byte(eap.CodeRequest), result[0], "Code should be Request")
	assert.Equal(t, byte(4), result[1], "Identifier should match")
	actualLen := (int(result[2]) << 8) | int(result[3])
	assert.Equal(t, 6, actualLen, "Length should be 6")
	assert.Equal(t, byte(eap.TypePEAP), result[4], "Type should be PEAP")
	assert.Equal(t, byte(tlsfragment.FlagStart), result[5], "Flags should have only the Start bit set (version 0)")
}

// TestPEAPHandler_HandleResponse_NeverAuthenticates ensures the M8.1 skeleton
// can never grant access: every handshake response is rejected with
// eap.ErrPEAPNotImplemented until the outer TLS tunnel (M8.2) exists.
func TestPEAPHandler_HandleResponse_NeverAuthenticates(t *testing.T) {
	h := NewPEAPHandler()

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	require.NoError(t, rfc2865.State_SetString(packet, "peap-state-1"))

	ctx := &eap.EAPContext{
		Context:      context.Background(),
		Request:      &radius.Request{Packet: packet},
		EAPMessage:   &eap.EAPMessage{Code: eap.CodeResponse, Identifier: 10, Type: eap.TypePEAP, Data: []byte{tlsfragment.FlagLengthIncluded}},
		StateManager: statemanager.NewMemoryStateManager(),
		Secret:       "secret",
	}

	success, err := h.HandleResponse(ctx)
	assert.False(t, success, "skeleton must never authenticate")
	require.Error(t, err)
	assert.True(t, errors.Is(err, eap.ErrPEAPNotImplemented))
}
