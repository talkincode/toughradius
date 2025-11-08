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

// mockResponseWriter 模拟 RADIUS 响应写入器
type mockResponseWriter struct {
	response *radius.Packet
}

func (m *mockResponseWriter) Write(p *radius.Packet) error {
	m.response = p
	return nil
}

// mockPasswordProvider 模拟密码提供者
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

	// 创建 RADIUS 请求
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2865.UserName_SetString(packet, "testuser")

	// 创建 EAP Identity Response
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

	// 调用 HandleIdentity
	handled, err := handler.HandleIdentity(ctx)

	// 验证结果
	require.NoError(t, err)
	assert.True(t, handled)
	assert.NotNil(t, writer.response)
	assert.Equal(t, radius.CodeAccessChallenge, writer.response.Code)

	// 验证 EAP-Message 属性
	eapMsg, err := rfc2869.EAPMessage_Lookup(writer.response)
	require.NoError(t, err)
	assert.NotNil(t, eapMsg)
	assert.Equal(t, uint8(eap.CodeRequest), eapMsg[0])   // EAP Code
	assert.Equal(t, uint8(eap.TypeMSCHAPv2), eapMsg[4])  // EAP Type
	assert.Equal(t, uint8(MSCHAPv2Challenge), eapMsg[5]) // MS-CHAPv2 OpCode

	// 验证状态已保存
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

	// 验证 EAP Header
	assert.Equal(t, uint8(eap.CodeRequest), data[0])
	assert.Equal(t, identifier, data[1])

	// 验证 EAP Type
	assert.Equal(t, uint8(eap.TypeMSCHAPv2), data[4])

	// 验证 MS-CHAPv2 OpCode
	assert.Equal(t, uint8(MSCHAPv2Challenge), data[5]) // 验证 MS-CHAPv2-ID
	assert.Equal(t, identifier, data[6])

	// 验证 Value-Size
	assert.Equal(t, MSCHAPChallengeSize, int(data[9]))

	// 验证 Challenge
	assert.Equal(t, challenge, data[10:10+MSCHAPChallengeSize])

	// 验证 Server Name
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

	// 这个测试需要真实的 MSCHAPv2 计算
	// 使用已知的测试向量
	username := "testuser"
	password := "password123"

	// 为了简化测试,这里只测试函数能否正常执行
	// 真实的密码验证需要使用 RFC 2759 的测试向量

	authChallenge := make([]byte, 16)
	peerChallenge := make([]byte, 16)
	ntResponse := make([]byte, 24)

	packet := radius.New(radius.CodeAccessAccept, []byte("secret"))

	// 这应该失败,因为 ntResponse 是空的
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
	assert.False(t, success) // 应该验证失败
}

func TestMSCHAPv2Handler_Integration(t *testing.T) {
	// 集成测试:模拟完整的 EAP-MSCHAPv2 认证流程
	handler := NewMSCHAPv2Handler()
	stateManager := statemanager.NewMemoryStateManager()
	pwdProvider := &mockPasswordProvider{password: "testpass"}

	// 1. Identity 阶段
	writer1 := &mockResponseWriter{}
	packet1 := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2865.UserName_SetString(packet1, "testuser")

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

	// 验证收到 Challenge
	assert.Equal(t, radius.CodeAccessChallenge, writer1.response.Code)
	stateID := rfc2865.State_GetString(writer1.response)
	assert.NotEmpty(t, stateID)

	// 注意:这里无法完成完整的 Response 阶段测试,
	// 因为需要客户端根据 Challenge 计算正确的 NT-Response
	// 这需要使用 RFC 2759 的测试向量或真实的客户端模拟
}
