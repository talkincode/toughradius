package eap

import (
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// HandlerRegistry EAP 处理器注册表接口
type HandlerRegistry interface {
	GetHandler(eapType uint8) (EAPHandler, bool)
}

// Coordinator EAP 协调器，负责 EAP 消息的分发和处理
type Coordinator struct {
	stateManager    EAPStateManager
	pwdProvider     PasswordProvider
	handlerRegistry HandlerRegistry
}

// NewCoordinator 创建新的 EAP 协调器
func NewCoordinator(stateManager EAPStateManager, pwdProvider PasswordProvider, handlerRegistry HandlerRegistry) *Coordinator {
	return &Coordinator{
		stateManager:    stateManager,
		pwdProvider:     pwdProvider,
		handlerRegistry: handlerRegistry,
	}
}

// HandleEAPRequest 处理 EAP 请求
// 返回值: (handled bool, success bool, err error)
// handled: 是否已处理该请求
// success: 认证是否成功 (只在 handled=true 时有效)
// err: 处理过程中的错误
func (c *Coordinator) HandleEAPRequest(
	w radius.ResponseWriter,
	r *radius.Request,
	user *domain.RadiusUser,
	nas *domain.NetNas,
	response *radius.Packet,
	secret string,
	isMacAuth bool,
	configuredMethod string, // 配置的 EAP 方法 (eap-md5, eap-mschapv2, etc.)
) (handled bool, success bool, err error) {

	// 解析 EAP 消息
	eapMsg, err := ParseEAPMessage(r.Packet)
	if err != nil {
		// 不是 EAP 请求
		return false, false, nil
	}

	// 创建 EAP 上下文
	ctx := &EAPContext{
		Request:        r,
		ResponseWriter: w,
		Response:       response,
		User:           user,
		NAS:            nas,
		EAPMessage:     eapMsg,
		Secret:         secret,
		IsMacAuth:      isMacAuth,
		StateManager:   c.stateManager,
		PwdProvider:    c.pwdProvider,
	}

	// 处理 EAP-Response/Identity
	if eapMsg.Code == CodeResponse && eapMsg.Type == TypeIdentity {
		return c.handleIdentityResponse(ctx, configuredMethod)
	}

	// 处理 EAP-Response/Nak
	if eapMsg.Code == CodeResponse && eapMsg.Type == TypeNak {
		return c.handleNak(ctx)
	}

	// 处理 EAP-Response (Challenge Response)
	if eapMsg.Code == CodeResponse {
		return c.handleChallengeResponse(ctx)
	}

	return false, false, fmt.Errorf("unsupported EAP code: %d", eapMsg.Code)
}

// handleIdentityResponse 处理 EAP-Response/Identity
// 根据配置的方法发送相应的 Challenge
func (c *Coordinator) handleIdentityResponse(ctx *EAPContext, configuredMethod string) (bool, bool, error) {
	// 根据配置的方法选择处理器
	var handler EAPHandler

	switch configuredMethod {
	case "eap-md5":
		handler, _ = c.handlerRegistry.GetHandler(TypeMD5Challenge)
	case "eap-mschapv2":
		handler, _ = c.handlerRegistry.GetHandler(TypeMSCHAPv2)
	case "eap-otp":
		handler, _ = c.handlerRegistry.GetHandler(TypeOTP)
	default:
		// 默认使用 MD5
		handler, _ = c.handlerRegistry.GetHandler(TypeMD5Challenge)
	}

	if handler == nil {
		return false, false, fmt.Errorf("EAP handler not found for method: %s", configuredMethod)
	}

	// 调用处理器发送 Challenge
	handled, err := handler.HandleIdentity(ctx)
	if err != nil {
		zap.L().Error("EAP HandleIdentity failed",
			zap.String("method", configuredMethod),
			zap.Error(err))
	}

	// Identity 阶段不返回成功，需要等待 Challenge Response
	return handled, false, err
}

// handleNak 处理 EAP-Response/Nak
// 客户端拒绝当前方法，建议其他方法
func (c *Coordinator) handleNak(ctx *EAPContext) (bool, bool, error) {
	if len(ctx.EAPMessage.Data) == 0 {
		return false, false, fmt.Errorf("Nak message has no alternative methods")
	}

	// 获取客户端建议的第一个方法
	suggestedType := ctx.EAPMessage.Data[0]

	handler, ok := c.handlerRegistry.GetHandler(suggestedType)
	if !ok {
		return false, false, fmt.Errorf("unsupported EAP type: %d", suggestedType)
	}

	// 使用建议的方法发送 Challenge
	handled, err := handler.HandleIdentity(ctx)
	return handled, false, err
}

// handleChallengeResponse 处理 EAP-Response (Challenge Response)
func (c *Coordinator) handleChallengeResponse(ctx *EAPContext) (bool, bool, error) {
	// 根据 EAP Type 获取处理器
	handler, ok := c.handlerRegistry.GetHandler(ctx.EAPMessage.Type)
	if !ok {
		return false, false, fmt.Errorf("unsupported EAP type: %d", ctx.EAPMessage.Type)
	}

	// 检查处理器是否可以处理该消息
	if !handler.CanHandle(ctx) {
		return false, false, fmt.Errorf("handler cannot handle this EAP message")
	}

	// 调用处理器验证响应
	success, err := handler.HandleResponse(ctx)
	if err != nil {
		zap.L().Error("EAP HandleResponse failed",
			zap.String("method", handler.Name()),
			zap.Error(err))
		return true, false, err
	}

	return true, success, nil
}

// SendEAPSuccess 发送 EAP-Success 响应
func (c *Coordinator) SendEAPSuccess(w radius.ResponseWriter, r *radius.Request, response *radius.Packet, secret string) error {
	eapMsg, _ := ParseEAPMessage(r.Packet)
	identifier := eapMsg.Identifier

	// 创建 EAP-Success 消息
	eapSuccess := EncodeEAPHeader(CodeSuccess, identifier)

	// 设置 EAP-Message 和 Message-Authenticator
	SetEAPMessageAndAuth(response, eapSuccess, secret)

	// 发送响应
	if app.GConfig().Radiusd.Debug {
		zap.L().Info("Sending EAP-Success",
			zap.Uint8("identifier", identifier))
	}

	return w.Write(response)
}

// SendEAPFailure 发送 EAP-Failure 响应
func (c *Coordinator) SendEAPFailure(w radius.ResponseWriter, r *radius.Request, secret string, reason error) error {
	eapMsg, _ := ParseEAPMessage(r.Packet)
	identifier := eapMsg.Identifier

	// 创建 EAP-Failure 消息
	eapFailure := EncodeEAPHeader(CodeFailure, identifier)

	// 创建 RADIUS Reject 响应
	response := r.Response(radius.CodeAccessReject)

	// 设置 EAP-Message 和 Message-Authenticator
	SetEAPMessageAndAuth(response, eapFailure, secret)

	// 记录日志
	zap.L().Warn("Sending EAP-Failure",
		zap.Uint8("identifier", identifier),
		zap.Error(reason))

	return w.Write(response)
}

// CleanupState 清理 EAP 状态
func (c *Coordinator) CleanupState(r *radius.Request) {
	stateID := rfc2865.State_GetString(r.Packet)
	if stateID != "" {
		c.stateManager.DeleteState(stateID)
	}
}
