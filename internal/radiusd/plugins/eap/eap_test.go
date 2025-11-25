package eap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"layeh.com/radius"
	"layeh.com/radius/rfc2869"
)

// Tests for errors.go

func TestErrors(t *testing.T) {
	assert.NotNil(t, ErrInvalidEAPMessage)
	assert.NotNil(t, ErrStateNotFound)
	assert.NotNil(t, ErrPasswordMismatch)
	assert.NotNil(t, ErrUnsupportedEAPType)
	assert.NotNil(t, ErrAuthenticationFailed)

	assert.Equal(t, "invalid EAP message", ErrInvalidEAPMessage.Error())
	assert.Equal(t, "EAP state not found", ErrStateNotFound.Error())
	assert.Equal(t, "password mismatch", ErrPasswordMismatch.Error())
	assert.Equal(t, "unsupported EAP type", ErrUnsupportedEAPType.Error())
	assert.Equal(t, "authentication failed", ErrAuthenticationFailed.Error())
}

// Tests for message.go

func TestEAPMessage_String(t *testing.T) {
	tests := []struct {
		name     string
		msg      *EAPMessage
		expected string
	}{
		{
			name:     "nil message",
			msg:      nil,
			expected: "EAPMessage<nil>",
		},
		{
			name: "basic message",
			msg: &EAPMessage{
				Code:       CodeRequest,
				Identifier: 1,
				Length:     10,
				Type:       TypeMD5Challenge,
				Data:       []byte{0xab, 0xcd},
			},
			expected: "EAPMessage{Code=1, Identifier=1, Length=10, Type=4, Data=abcd}",
		},
		{
			name: "empty data",
			msg: &EAPMessage{
				Code:       CodeSuccess,
				Identifier: 5,
				Length:     4,
				Type:       0,
				Data:       nil,
			},
			expected: "EAPMessage{Code=3, Identifier=5, Length=4, Type=0, Data=}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for utils.go

func TestParseEAPMessage(t *testing.T) {
	tests := []struct {
		name      string
		setupPkt  func() *radius.Packet
		expectErr bool
		validate  func(*testing.T, *EAPMessage)
	}{
		{
			name: "valid EAP message",
			setupPkt: func() *radius.Packet {
				p := radius.New(radius.CodeAccessRequest, []byte("secret"))
				// EAP-Message: Code=1, ID=1, Length=10, Type=4, Data=5 bytes
				eapMsg := []byte{1, 1, 0, 10, 4, 0x01, 0x02, 0x03, 0x04, 0x05}
				rfc2869.EAPMessage_Set(p, eapMsg)
				return p
			},
			expectErr: false,
			validate: func(t *testing.T, msg *EAPMessage) {
				assert.Equal(t, uint8(1), msg.Code)
				assert.Equal(t, uint8(1), msg.Identifier)
				assert.Equal(t, uint16(10), msg.Length)
				assert.Equal(t, uint8(4), msg.Type)
				assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x05}, msg.Data)
			},
		},
		{
			name: "no EAP-Message attribute",
			setupPkt: func() *radius.Packet {
				return radius.New(radius.CodeAccessRequest, []byte("secret"))
			},
			expectErr: true,
		},
		{
			name: "short EAP message",
			setupPkt: func() *radius.Packet {
				p := radius.New(radius.CodeAccessRequest, []byte("secret"))
				// Only 3 bytes - too short
				rfc2869.EAPMessage_Set(p, []byte{1, 2, 3})
				return p
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := tt.setupPkt()
			msg, err := ParseEAPMessage(packet)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, msg)
				}
			}
		})
	}
}

func TestEAPMessage_Encode(t *testing.T) {
	tests := []struct {
		name     string
		msg      *EAPMessage
		validate func(*testing.T, []byte)
	}{
		{
			name: "encode with data",
			msg: &EAPMessage{
				Code:       CodeRequest,
				Identifier: 1,
				Type:       TypeMD5Challenge,
				Data:       []byte{0x01, 0x02, 0x03},
			},
			validate: func(t *testing.T, b []byte) {
				assert.Equal(t, uint8(CodeRequest), b[0])
				assert.Equal(t, uint8(1), b[1])
				// Length should be 8 (5 header + 3 data)
				assert.Equal(t, uint16(8), uint16(b[2])<<8|uint16(b[3]))
				assert.Equal(t, uint8(TypeMD5Challenge), b[4])
				assert.Equal(t, []byte{0x01, 0x02, 0x03}, b[5:])
			},
		},
		{
			name: "encode without data",
			msg: &EAPMessage{
				Code:       CodeSuccess,
				Identifier: 5,
				Type:       0,
				Data:       nil,
			},
			validate: func(t *testing.T, b []byte) {
				assert.Equal(t, uint8(CodeSuccess), b[0])
				assert.Equal(t, uint8(5), b[1])
				// Length should be 5 (header only with type)
				assert.Equal(t, uint16(5), uint16(b[2])<<8|uint16(b[3]))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.Encode()
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestEncodeEAPHeader(t *testing.T) {
	tests := []struct {
		name       string
		code       uint8
		identifier uint8
	}{
		{
			name:       "success",
			code:       CodeSuccess,
			identifier: 1,
		},
		{
			name:       "failure",
			code:       CodeFailure,
			identifier: 255,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeEAPHeader(tt.code, tt.identifier)

			assert.Len(t, result, 4)
			assert.Equal(t, tt.code, result[0])
			assert.Equal(t, tt.identifier, result[1])
			// Length should be 4
			length := uint16(result[2])<<8 | uint16(result[3])
			assert.Equal(t, uint16(4), length)
		})
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"16 bytes", 16},
		{"32 bytes", 32},
		{"1 byte", 1},
		{"0 bytes", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateRandomBytes(tt.length)

			require.NoError(t, err)
			assert.Len(t, result, tt.length)
		})
	}
}

func TestGenerateRandomBytes_Uniqueness(t *testing.T) {
	// Generate multiple random byte arrays and verify they're different
	b1, err1 := GenerateRandomBytes(16)
	b2, err2 := GenerateRandomBytes(16)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// They should be different (with extremely high probability)
	assert.NotEqual(t, b1, b2)
}

func TestGenerateMessageAuthenticator(t *testing.T) {
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	secret := "test-secret"

	result := GenerateMessageAuthenticator(packet, secret)

	assert.Len(t, result, 16) // MD5 hash is always 16 bytes
}

func TestSetEAPMessageAndAuth(t *testing.T) {
	response := radius.New(radius.CodeAccessChallenge, []byte("secret"))
	eapData := []byte{1, 1, 0, 5, 4} // Minimal EAP message
	secret := "test-secret"

	SetEAPMessageAndAuth(response, eapData, secret)

	// Verify EAP-Message was set
	eapMsg, err := rfc2869.EAPMessage_Lookup(response)
	require.NoError(t, err)
	assert.Equal(t, eapData, eapMsg)

	// Verify Message-Authenticator was set
	auth, err := rfc2869.MessageAuthenticator_Lookup(response)
	require.NoError(t, err)
	assert.Len(t, auth, 16)
}

func TestComputeMD5Hash(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty",
			data: []byte{},
		},
		{
			name: "hello",
			data: []byte("hello"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeMD5Hash(tt.data)
			assert.Len(t, result, 16)
		})
	}
}

func TestVerifyMD5Hash(t *testing.T) {
	tests := []struct {
		name     string
		expected []byte
		actual   []byte
		match    bool
	}{
		{
			name:     "matching",
			expected: []byte{1, 2, 3, 4},
			actual:   []byte{1, 2, 3, 4},
			match:    true,
		},
		{
			name:     "not matching",
			expected: []byte{1, 2, 3, 4},
			actual:   []byte{4, 3, 2, 1},
			match:    false,
		},
		{
			name:     "different lengths",
			expected: []byte{1, 2, 3},
			actual:   []byte{1, 2, 3, 4},
			match:    false,
		},
		{
			name:     "both empty",
			expected: []byte{},
			actual:   []byte{},
			match:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyMD5Hash(tt.expected, tt.actual)
			assert.Equal(t, tt.match, result)
		})
	}
}

// Tests for password_provider.go

func TestNewDefaultPasswordProvider(t *testing.T) {
	provider := NewDefaultPasswordProvider()
	assert.NotNil(t, provider)
}

func TestDefaultPasswordProvider_GetPassword(t *testing.T) {
	provider := NewDefaultPasswordProvider()

	tests := []struct {
		name      string
		user      *domain.RadiusUser
		isMacAuth bool
		expected  string
	}{
		{
			name: "regular auth",
			user: &domain.RadiusUser{
				Username: "testuser",
				Password: "testpass",
			},
			isMacAuth: false,
			expected:  "testpass",
		},
		{
			name: "mac auth with mac address",
			user: &domain.RadiusUser{
				Username: "aa:bb:cc:dd:ee:ff",
				Password: "testpass",
				MacAddr:  "aabbccddeeff",
			},
			isMacAuth: true,
			expected:  "aabbccddeeff",
		},
		{
			name: "mac auth without mac address",
			user: &domain.RadiusUser{
				Username: "aa:bb:cc:dd:ee:ff",
				Password: "testpass",
				MacAddr:  "",
			},
			isMacAuth: true,
			expected:  "aa:bb:cc:dd:ee:ff", // Falls back to username
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.GetPassword(tt.user, tt.isMacAuth)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for interfaces.go - only constants

func TestEAPConstants(t *testing.T) {
	// EAP code constants
	assert.Equal(t, uint8(1), uint8(CodeRequest))
	assert.Equal(t, uint8(2), uint8(CodeResponse))
	assert.Equal(t, uint8(3), uint8(CodeSuccess))
	assert.Equal(t, uint8(4), uint8(CodeFailure))

	// EAP type constants
	assert.Equal(t, uint8(1), uint8(TypeIdentity))
	assert.Equal(t, uint8(2), uint8(TypeNotification))
	assert.Equal(t, uint8(3), uint8(TypeNak))
	assert.Equal(t, uint8(4), uint8(TypeMD5Challenge))
	assert.Equal(t, uint8(5), uint8(TypeOTP))
	assert.Equal(t, uint8(6), uint8(TypeGTC))
	assert.Equal(t, uint8(13), uint8(TypeTLS))
	assert.Equal(t, uint8(26), uint8(TypeMSCHAPv2))
}

// Test PasswordProvider interface

func TestPasswordProviderInterface(t *testing.T) {
	var _ PasswordProvider = (*DefaultPasswordProvider)(nil)
}
