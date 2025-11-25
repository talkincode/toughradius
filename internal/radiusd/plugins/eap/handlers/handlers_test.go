package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
)

// Mock state manager for testing
type mockStateManagerForTest struct {
	states map[string]*eap.EAPState
}

func newMockStateManagerForTest() *mockStateManagerForTest {
	return &mockStateManagerForTest{
		states: make(map[string]*eap.EAPState),
	}
}

func (m *mockStateManagerForTest) GetState(stateID string) (*eap.EAPState, error) {
	state, ok := m.states[stateID]
	if !ok {
		return nil, eap.ErrStateNotFound
	}
	return state, nil
}

func (m *mockStateManagerForTest) SetState(stateID string, state *eap.EAPState) error {
	m.states[stateID] = state
	return nil
}

func (m *mockStateManagerForTest) DeleteState(stateID string) error {
	delete(m.states, stateID)
	return nil
}

// Tests for MD5Handler

func TestNewMD5Handler(t *testing.T) {
	h := NewMD5Handler()
	assert.NotNil(t, h)
}

func TestMD5Handler_Name(t *testing.T) {
	h := NewMD5Handler()
	assert.Equal(t, "eap-md5", h.Name())
}

func TestMD5Handler_EAPType(t *testing.T) {
	h := NewMD5Handler()
	assert.Equal(t, uint8(eap.TypeMD5Challenge), h.EAPType())
}

func TestMD5Handler_CanHandle(t *testing.T) {
	h := NewMD5Handler()

	tests := []struct {
		name     string
		ctx      *eap.EAPContext
		expected bool
	}{
		{
			name: "can handle MD5 type",
			ctx: &eap.EAPContext{
				EAPMessage: &eap.EAPMessage{
					Type: eap.TypeMD5Challenge,
				},
			},
			expected: true,
		},
		{
			name: "cannot handle other type",
			ctx: &eap.EAPContext{
				EAPMessage: &eap.EAPMessage{
					Type: eap.TypeOTP,
				},
			},
			expected: false,
		},
		{
			name: "cannot handle nil message",
			ctx: &eap.EAPContext{
				EAPMessage: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.CanHandle(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMD5Handler_buildChallengeRequest(t *testing.T) {
	h := NewMD5Handler()

	challenge := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	identifier := uint8(1)

	result := h.buildChallengeRequest(identifier, challenge)

	// Verify EAP header
	assert.Equal(t, byte(eap.CodeRequest), result[0], "Code should be Request")
	assert.Equal(t, identifier, result[1], "Identifier should match")
	assert.Equal(t, byte(eap.TypeMD5Challenge), result[4], "Type should be MD5-Challenge")

	// Verify length
	expectedLen := 5 + 1 + len(challenge) // header + type + value-size + challenge
	actualLen := (int(result[2]) << 8) | int(result[3])
	assert.Equal(t, expectedLen, actualLen, "Length should be correct")

	// Verify value-size
	assert.Equal(t, byte(len(challenge)), result[5], "Value-Size should be challenge length")

	// Verify challenge is included
	assert.Equal(t, challenge, result[6:6+len(challenge)], "Challenge should be included")
}

func TestMD5Handler_verifyMD5Response_EmptyResponse(t *testing.T) {
	h := NewMD5Handler()

	result := h.verifyMD5Response(1, "password", []byte("challenge"), []byte{})
	assert.False(t, result, "Empty response should fail verification")
}

// Tests for OTPHandler

func TestNewOTPHandler(t *testing.T) {
	h := NewOTPHandler()
	assert.NotNil(t, h)
}

func TestOTPHandler_Name(t *testing.T) {
	h := NewOTPHandler()
	assert.Equal(t, "eap-otp", h.Name())
}

func TestOTPHandler_EAPType(t *testing.T) {
	h := NewOTPHandler()
	assert.Equal(t, uint8(eap.TypeOTP), h.EAPType())
}

func TestOTPHandler_CanHandle(t *testing.T) {
	h := NewOTPHandler()

	tests := []struct {
		name     string
		ctx      *eap.EAPContext
		expected bool
	}{
		{
			name: "can handle OTP type",
			ctx: &eap.EAPContext{
				EAPMessage: &eap.EAPMessage{
					Type: eap.TypeOTP,
				},
			},
			expected: true,
		},
		{
			name: "cannot handle other type",
			ctx: &eap.EAPContext{
				EAPMessage: &eap.EAPMessage{
					Type: eap.TypeMD5Challenge,
				},
			},
			expected: false,
		},
		{
			name: "cannot handle nil message",
			ctx: &eap.EAPContext{
				EAPMessage: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.CanHandle(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOTPHandler_buildChallengeRequest(t *testing.T) {
	h := NewOTPHandler()

	challenge := []byte("Please enter a one-time password")
	identifier := uint8(2)

	result := h.buildChallengeRequest(identifier, challenge)

	// Verify EAP header
	assert.Equal(t, byte(eap.CodeRequest), result[0], "Code should be Request")
	assert.Equal(t, identifier, result[1], "Identifier should match")
	assert.Equal(t, byte(eap.TypeOTP), result[4], "Type should be OTP")

	// Verify length
	expectedLen := 5 + len(challenge) // header + type + challenge
	actualLen := (int(result[2]) << 8) | int(result[3])
	assert.Equal(t, expectedLen, actualLen, "Length should be correct")

	// Verify challenge is included
	assert.Equal(t, challenge, result[5:], "Challenge should be included")
}

// Tests for MSCHAPv2Handler - only tests not covered in mschapv2_handler_test.go

// Note: Most MSCHAPv2 tests are in mschapv2_handler_test.go
// This file adds additional edge case coverage

// Helper test to verify handler interface compliance

func TestHandlers_ImplementInterface(t *testing.T) {
	var _ eap.EAPHandler = (*MD5Handler)(nil)
	var _ eap.EAPHandler = (*OTPHandler)(nil)
	var _ eap.EAPHandler = (*MSCHAPv2Handler)(nil)
}

// Tests for constants

func TestConstants(t *testing.T) {
	assert.Equal(t, 16, MD5ChallengeLength, "MD5 challenge should be 16 bytes")
	assert.Equal(t, "eap-md5", EAPMethodMD5)
	assert.Equal(t, "Please enter a one-time password", OTPChallengeMessage)
	assert.Equal(t, "eap-otp", EAPMethodOTP)
}

// Test buildChallengeRequest with various sizes

func TestMD5Handler_buildChallengeRequest_VariousSizes(t *testing.T) {
	h := NewMD5Handler()

	tests := []struct {
		name      string
		challenge []byte
	}{
		{
			name:      "standard 16 byte challenge",
			challenge: make([]byte, 16),
		},
		{
			name:      "short challenge",
			challenge: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name:      "long challenge",
			challenge: make([]byte, 32),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.buildChallengeRequest(1, tt.challenge)
			require.NotNil(t, result)

			// Check length calculation
			actualLen := (int(result[2]) << 8) | int(result[3])
			expectedLen := 5 + 1 + len(tt.challenge) // header + type + value-size + challenge
			assert.Equal(t, expectedLen, actualLen)

			// Check value-size
			assert.Equal(t, byte(len(tt.challenge)), result[5])
		})
	}
}

func TestOTPHandler_buildChallengeRequest_VariousSizes(t *testing.T) {
	h := NewOTPHandler()

	tests := []struct {
		name      string
		challenge []byte
	}{
		{
			name:      "short message",
			challenge: []byte("OTP"),
		},
		{
			name:      "standard message",
			challenge: []byte("Please enter your OTP code"),
		},
		{
			name:      "long message",
			challenge: []byte("Please enter your one-time password from your authenticator app"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.buildChallengeRequest(1, tt.challenge)
			require.NotNil(t, result)

			// Check length calculation
			actualLen := (int(result[2]) << 8) | int(result[3])
			expectedLen := 5 + len(tt.challenge) // header + type + challenge
			assert.Equal(t, expectedLen, actualLen)

			// Verify challenge content
			assert.Equal(t, tt.challenge, result[5:])
		})
	}
}
