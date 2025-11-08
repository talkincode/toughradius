package handlers

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	OTPChallengeMessage = "Please enter a one-time password"
	EAPMethodOTP        = "eap-otp"
)

// OTPHandler EAP-OTP 认证处理器
type OTPHandler struct{}

// NewOTPHandler 创建 EAP-OTP 处理器
func NewOTPHandler() *OTPHandler {
	return &OTPHandler{}
}

// Name 返回处理器名称
func (h *OTPHandler) Name() string {
	return EAPMethodOTP
}

// EAPType 返回 EAP 类型码
func (h *OTPHandler) EAPType() uint8 {
	return eap.TypeOTP
}

// CanHandle 判断是否可以处理该 EAP 消息
func (h *OTPHandler) CanHandle(ctx *eap.EAPContext) bool {
	if ctx.EAPMessage == nil {
		return false
	}
	return ctx.EAPMessage.Type == eap.TypeOTP
}

// HandleIdentity 处理 EAP-Response/Identity，发送 OTP Challenge
func (h *OTPHandler) HandleIdentity(ctx *eap.EAPContext) (bool, error) {
	// OTP Challenge 是一个文本消息
	challenge := []byte(OTPChallengeMessage)

	// 创建 Challenge Request
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
		Method:    EAPMethodOTP,
		Success:   false,
	}

	if err := ctx.StateManager.SetState(stateID, state); err != nil {
		return false, err
	}

	// 设置 State 属性
	rfc2865.State_SetString(response, stateID)

	// 设置 EAP-Message 和 Message-Authenticator
	eap.SetEAPMessageAndAuth(response, eapData, ctx.Secret)

	// 发送响应
	return true, ctx.ResponseWriter.Write(response)
}

// HandleResponse 处理 EAP-Response (OTP Response)
func (h *OTPHandler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	// 获取状态
	stateID := rfc2865.State_GetString(ctx.Request.Packet)
	if stateID == "" {
		return false, eap.ErrStateNotFound
	}

	state, err := ctx.StateManager.GetState(stateID)
	if err != nil {
		return false, err
	}

	// 获取 OTP 密码 (这里简化处理，实际应该调用 OTP 验证服务)
	// TODO: 集成真实的 OTP 验证逻辑
	expectedOTP := "123456" // 示例固定值，实际应该从验证服务获取

	// 从 EAP Data 中获取用户输入的 OTP
	userOTP := string(ctx.EAPMessage.Data)

	// 验证 OTP
	if userOTP != expectedOTP {
		return false, eap.ErrPasswordMismatch
	}

	// 标记认证成功
	state.Success = true
	ctx.StateManager.SetState(stateID, state)

	return true, nil
}

// buildChallengeRequest 构建 OTP Challenge Request
func (h *OTPHandler) buildChallengeRequest(identifier uint8, challenge []byte) []byte {
	// EAP-OTP 格式:
	// Code (1) | Identifier (1) | Length (2) | Type (1) | Challenge (variable)

	dataLen := len(challenge)
	totalLen := 5 + dataLen // EAP header (4) + Type (1) + challenge

	buffer := make([]byte, totalLen)
	buffer[0] = eap.CodeRequest
	buffer[1] = identifier
	buffer[2] = byte(totalLen >> 8)
	buffer[3] = byte(totalLen)
	buffer[4] = eap.TypeOTP
	copy(buffer[5:], challenge)

	return buffer
}
