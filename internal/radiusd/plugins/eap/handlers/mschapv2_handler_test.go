package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// mockResponseWriter simulates a RADIUS response writer
type mockResponseWriter struct {
	response *radius.Packet
}

func (m *mockResponseWriter) Write(p *radius.Packet) error {
	m.response = p
	return nil
}

// mockPasswordProvider simulates a password provider
type mockPasswordProvider struct {
	password string
}

func (m *mockPasswordProvider) GetPassword(user *domain.RadiusUser, isMacAuth bool) (string, error) {
	if m.password != "" {
		return m.password, nil
	}
	return user.Password, nil
}

func TestMSCHAPv2Handler_Name(t *testing.T) {
	handler := NewMSCHAPv2Handler()
	assert.Equal(t, EAPMethodMSCHAPv2, handler.Name())
}

func TestMSCHAPv2Handler_EAPType(t *testing.T) {
	handler := NewMSCHAPv2Handler()
	assert.Equal(t, uint8(eap.TypeMSCHAPv2), handler.EAPType())
}

func TestMSCHAPv2Handler_CanHandle(t *testing.T) {
	handler := NewMSCHAPv2Handler()

	tests := []struct {
		name     string
		eapMsg   *eap.EAPMessage
		expected bool
	}{
		{
			name:     "nil message",
			eapMsg:   nil,
			expected: false,
		},
		{
			name: "correct type",
			eapMsg: &eap.EAPMessage{
				Type: eap.TypeMSCHAPv2,
			},
			expected: true,
		},
		{
			name: "wrong type",
			eapMsg: &eap.EAPMessage{
				Type: eap.TypeMD5Challenge,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &eap.EAPContext{
				EAPMessage: tt.eapMsg,
			}
			result := handler.CanHandle(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMSCHAPv2Handler_HandleIdentity(t *testing.T) {
	handler := NewMSCHAPv2Handler()
	stateManager := statemanager.NewMemoryStateManager()
	writer := &mockResponseWriter{}

	// Create RADIUS request
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	_ = rfc2865.UserName_SetString(packet, "testuser") //nolint:errcheck

	// Create EAP Identity Response
	identityMsg := &eap.EAPMessage{
		Code:       eap.CodeResponse,
		Identifier: 1,
		Type:       eap.TypeIdentity,
		Data:       []byte("testuser"),
	}

	req := &radius.Request{
		Packet: packet,
	}

	ctx := &eap.EAPContext{
		Context:        context.Background(),
		Request:        req,
		ResponseWriter: writer,
		EAPMessage:     identityMsg,
		StateManager:   stateManager,
		Secret:         "secret",
	}

	// Call HandleIdentity
	handled, err := handler.HandleIdentity(ctx)

	// Validate the result
	require.NoError(t, err)
	assert.True(t, handled)
	assert.NotNil(t, writer.response)
	assert.Equal(t, radius.CodeAccessChallenge, writer.response.Code)

	// Validate the EAP-Message attributes
	eapMsg, err := rfc2869.EAPMessage_Lookup(writer.response)
	require.NoError(t, err)
	assert.NotNil(t, eapMsg)
	assert.Equal(t, uint8(eap.CodeRequest), eapMsg[0])   // EAP Code
	assert.Equal(t, uint8(eap.TypeMSCHAPv2), eapMsg[4])  // EAP Type
	assert.Equal(t, uint8(MSCHAPv2Challenge), eapMsg[5]) // MS-CHAPv2 OpCode

	// Validate that status is stored
	stateID := rfc2865.State_GetString(writer.response)
	assert.NotEmpty(t, stateID)

	savedState, err := stateManager.GetState(stateID)
	require.NoError(t, err)
	assert.Equal(t, "testuser", savedState.Username)
	assert.Equal(t, EAPMethodMSCHAPv2, savedState.Method)
	assert.Len(t, savedState.Challenge, MSCHAPChallengeSize)
}

func TestMSCHAPv2Handler_buildChallengeRequest(t *testing.T) {
	handler := NewMSCHAPv2Handler()
	identifier := uint8(1)
	challenge := make([]byte, MSCHAPChallengeSize)
	for i := range challenge {
		challenge[i] = byte(i)
	}

	data := handler.buildChallengeRequest(identifier, challenge)

	// Validate EAP Header
	assert.Equal(t, uint8(eap.CodeRequest), data[0])
	assert.Equal(t, identifier, data[1])

	// Validate EAP Type
	assert.Equal(t, uint8(eap.TypeMSCHAPv2), data[4])

	// Validate MS-CHAPv2 OpCode
	assert.Equal(t, uint8(MSCHAPv2Challenge), data[5]) // Validate MS-CHAPv2-ID
	assert.Equal(t, identifier, data[6])

	// Validate Value-Size
	assert.Equal(t, MSCHAPChallengeSize, int(data[9]))

	// Validate Challenge
	assert.Equal(t, challenge, data[10:10+MSCHAPChallengeSize])

	// Validate Server Name
	assert.Equal(t, []byte(ServerName), data[10+MSCHAPChallengeSize:])
}

func TestMSCHAPv2Handler_parseResponse(t *testing.T) {
	handler := NewMSCHAPv2Handler()

	tests := []struct {
		name      string
		data      []byte
		expectErr bool
		validate  func(*testing.T, *MSCHAPv2ResponseData)
	}{
		{
			name:      "too short",
			data:      []byte{1, 2, 3},
			expectErr: true,
		},
		{
			name: "invalid opcode",
			data: []byte{
				99,    // Invalid OpCode
				1,     // MS-CHAPv2-ID
				0, 55, // MS-Length
				49, // Value-Size
			},
			expectErr: true,
		},
		{
			name: "invalid value size",
			data: []byte{
				MSCHAPv2Response, // OpCode
				1,                // MS-CHAPv2-ID
				0, 55,            // MS-Length
				10, // Invalid Value-Size
			},
			expectErr: true,
		},
		{
			name: "valid response",
			data: func() []byte {
				d := make([]byte, 5+MSCHAPResponseSize+4)
				d[0] = MSCHAPv2Response                 // OpCode
				d[1] = 1                                // MS-CHAPv2-ID
				d[2] = 0                                // MS-Length high
				d[3] = byte(5 + MSCHAPResponseSize + 4) // MS-Length low
				d[4] = MSCHAPResponseSize               // Value-Size

				// Peer-Challenge (16 bytes)
				for i := 0; i < 16; i++ {
					d[5+i] = byte(i)
				}
				// Reserved (8 bytes)
				// NT-Response (24 bytes)
				for i := 0; i < 24; i++ {
					d[5+16+8+i] = byte(i + 16)
				}
				// Flags
				d[5+16+8+24] = 0

				// Name
				copy(d[5+MSCHAPResponseSize:], []byte("user"))

				return d
			}(),
			expectErr: false,
			validate: func(t *testing.T, resp *MSCHAPv2ResponseData) {
				assert.Equal(t, uint8(MSCHAPv2Response), resp.OpCode)
				assert.Equal(t, uint8(1), resp.MsIdentifier)
				assert.Equal(t, uint8(MSCHAPResponseSize), resp.ValueSize)
				assert.Len(t, resp.PeerChallenge, 16)
				assert.Len(t, resp.NTResponse, 24)
				assert.Equal(t, []byte("user"), resp.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.parseResponse(tt.data)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}
		})
	}
}

func TestMSCHAPv2Handler_verifyResponse(t *testing.T) {
	handler := NewMSCHAPv2Handler()

	// This test needs a real MSCHAPv2 calculation
	// Use known test vectors
	username := "testuser"
	password := "password123"

	// To simplify the test, only verify that the function executes
	// Real password validation requires RFC 2759 test vectors

	authChallenge := make([]byte, 16)
	peerChallenge := make([]byte, 16)
	ntResponse := make([]byte, 24)

	packet := radius.New(radius.CodeAccessAccept, []byte("secret"))

	// This should fail because ntResponse is empty
	success, err := handler.verifyResponse(
		username,
		password,
		authChallenge,
		peerChallenge,
		ntResponse,
		packet,
		1,
	)

	require.NoError(t, err)
	assert.False(t, success) // Should fail validation
}

func TestMSCHAPv2Handler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	// Integration test: simulate the full EAP-MSCHAPv2 authentication flow
	handler := NewMSCHAPv2Handler()
	stateManager := statemanager.NewMemoryStateManager()
	pwdProvider := &mockPasswordProvider{password: "testpass"}

	// 1. Identity phase
	writer1 := &mockResponseWriter{}
	packet1 := radius.New(radius.CodeAccessRequest, []byte("secret"))
	_ = rfc2865.UserName_SetString(packet1, "testuser") //nolint:errcheck

	identityMsg := &eap.EAPMessage{
		Code:       eap.CodeResponse,
		Identifier: 1,
		Type:       eap.TypeIdentity,
		Data:       []byte("testuser"),
	}

	req1 := &radius.Request{Packet: packet1}
	ctx1 := &eap.EAPContext{
		Context:        context.Background(),
		Request:        req1,
		ResponseWriter: writer1,
		EAPMessage:     identityMsg,
		StateManager:   stateManager,
		PwdProvider:    pwdProvider,
		Secret:         "secret",
		User:           &domain.RadiusUser{Username: "testuser", Password: "testpass"},
	}

	handled, err := handler.HandleIdentity(ctx1)
	require.NoError(t, err)
	assert.True(t, handled)

	// Validate that a challenge was received
	assert.Equal(t, radius.CodeAccessChallenge, writer1.response.Code)
	stateID := rfc2865.State_GetString(writer1.response)
	assert.NotEmpty(t, stateID)

	// Note: it is not possible to complete the full Response phase test,
	// because the client must compute the correct NT-Response based on the challenge
	// This requires RFC 2759 test vectors or a real client simulation
}
