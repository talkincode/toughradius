package handlers

import (
	"crypto/md5"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	MD5ChallengeLength = 16
	EAPMethodMD5       = "eap-md5"
)

// MD5Handler EAP-MD5 认证处理器
type MD5Handler struct{}

// NewMD5Handler 创建 EAP-MD5 处理器
func NewMD5Handler() *MD5Handler {
	return &MD5Handler{}
}

// Name 返回处理器名称
func (h *MD5Handler) Name() string {
	return EAPMethodMD5
}

// EAPType 返回 EAP 类型码
func (h *MD5Handler) EAPType() uint8 {
	return eap.TypeMD5Challenge
}

// CanHandle 判断是否可以处理该 EAP 消息
func (h *MD5Handler) CanHandle(ctx *eap.EAPContext) bool {
	if ctx.EAPMessage == nil {
		return false
	}
	return ctx.EAPMessage.Type == eap.TypeMD5Challenge
}

// HandleIdentity 处理 EAP-Response/Identity，发送 MD5 Challenge
func (h *MD5Handler) HandleIdentity(ctx *eap.EAPContext) (bool, error) {
	// 生成随机 challenge
	challenge, err := eap.GenerateRandomBytes(MD5ChallengeLength)
	if err != nil {
		return false, err
	}

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
		Method:    EAPMethodMD5,
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

// HandleResponse 处理 EAP-Response (Challenge Response)
func (h *MD5Handler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	// 获取状态
	stateID := rfc2865.State_GetString(ctx.Request.Packet)
	if stateID == "" {
		return false, eap.ErrStateNotFound
	}

	state, err := ctx.StateManager.GetState(stateID)
	if err != nil {
		return false, err
	}

	// 获取密码
	password, err := ctx.PwdProvider.GetPassword(ctx.User, ctx.IsMacAuth)
	if err != nil {
		return false, err
	}

	// 验证 MD5 响应
	if !h.verifyMD5Response(ctx.EAPMessage.Identifier, password, state.Challenge, ctx.EAPMessage.Data) {
		return false, eap.ErrPasswordMismatch
	}

	// 标记认证成功
	state.Success = true
	ctx.StateManager.SetState(stateID, state)

	return true, nil
}

// buildChallengeRequest 构建 MD5 Challenge Request
func (h *MD5Handler) buildChallengeRequest(identifier uint8, challenge []byte) []byte {
	// EAP-MD5 格式:
	// Code (1) | Identifier (1) | Length (2) | Type (1) | Value-Size (1) | Value (16) | Name (可选)

	valueSize := byte(len(challenge))
	dataLen := 1 + len(challenge) // Value-Size + Value
	totalLen := 5 + dataLen       // EAP header (4) + Type (1) + data

	buffer := make([]byte, totalLen)
	buffer[0] = eap.CodeRequest
	buffer[1] = identifier
	buffer[2] = byte(totalLen >> 8)
	buffer[3] = byte(totalLen)
	buffer[4] = eap.TypeMD5Challenge
	buffer[5] = valueSize
	copy(buffer[6:], challenge)

	return buffer
}

// verifyMD5Response 验证 MD5 响应
// MD5(identifier + password + challenge) == response
func (h *MD5Handler) verifyMD5Response(identifier uint8, password string, challenge, response []byte) bool {
	if len(response) < 1 {
		return false
	}

	// 响应格式: Value-Size (1) + Value (16)
	// 跳过 Value-Size
	actualResponse := response[1:]

	// 计算期望的 MD5
	hash := md5.New()
	hash.Write([]byte{identifier})
	hash.Write([]byte(password))
	hash.Write(challenge)
	expectedResponse := hash.Sum(nil)

	return eap.VerifyMD5Hash(expectedResponse, actualResponse)
}
