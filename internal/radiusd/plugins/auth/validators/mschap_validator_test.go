package validators

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/microsoft"
	"layeh.com/radius"
)

func TestMSCHAPValidator_Name(t *testing.T) {
	validator := &MSCHAPValidator{}
	assert.Equal(t, "mschap", validator.Name())
}

func TestMSCHAPValidator_CanHandle(t *testing.T) {
	validator := &MSCHAPValidator{}

	tests := []struct {
		name        string
		setupPacket func(*radius.Packet)
		expected    bool
	}{
		{
			name: "with both challenge and response",
			setupPacket: func(packet *radius.Packet) {
				microsoft.MSCHAPChallenge_Add(packet, make([]byte, 16))
				microsoft.MSCHAP2Response_Add(packet, make([]byte, 50))
			},
			expected: true,
		},
		{
			name: "with challenge only",
			setupPacket: func(packet *radius.Packet) {
				microsoft.MSCHAPChallenge_Add(packet, make([]byte, 16))
			},
			expected: false,
		},
		{
			name: "with response only",
			setupPacket: func(packet *radius.Packet) {
				microsoft.MSCHAP2Response_Add(packet, make([]byte, 50))
			},
			expected: false,
		},
		{
			name: "without attributes",
			setupPacket: func(packet *radius.Packet) {
				// Do not add any attributes
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
			tt.setupPacket(packet)

			req := &radius.Request{Packet: packet}
			authCtx := &auth.AuthContext{Request: req}

			result := validator.CanHandle(authCtx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMSCHAPValidator_Validate_InvalidLength(t *testing.T) {
	validator := &MSCHAPValidator{}
	ctx := context.Background()

	tests := []struct {
		name          string
		challengeLen  int
		responseLen   int
		expectError   bool
		errorContains string
	}{
		{
			name:          "invalid challenge length",
			challengeLen:  8,
			responseLen:   50,
			expectError:   true,
			errorContains: "challenge len or response len error",
		},
		{
			name:          "invalid response length",
			challengeLen:  16,
			responseLen:   30,
			expectError:   true,
			errorContains: "challenge len or response len error",
		},
		{
			name:          "both invalid",
			challengeLen:  8,
			responseLen:   30,
			expectError:   true,
			errorContains: "challenge len or response len error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))

			microsoft.MSCHAPChallenge_Add(packet, make([]byte, tt.challengeLen))
			microsoft.MSCHAP2Response_Add(packet, make([]byte, tt.responseLen))

			req := &radius.Request{Packet: packet}
			user := &domain.RadiusUser{Username: "testuser", Password: "testpass"}
			authCtx := &auth.AuthContext{
				Request:  req,
				Response: response,
				User:     user,
			}

			err := validator.Validate(ctx, authCtx, "testpass")

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMSCHAPValidator_Validate_PasswordMismatch(t *testing.T) {
	// This test uses simple data for a mismatched password validation
	// Full MSCHAPv2 validation should use RFC 2759 test vectors
	validator := &MSCHAPValidator{}
	ctx := context.Background()

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	response := radius.New(radius.CodeAccessAccept, []byte("secret"))

	// Create a valid-length but random challenge and response
	challenge := make([]byte, 16)
	mschapResponse := make([]byte, 50)

	// Set up the basic structure
	mschapResponse[0] = 1 // ident
	// Keep the remaining bytes zero

	microsoft.MSCHAPChallenge_Add(packet, challenge)
	microsoft.MSCHAP2Response_Add(packet, mschapResponse)

	req := &radius.Request{Packet: packet}
	user := &domain.RadiusUser{Username: "testuser", Password: "testpass"}
	authCtx := &auth.AuthContext{
		Request:  req,
		Response: response,
		User:     user,
	}

	// With random challenge and response, validation should fail
	err := validator.Validate(ctx, authCtx, "testpass")
	require.Error(t, err)
}
