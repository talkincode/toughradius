package validators

import (
	"context"
	"crypto/md5"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestCHAPValidator_Name(t *testing.T) {
	validator := &CHAPValidator{}
	assert.Equal(t, "chap", validator.Name())
}

func TestCHAPValidator_CanHandle(t *testing.T) {
	validator := &CHAPValidator{}

	tests := []struct {
		name         string
		chapPassword []byte
		expected     bool
	}{
		{
			name:         "with chap password",
			chapPassword: make([]byte, 17),
			expected:     true,
		},
		{
			name:         "without chap password",
			chapPassword: nil,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
			if tt.chapPassword != nil {
				_ = rfc2865.CHAPPassword_Add(packet, tt.chapPassword) //nolint:errcheck
			}

			req := &radius.Request{Packet: packet}
			authCtx := &auth.AuthContext{Request: req}

			result := validator.CanHandle(authCtx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCHAPValidator_Validate(t *testing.T) {
	validator := &CHAPValidator{}
	ctx := context.Background()

	tests := []struct {
		name        string
		password    string
		setupPacket func(*radius.Packet, string)
		expectError bool
		errorMsg    string
	}{
		{
			name:     "valid chap authentication",
			password: "testpass123",
			setupPacket: func(packet *radius.Packet, password string) {
				// Construct a valid CHAP password
				chapID := byte(1)
				challenge := make([]byte, 16)
				for i := range challenge {
					challenge[i] = byte(i)
				}

				// Compute the CHAP response
				w := md5.New()
				w.Write([]byte{chapID})
				w.Write([]byte(password))
				w.Write(challenge)
				response := w.Sum(nil)

				// Construct the CHAP password (ID + response)
				chapPassword := make([]byte, 17)
				chapPassword[0] = chapID
				copy(chapPassword[1:], response)

				_ = rfc2865.CHAPPassword_Add(packet, chapPassword) //nolint:errcheck
				_ = rfc2865.CHAPChallenge_Add(packet, challenge)   //nolint:errcheck
			},
			expectError: false,
		},
		{
			name:     "wrong password",
			password: "testpass123",
			setupPacket: func(packet *radius.Packet, password string) {
				chapID := byte(1)
				challenge := make([]byte, 16)
				for i := range challenge {
					challenge[i] = byte(i)
				}

				// Use an incorrect password for computation
				w := md5.New()
				w.Write([]byte{chapID})
				w.Write([]byte("wrongpassword"))
				w.Write(challenge)
				response := w.Sum(nil)

				chapPassword := make([]byte, 17)
				chapPassword[0] = chapID
				copy(chapPassword[1:], response)

				_ = rfc2865.CHAPPassword_Add(packet, chapPassword) //nolint:errcheck
				_ = rfc2865.CHAPChallenge_Add(packet, challenge)   //nolint:errcheck
			},
			expectError: true,
		},
		{
			name:     "invalid chap password length",
			password: "testpass123",
			setupPacket: func(packet *radius.Packet, password string) {
				// CHAP password with an incorrect length
				chapPassword := make([]byte, 10)
				challenge := make([]byte, 16)

				_ = rfc2865.CHAPPassword_Add(packet, chapPassword) //nolint:errcheck
				_ = rfc2865.CHAPChallenge_Add(packet, challenge)   //nolint:errcheck
			},
			expectError: true,
			errorMsg:    "must be 17 bytes",
		},
		{
			name:     "invalid chap challenge length",
			password: "testpass123",
			setupPacket: func(packet *radius.Packet, password string) {
				chapPassword := make([]byte, 17)
				// Challenge with incorrect length
				challenge := make([]byte, 8)

				_ = rfc2865.CHAPPassword_Add(packet, chapPassword) //nolint:errcheck
				_ = rfc2865.CHAPChallenge_Add(packet, challenge)   //nolint:errcheck
			},
			expectError: true,
			errorMsg:    "must be 16 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
			tt.setupPacket(packet, tt.password)

			req := &radius.Request{Packet: packet}
			user := &domain.RadiusUser{Username: "testuser", Password: tt.password}
			authCtx := &auth.AuthContext{
				Request: req,
				User:    user,
			}

			err := validator.Validate(ctx, authCtx, tt.password)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
