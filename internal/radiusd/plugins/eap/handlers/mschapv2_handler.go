package handlers

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/microsoft"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2759"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc3079"
)

const (
	// MSCHAPv2 OpCodes
	MSCHAPv2Challenge = 1
	MSCHAPv2Response  = 2
	MSCHAPv2Success   = 3
	MSCHAPv2Failure   = 4

	// MSCHAPv2 常量
	MSCHAPChallengeSize = 16
	MSCHAPResponseSize  = 49 // PeerChallenge(16) + Reserved(8) + NTResponse(24) + Flags(1)
	EAPMethodMSCHAPv2   = "eap-mschapv2"
	ServerName          = "toughradius"
)

// MSCHAPv2Handler EAP-MSCHAPv2 认证处理器
type MSCHAPv2Handler struct{}

// NewMSCHAPv2Handler 创建 EAP-MSCHAPv2 处理器
func NewMSCHAPv2Handler() *MSCHAPv2Handler {
	return &MSCHAPv2Handler{}
}

// Name 返回处理器名称
func (h *MSCHAPv2Handler) Name() string {
	return EAPMethodMSCHAPv2
}

// EAPType 返回 EAP 类型码
func (h *MSCHAPv2Handler) EAPType() uint8 {
	return eap.TypeMSCHAPv2
}

// CanHandle 判断是否可以处理该 EAP 消息
func (h *MSCHAPv2Handler) CanHandle(ctx *eap.EAPContext) bool {
	if ctx.EAPMessage == nil {
		return false
	}
	return ctx.EAPMessage.Type == eap.TypeMSCHAPv2
}

// HandleIdentity 处理 EAP-Response/Identity，发送 MSCHAPv2 Challenge
func (h *MSCHAPv2Handler) HandleIdentity(ctx *eap.EAPContext) (bool, error) {
	// 生成随机 Authenticator Challenge (16 字节)
	challenge, err := eap.GenerateRandomBytes(MSCHAPChallengeSize)
	if err != nil {
		return false, fmt.Errorf("failed to generate challenge: %w", err)
	}

	// 构建 EAP-MSCHAPv2 Challenge Request
	eapData := h.buildChallengeRequest(ctx.EAPMessage.Identifier, challenge)

	// 创建 RADIUS Access-Challenge 响应
	response := ctx.Request.Response(radius.CodeAccessChallenge)

	// 生成并保存状态
	stateID := common.UUID()
	username := rfc2865.UserName_GetString(ctx.Request.Packet)

	state := &eap.EAPState{
		Username:  username,
		Challenge: challenge,
		StateID:   stateID,
		Method:    EAPMethodMSCHAPv2,
		Success:   false,
		Data:      make(map[string]interface{}),
	}

	// 保存 MSIdentifier 供后续使用
	state.Data["ms_identifier"] = ctx.EAPMessage.Identifier

	if err := ctx.StateManager.SetState(stateID, state); err != nil {
		return false, fmt.Errorf("failed to save state: %w", err)
	}

	// 设置 State 属性
	rfc2865.State_SetString(response, stateID)

	// 设置 EAP-Message 和 Message-Authenticator
	eap.SetEAPMessageAndAuth(response, eapData, ctx.Secret)

	// 发送响应
	return true, ctx.ResponseWriter.Write(response)
}

// HandleResponse 处理 EAP-Response (MSCHAPv2 Response)
func (h *MSCHAPv2Handler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	// 获取状态
	stateID := rfc2865.State_GetString(ctx.Request.Packet)
	if stateID == "" {
		return false, eap.ErrStateNotFound
	}

	state, err := ctx.StateManager.GetState(stateID)
	if err != nil {
		return false, err
	}

	// 解析 MSCHAPv2 Response
	msResp, err := h.parseResponse(ctx.EAPMessage.Data)
	if err != nil {
		return false, fmt.Errorf("failed to parse MSCHAPv2 response: %w", err)
	}

	// 获取密码
	password, err := ctx.PwdProvider.GetPassword(ctx.User, ctx.IsMacAuth)
	if err != nil {
		return false, err
	}

	// 验证 MSCHAPv2 Response
	success, err := h.verifyResponse(
		ctx.User.Username,
		password,
		state.Challenge,
		msResp.PeerChallenge,
		msResp.NTResponse,
		ctx.Response,
		msResp.MsIdentifier,
	)

	if err != nil {
		return false, err
	}

	if !success {
		return false, eap.ErrPasswordMismatch
	}

	// 标记认证成功
	state.Success = true
	ctx.StateManager.SetState(stateID, state)

	return true, nil
}

// buildChallengeRequest 构建 MSCHAPv2 Challenge Request
// EAP-MSCHAPv2 Challenge 格式:
// Code(1) | Identifier(1) | Length(2) | Type(1) | OpCode(1) | MS-CHAPv2-ID(1) |
// MS-Length(2) | Value-Size(1) | Challenge(16) | Name(variable)
func (h *MSCHAPv2Handler) buildChallengeRequest(identifier uint8, challenge []byte) []byte {
	serverName := []byte(ServerName)

	// 计算 MS-CHAPv2 数据长度
	// OpCode(1) + MS-CHAPv2-ID(1) + MS-Length(2) + Value-Size(1) + Challenge(16) + Name(len)
	msDataLen := 1 + 1 + 2 + 1 + MSCHAPChallengeSize + len(serverName)

	// 计算 EAP 总长度
	// EAP Header(4) + Type(1) + MS-CHAPv2 Data
	totalLen := 4 + 1 + msDataLen

	buffer := make([]byte, totalLen)

	// EAP Header
	buffer[0] = eap.CodeRequest
	buffer[1] = identifier
	binary.BigEndian.PutUint16(buffer[2:4], uint16(totalLen))

	// EAP Type
	buffer[4] = eap.TypeMSCHAPv2

	// MS-CHAPv2 Data
	offset := 5
	buffer[offset] = MSCHAPv2Challenge                                       // OpCode
	buffer[offset+1] = identifier                                            // MS-CHAPv2-ID (同 EAP Identifier)
	binary.BigEndian.PutUint16(buffer[offset+2:offset+4], uint16(msDataLen)) // MS-Length
	buffer[offset+4] = MSCHAPChallengeSize                                   // Value-Size
	copy(buffer[offset+5:offset+5+MSCHAPChallengeSize], challenge)           // Challenge
	copy(buffer[offset+5+MSCHAPChallengeSize:], serverName)                  // Name

	return buffer
}

// MSCHAPv2ResponseData MSCHAPv2 响应数据结构
type MSCHAPv2ResponseData struct {
	OpCode        uint8
	MsIdentifier  uint8
	MsLength      uint16
	ValueSize     uint8
	PeerChallenge []byte // 16 bytes
	Reserved      []byte // 8 bytes
	NTResponse    []byte // 24 bytes
	Flags         uint8
	Name          []byte
}

// parseResponse 解析 MSCHAPv2 Response
// EAP-MSCHAPv2 Response 格式:
// OpCode(1) | MS-CHAPv2-ID(1) | MS-Length(2) | Value-Size(1) |
// Peer-Challenge(16) | Reserved(8) | NT-Response(24) | Flags(1) | Name(variable)
func (h *MSCHAPv2Handler) parseResponse(data []byte) (*MSCHAPv2ResponseData, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("MSCHAPv2 response too short: %d bytes", len(data))
	}

	resp := &MSCHAPv2ResponseData{}
	offset := 0

	resp.OpCode = data[offset]
	resp.MsIdentifier = data[offset+1]
	resp.MsLength = binary.BigEndian.Uint16(data[offset+2 : offset+4])
	resp.ValueSize = data[offset+4]
	offset += 5

	// 检查 OpCode 和 ValueSize
	if resp.OpCode != MSCHAPv2Response {
		return nil, fmt.Errorf("invalid OpCode: expected %d, got %d", MSCHAPv2Response, resp.OpCode)
	}

	if resp.ValueSize != MSCHAPResponseSize {
		return nil, fmt.Errorf("invalid ValueSize: expected %d, got %d", MSCHAPResponseSize, resp.ValueSize)
	}

	// 检查剩余数据长度
	if len(data) < offset+int(resp.ValueSize) {
		return nil, fmt.Errorf("insufficient data for response value: need %d, have %d",
			offset+int(resp.ValueSize), len(data))
	}

	// 解析 Peer-Challenge, Reserved, NT-Response, Flags
	resp.PeerChallenge = data[offset : offset+16]
	resp.Reserved = data[offset+16 : offset+24]
	resp.NTResponse = data[offset+24 : offset+48]
	resp.Flags = data[offset+48]
	offset += int(resp.ValueSize)

	// Name (剩余部分)
	if len(data) > offset {
		resp.Name = data[offset:]
	}

	return resp, nil
}

// verifyResponse 验证 MSCHAPv2 Response 并生成加密密钥
func (h *MSCHAPv2Handler) verifyResponse(
	username string,
	password string,
	authChallenge []byte,
	peerChallenge []byte,
	ntResponse []byte,
	response *radius.Packet,
	msIdentifier uint8,
) (bool, error) {
	byteUser := []byte(username)
	bytePwd := []byte(password)

	// 使用 RFC 2759 生成 NT-Response
	expectedNTResponse, err := rfc2759.GenerateNTResponse(
		authChallenge,
		peerChallenge,
		byteUser,
		bytePwd,
	)
	if err != nil {
		return false, fmt.Errorf("failed to generate NT-Response: %w", err)
	}

	// 验证 NT-Response
	if !bytes.Equal(expectedNTResponse, ntResponse) {
		return false, nil
	}

	// 生成 MPPE 密钥
	recvKey, err := rfc3079.MakeKey(expectedNTResponse, bytePwd, false)
	if err != nil {
		return false, fmt.Errorf("failed to generate recv key: %w", err)
	}

	sendKey, err := rfc3079.MakeKey(expectedNTResponse, bytePwd, true)
	if err != nil {
		return false, fmt.Errorf("failed to generate send key: %w", err)
	}

	// 生成 Authenticator Response (RFC 2759)
	authenticatorResponse, err := rfc2759.GenerateAuthenticatorResponse(
		authChallenge,
		peerChallenge,
		expectedNTResponse,
		byteUser,
		bytePwd,
	)
	if err != nil {
		return false, fmt.Errorf("failed to generate authenticator response: %w", err)
	}

	// 构建 MSCHAP2-Success 属性值
	// 格式: Ident(1) + Authenticator-Response(42)
	success := make([]byte, 43)
	success[0] = msIdentifier
	copy(success[1:], authenticatorResponse)

	// 添加 Microsoft 特定属性到响应
	microsoft.MSCHAP2Success_Add(response, success)
	microsoft.MSMPPERecvKey_Add(response, recvKey)
	microsoft.MSMPPESendKey_Add(response, sendKey)
	microsoft.MSMPPEEncryptionPolicy_Add(response,
		microsoft.MSMPPEEncryptionPolicy_Value_EncryptionAllowed)
	microsoft.MSMPPEEncryptionTypes_Add(response,
		microsoft.MSMPPEEncryptionTypes_Value_RC440or128BitAllowed)

	return true, nil
}
